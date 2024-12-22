package pbft

import (
    "crypto/md5"
    "github.com/salemmohammed/PaxiBFT"
    "github.com/salemmohammed/PaxiBFT/log"
    "time"
    "bytes"
    "sync"
)

// Status represents different states of a request
type status int8

const (
    NONE status = iota
    PREPREPARED
    PREPARED
    COMMITTED
    RECEIVED
)

// BackgroundTransfer handles asynchronous data transmission
type BackgroundTransfer struct {
    data     chan *PaxiBFT.Request
    pending  map[string]*PaxiBFT.Request
    mu       sync.RWMutex
}

// Entry represents a log entry for request processing
type entry struct {
    ballot    PaxiBFT.Ballot
    view      PaxiBFT.View
    command   PaxiBFT.Command
    commit    bool
    active    bool
    Leader    bool
    request   *PaxiBFT.Request
    timestamp time.Time
    Digest    []byte
    Q1        *PaxiBFT.Quorum
    Q2        *PaxiBFT.Quorum
    Q3        *PaxiBFT.Quorum
    Q4        *PaxiBFT.Quorum
    Pstatus   status
    Cstatus   status
    Rstatus   status
    dataChan  chan struct{}
}

// Pbft represents the main PBFT protocol instance
type Pbft struct {
    PaxiBFT.Node
    config          []PaxiBFT.ID
    N               PaxiBFT.Config
    log             map[int]*entry
    slot            int
    view            PaxiBFT.View
    ballot          PaxiBFT.Ballot
    execute         int
    requests        []*PaxiBFT.Request
    quorum          *PaxiBFT.Quorum
    backgroundTx    *BackgroundTransfer
    ReplyWhenCommit bool
    RecivedReq      bool
    Member          *PaxiBFT.Memberlist
    dataLock        sync.RWMutex
}

// NewPbft creates and initializes a new PBFT instance
func NewPbft(n PaxiBFT.Node, options ...func(*Pbft)) *Pbft {
    p := &Pbft{
        Node:            n,
        log:             make(map[int]*entry, PaxiBFT.GetConfig().BufferSize),
        quorum:          PaxiBFT.NewQuorum(),
        slot:            -1,
        requests:        make([]*PaxiBFT.Request, 0),
        ReplyWhenCommit: false,
        RecivedReq:      false,
        Member:          PaxiBFT.NewMember(),
        backgroundTx:    &BackgroundTransfer{
            data:    make(chan *PaxiBFT.Request, 1000),
            pending: make(map[string]*PaxiBFT.Request),
        },
    }

    for _, opt := range options {
        opt(p)
    }
    return p
}

// GetMD5Hash generates MD5 hash of request data
func GetMD5Hash(r *PaxiBFT.Request) []byte {
    hasher := md5.New()
    hasher.Write([]byte(r.Command.Value))
    return hasher.Sum(nil)
}

// HandleRequest processes new requests with background data transmission
func (p *Pbft) HandleRequest(r PaxiBFT.Request, s int) {
    log.Debugf("<--------------------HandleRequest------------------>")
    e := p.log[s]
    digest := GetMD5Hash(&r)
    e.Digest = digest

    // Start background data transmission
    go p.sendDataInBackground(&r, s)

    // Send PrePrepare with hash only
    p.PrePrepare(&r, s)
    log.Debugf("<--------------------EndHandleRequest------------------>")
}

// sendDataInBackground handles asynchronous data transmission
func (p *Pbft) sendDataInBackground(r *PaxiBFT.Request, slot int) {
    dataMsg := DataMessage{
        ID:      p.ID(),
        Slot:    slot,
        Request: r,
        Digest:  GetMD5Hash(r),
    }

    // Send to all replicas except self
    for _, id := range p.config {
        if id != p.ID() {
            go func(target PaxiBFT.ID) {
                // Try sending with timeout
                done := make(chan bool)
                go func() {
                    p.Send(target, dataMsg)
                    done <- true
                }()

                select {
                case <-done:
                    log.Debugf("Data sent to %v for slot %d", target, slot)
                case <-time.After(time.Second):
                    log.Warningf("Timeout sending data to %v for slot %d", target, slot)
                }
            }(id)
        }
    }
}

// PrePrepare initiates the consensus process
func (p *Pbft) PrePrepare(r *PaxiBFT.Request, slot int) {
    log.Debugf("<--------------------PrePrepare------------------>")
    p.Broadcast(PrePrepare{
        Ballot:     p.ballot,
        ID:         p.ID(),
        View:       p.view,
        Slot:       slot,
        Digest:     GetMD5Hash(r),
    })
    log.Debugf("++++++ PrePrepare Done ++++++")
}

