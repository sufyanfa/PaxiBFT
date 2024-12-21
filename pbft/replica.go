// Package pbft implements a replica node in PBFT consensus protocol
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
// It extends the basic Node type and includes PBFT-specific functionality
type Replica struct {
    PaxiBFT.Node    // Embedded Node type for basic node functionality
    *Pbft           // Pointer to PBFT protocol instance
}

// NewReplica creates and initializes a new replica instance
// id: unique identifier for the replica
func NewReplica(id PaxiBFT.ID) *Replica {
    r := new(Replica)
    r.Node = PaxiBFT.NewNode(id)
    r.Pbft = NewPbft(r)
    
    // Register message handlers for different types of PBFT messages
    r.Register(PaxiBFT.Request{}, r.handleRequest)
    r.Register(PrePrepare{}, r.HandlePre)
    r.Register(Prepare{}, r.HandlePrepare)
    r.Register(Commit{}, r.HandleCommit)
    
    return r
}

// handleRequest processes incoming client requests
// This is the main entry point for new requests in the system
func (p *Replica) handleRequest(m PaxiBFT.Request) {
    log.Debugf("<---------------------handleRequest------------------------->")
    p.slot++ // Increment sequence number

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
        // Initialize new log entry with request details
        p.log[p.slot] = &entry{
            ballot:    p.ballot,        // Current ballot number
            view:      p.view,          // Current view number
            command:   m.Command,        // Client command
            commit:    false,           // Commit flag
            active:    false,           // Active view flag
            Leader:    false,           // Leader flag
            request:   &m,              // Original request
            timestamp: time.Now(),      // Request timestamp
            Digest:    GetMD5Hash(&m),  // Request hash
            // Initialize quorums for different phases
            Q1:        PaxiBFT.NewQuorum(),
            Q2:        PaxiBFT.NewQuorum(),
            Q3:        PaxiBFT.NewQuorum(),
            Q4:        PaxiBFT.NewQuorum(),
        }
    }
    e = p.log[p.slot]
    e.command = m.Command
    e.request = &m

    log.Debugf("p.slot = %v ", p.slot)
    log.Debugf("Key = %v ", m.Command.Key)
    
    // Execute if already committed
    if e.commit {
        log.Debugf("Executed")
        p.exec()
    }

    // Calculate request size for monitoring
    requestSize := unsafe.Sizeof(m)
    if m.Command.Value != nil {
        requestSize += uintptr(len(m.Command.Value))
    }
    for k, v := range m.Properties {
        requestSize += uintptr(len(k)) + uintptr(len(v))
    }
    log.Debugf("Received request of size: %d bytes", requestSize)

    // Check if this replica is the leader for this view
    Node_ID := PaxiBFT.ID(strconv.Itoa(1) + "." + strconv.Itoa(1))
    log.Debugf("Node_ID = %v", Node_ID)
    
    if Node_ID == p.ID() {
        e.active = true
    }
    
    // Leader specific handling
    if e.active {
        log.Debugf("The view leader : %v ", p.ID())
        e.Leader = true
        p.ballot.Next(p.ID())     // Advance ballot
        p.view.Next(p.ID())       // Advance view
        p.requests = append(p.requests, &m)
        p.Pbft.HandleRequest(m, p.slot)
    }
    
    // Update request status and check for execution
    e.Rstatus = RECEIVED
    if e.Cstatus == COMMITTED && e.Pstatus == PREPARED && e.Rstatus == RECEIVED {
        e.commit = true
        p.exec()
    }
}