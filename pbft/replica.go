package pbft

import (
    "fmt"
    "github.com/salemmohammed/PaxiBFT"
    "github.com/salemmohammed/PaxiBFT/log"
    "strconv"
    "time"
    "unsafe"
)

// Replica represents a node in the PBFT network
type Replica struct {
    PaxiBFT.Node    // Embedded Node type for basic node functionality
    *Pbft           // Pointer to PBFT protocol instance
}

// NewReplica creates and initializes a new replica instance
func NewReplica(id PaxiBFT.ID) *Replica {
    r := new(Replica)
    r.Node = PaxiBFT.NewNode(id)
    r.Pbft = NewPbft(r)
    
    // Register message handlers
    r.Register(PaxiBFT.Request{}, r.handleRequest)
    r.Register(PrePrepare{}, r.HandlePre)
    r.Register(Prepare{}, r.HandlePrepare)
    r.Register(Commit{}, r.HandleCommit)
    r.Register(DataMessage{}, r.HandleDataMessage)
    
    return r
}

// handleRequest processes incoming client requests
func (p *Replica) handleRequest(m PaxiBFT.Request) {
    log.Debugf("<---------------------handleRequest------------------------->")
    p.slot++

    // Print initial PBFT message for first request
    if p.slot == 0 {
        fmt.Println("-------------------PBFT-------------------------")
    }
    
    // Print progress message every 1000 requests
    if p.slot%1000 == 0 {
        fmt.Printf("Processing batch %d, slot %d\n", p.slot/1000, p.slot)
    }

    // Create or get the log entry for this slot
    e, ok := p.log[p.slot]
    if !ok {
        // Initialize new log entry
        p.log[p.slot] = &entry{
            ballot:    p.ballot,
            view:      p.view,
            command:   m.Command,
            commit:    false,
            active:    false,
            Leader:    false,
            request:   &m,
            timestamp: time.Now(),
            Digest:    GetMD5Hash(&m),
            Q1:        PaxiBFT.NewQuorum(),
            Q2:        PaxiBFT.NewQuorum(),
            Q3:        PaxiBFT.NewQuorum(),
            Q4:        PaxiBFT.NewQuorum(),
            dataChan:  make(chan struct{}),
        }
        e = p.log[p.slot]
    }

    // Update entry with request details
    e.command = m.Command
    e.request = &m

    // Calculate request size for monitoring
    requestSize := unsafe.Sizeof(m)
    if m.Command.Value != nil {
        requestSize += uintptr(len(m.Command.Value))
    }
    for k, v := range m.Properties {
        requestSize += uintptr(len(k)) + uintptr(len(v))
    }
    log.Debugf("Request size: %d bytes", requestSize)

    // Check if this replica is the leader
    Node_ID := PaxiBFT.ID(strconv.Itoa(1) + "." + strconv.Itoa(1))
    log.Debugf("Node_ID = %v", Node_ID)
    
    if Node_ID == p.ID() {
        log.Debugf("Acting as view leader: %v", p.ID())
        e.active = true
        e.Leader = true
        p.ballot.Next(p.ID())
        p.view.Next(p.ID())
        p.requests = append(p.requests, &m)
        
        // Leader initiates consensus with background data transmission
        p.Pbft.HandleRequest(m, p.slot)
    } else {
        // Non-leader nodes wait for data if needed
        if e.request == nil {
            select {
            case <-e.dataChan:
                log.Debugf("Received data for slot %d", p.slot)
            case <-time.After(5 * time.Second):
                log.Warningf("Timeout waiting for data in slot %d", p.slot)
                return
            }
        }
    }

    // Handle execution if already committed
    if e.commit {
        log.Debugf("Request already committed, executing")
        p.exec()
        return
    }

    // Update request status and check for execution conditions
    e.Rstatus = RECEIVED
    if e.Cstatus == COMMITTED && e.Pstatus == PREPARED {
        e.commit = true
        p.exec()
    }
    
    log.Debugf("Request processing complete for slot %d", p.slot)
}