// HandlePre processes PrePrepare messages
func (p *Pbft) HandlePre(m PrePrepare) {
    log.Debugf("<--------------------HandlePre------------------>")
    log.Debugf(" Sender  %v ", m.ID)
    log.Debugf(" m.Slot  %v ", m.Slot)

    if m.Ballot > p.ballot {
        p.ballot = m.Ballot
        p.view = m.View
    }

    e, ok := p.log[m.Slot]
    if !ok {
        p.log[m.Slot] = &entry{
            ballot:    p.ballot,
            view:      p.view,
            commit:    false,
            active:    false,
            Leader:    false,
            timestamp: time.Now(),
            Digest:    m.Digest,
            Q1:        PaxiBFT.NewQuorum(),
            Q2:        PaxiBFT.NewQuorum(),
            Q3:        PaxiBFT.NewQuorum(),
            Q4:        PaxiBFT.NewQuorum(),
            dataChan:  make(chan struct{}),
        }
        e = p.log[m.Slot]
    }

    if e != nil && (e.Digest == nil || bytes.Equal(e.Digest, m.Digest)) {
        log.Debugf("Broadcasting Prepare message")
        p.Broadcast(Prepare{
            Ballot:  p.ballot,
            ID:      p.ID(),
            View:    m.View,
            Slot:    m.Slot,
            Digest:  m.Digest,
        })
    }
    log.Debugf("++++++ HandlePre Done ++++++")
}

// HandlePrepare processes Prepare messages
func (p *Pbft) HandlePrepare(m Prepare) {
    log.Debugf("<--------------------HandlePrepare------------------>")
    log.Debugf(" Sender  %v ", m.ID)
    log.Debugf("p.slot=%v", p.slot)
    log.Debugf("m.slot=%v", m.Slot)

    e, ok := p.log[m.Slot]
    if !ok {
        p.log[m.Slot] = &entry{
            ballot:    p.ballot,
            view:      p.view,
            commit:    false,
            active:    false,
            Leader:    false,
            timestamp: time.Now(),
            Digest:    m.Digest,
            Q1:        PaxiBFT.NewQuorum(),
            Q2:        PaxiBFT.NewQuorum(),
            Q3:        PaxiBFT.NewQuorum(),
            Q4:        PaxiBFT.NewQuorum(),
            dataChan:  make(chan struct{}),
        }
        e = p.log[m.Slot]
    }

    if e != nil {
        e.Q1.ACK(m.ID)
        if e.Q1.Majority() {
            e.Q1.Reset()
            e.Pstatus = PREPARED
            p.Broadcast(Commit{
                Ballot:  p.ballot,
                ID:      p.ID(),
                View:    p.view,
                Slot:    m.Slot,
                Digest:  m.Digest,
            })
        }
        
        // Check if we can execute
        if e.Cstatus == COMMITTED && e.Pstatus == PREPARED && e.Rstatus == RECEIVED {
            e.commit = true
            p.exec()
        }
    }
    log.Debugf("++++++ HandlePrepare Done ++++++")
}

// HandleCommit processes Commit messages
func (p *Pbft) HandleCommit(m Commit) {
    log.Debugf("<--------------------HandleCommit------------------>")
    if p.execute > m.Slot {
        return
    }

    e, ok := p.log[m.Slot]
    if !ok || e == nil {
        return
    }

    e.Q2.ACK(m.ID)
    if e.Q2.Majority() {
        e.Cstatus = COMMITTED
        if e.Pstatus == PREPARED && e.Rstatus == RECEIVED {
            e.commit = true
            p.exec()
        }
    }
    log.Debugf("++++++ HandleCommit Done ++++++")
}

// HandleDataMessage processes data messages received in background
func (p *Pbft) HandleDataMessage(m DataMessage) {
    p.dataLock.Lock()
    defer p.dataLock.Unlock()

    // Verify digest
    if digest := GetMD5Hash(m.Request); !bytes.Equal(digest, m.Digest) {
        log.Warningf("Data message digest mismatch for slot %d", m.Slot)
        return
    }

    // Update log entry with received data
    if e, exists := p.log[m.Slot]; exists {
        e.request = m.Request
        e.command = m.Request.Command
        e.Rstatus = RECEIVED

        // Signal data reception
        if e.dataChan != nil {
            close(e.dataChan)
        }

        // Check if we can execute
        if e.Cstatus == COMMITTED && e.Pstatus == PREPARED {
            e.commit = true
            p.exec()
        }
    }
}

// exec executes committed requests in order
func (p *Pbft) exec() {
    for {
        e, ok := p.log[p.execute]
        if !ok || !e.commit {
            break
        }

        if e.request != nil {
            value := p.Execute(e.command)
            reply := PaxiBFT.Reply{
                Command:    e.command,
                Value:     value,
                Properties: make(map[string]string),
            }

            if e.request != nil {
                if e.Leader {
                    e.request.Reply(reply)
                } else {
                    e.request.Reply(reply)
                }
                e.request = nil
            }
        }

        delete(p.log, p.execute)
        p.execute++
    }
}