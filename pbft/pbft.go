package pbft

import (
    "crypto/md5"
    "github.com/salemmohammed/PaxiBFT"
    "github.com/salemmohammed/PaxiBFT/log"
    "time"
    "bytes"
)

// Status represents different states of a request in the PBFT protocol
type status int8

const (
    NONE status = iota        // Initial state
    PREPREPARED              // Request has been pre-prepared
    PREPARED                 // Request has been prepared
    COMMITTED               // Request has been committed
    RECEIVED                // Request has been received
)

// Entry represents a log entry for tracking request state
type entry struct {
    ballot    PaxiBFT.Ballot       // Current ballot number
    view      PaxiBFT.View         // Current view number
    command   PaxiBFT.Command      // Client command to execute
    commit    bool                 // Whether entry is committed
    active    bool                 // Whether view is active
    Leader    bool                 // Whether this node is leader
    request   *PaxiBFT.Request     // Original client request
    timestamp time.Time            // Time request was received
    Digest    []byte               // Hash of request data
    Q1        *PaxiBFT.Quorum     // Quorum for PrePrepare phase
    Q2        *PaxiBFT.Quorum     // Quorum for Prepare phase
    Q3        *PaxiBFT.Quorum     // Quorum for Commit phase
    Q4        *PaxiBFT.Quorum     // Additional quorum if needed
    Pstatus   status              // Prepare status
    Cstatus   status              // Commit status 
    Rstatus   status              // Received status
}

// Pbft represents the main PBFT protocol instance
type Pbft struct {
    PaxiBFT.Node                   // Embedded Node type
    config          []PaxiBFT.ID   // Node configuration
    N               PaxiBFT.Config // Network configuration
    log             map[int]*entry // Log of requests/operations
    slot            int            // Current sequence number
    view            PaxiBFT.View   // Current view number
    ballot          PaxiBFT.Ballot // Current ballot number
    execute         int            // Next execution sequence
    requests        []*PaxiBFT.Request // Pending requests
    quorum          *PaxiBFT.Quorum   // Quorum tracker
    ReplyWhenCommit bool              // Reply policy flag
    RecivedReq      bool              // Request received flag
    Member          *PaxiBFT.Memberlist // Member list
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
    }
    // Apply any custom options
    for _, opt := range options {
        opt(p)
    }
    return p
}

// GetMD5Hash generates MD5 hash of request data for integrity checking
func GetMD5Hash(r *PaxiBFT.Request) []byte {
    hasher := md5.New()
    hasher.Write([]byte(r.Command.Value))
    return hasher.Sum(nil)
}

// HandleRequest initiates request processing by creating PrePrepare message
func (p *Pbft) HandleRequest(r PaxiBFT.Request, s int) {
    log.Debugf("<--------------------HandleRequest------------------>")
    e := p.log[s]
    e.Digest = GetMD5Hash(&r)
    p.PrePrepare(&r, s)
    log.Debugf("<--------------------EndHandleRequest------------------>")
}

// PrePrepare starts the consensus process by broadcasting PrePrepare message
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

// HandlePre processes PrePrepare messages and initiates Prepare phase
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

// HandlePrepare processes Prepare messages and moves to Commit phase if quorum reached
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

// HandleCommit processes Commit messages and executes command if commit quorum reached
func (p *Pbft) HandleCommit(m Commit) {
    log.Debugf("<--------------------HandleCommit------------------>")
    // Skip if already executed
    if p.execute > m.Slot {
        return
    }

    // Get or validate log entry
    e, ok := p.log[m.Slot]
    if !ok {
        return
    }

    // Track commit acknowledgments
    e.Q2.ACK(m.ID)
    if e.Q2.Majority() {
        e.Cstatus = COMMITTED
        // Execute if all conditions met
        if e.Pstatus == PREPARED && e.Rstatus == RECEIVED {
            e.commit = true
            p.exec()
        }
    }
}

// exec executes all committed commands in order
func (p *Pbft) exec() {
    for {
        // Get next entry to execute
        e, ok := p.log[p.execute]
        if !ok || !e.commit {
            break
        }
        
        // Execute command and send reply
        if e.request != nil {
            value := p.Execute(e.command)
            reply := PaxiBFT.Reply{
                Command:    e.command,
                Value:     value,
                Properties: make(map[string]string),
            }
            
            // Send reply to client
            if e.request != nil && e.Leader {
                e.request.Reply(reply)
            } else {
                e.request.Reply(reply)
            }
            e.request = nil
        }
        
        // Cleanup and advance execution counter
        delete(p.log, p.execute)
        p.execute++
    }
}