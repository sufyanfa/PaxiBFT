package pbft

import (
    "crypto/md5"
    "github.com/salemmohammed/PaxiBFT"
    "github.com/salemmohammed/PaxiBFT/log"
    "time"
    "bytes"
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

// Entry represents a log entry in the PBFT protocol
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
    ReplyWhenCommit bool
    RecivedReq      bool
    Member          *PaxiBFT.Memberlist
}

// NewPbft creates a new PBFT instance
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
    }
    for _, opt := range options {
        opt(p)
    }
    return p
}

// GetMD5Hash calculates hash of request data
func GetMD5Hash(r *PaxiBFT.Request) []byte {
    hasher := md5.New()
    hasher.Write([]byte(r.Command.Value))
    return hasher.Sum(nil)
}

// HandleRequest processes incoming requests
func (p *Pbft) HandleRequest(r PaxiBFT.Request, s int) {
    log.Debugf("<--------------------HandleRequest------------------>")
    e := p.log[s]
    e.Digest = GetMD5Hash(&r)
    p.PrePrepare(&r, s)
    log.Debugf("<--------------------EndHandleRequest------------------>")
}

// PrePrepare initiates the PrePrepare phase
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

// HandlePre processes PrePrepare messages and handles hash verification
func (p *Pbft) HandlePre(m PrePrepare) {
    log.Debugf("<--------------------HandlePre------------------>")
    log.Debugf(" Sender  %v ", m.ID)
    log.Debugf(" m.Slot  %v ", m.Slot)

    // Update ballot and view if needed
    if m.Ballot > p.ballot {
        p.ballot = m.Ballot
        p.view = m.View
    }

    // Create or get log entry
    e, ok := p.log[m.Slot]
    if !ok {
        // Create new entry if doesn't exist
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
        }
        e = p.log[m.Slot]
    }

    // Verify digest and broadcast Prepare
    if e != nil && (e.Digest == nil || bytes.Equal(e.Digest, m.Digest)) {
        log.Debugf("m.Ballot=%v, p.ballot=%v, m.view=%v", m.Ballot, p.ballot, m.View)
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

// HandlePrepare processes Prepare messages with improved error handling
func (p *Pbft) HandlePrepare(m Prepare) {
    log.Debugf("<--------------------HandlePrepare------------------>")
    log.Debugf(" Sender  %v ", m.ID)
    log.Debugf("p.slot=%v", p.slot)
    log.Debugf("m.slot=%v", m.Slot)

    // Create or get log entry
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
    if !ok {
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
}

// exec executes committed commands
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
            
            if e.request != nil && e.Leader {
                e.request.Reply(reply)
            } else {
                e.request.Reply(reply)
            }
            e.request = nil
        }
        
        delete(p.log, p.execute)
        p.execute++
    }
}