package pbft

import (
    "encoding/gob"
    "fmt"
    "github.com/salemmohammed/PaxiBFT"
)

// init registers message types for serialization
func init() {
    gob.Register(PrePrepare{})
    gob.Register(Prepare{})
    gob.Register(Commit{})
    gob.Register(DataMessage{})
}

// DataMessage represents a background data transmission message
type DataMessage struct {
    ID      PaxiBFT.ID          // ID of the sending node
    Slot    int                 // Sequence number
    Request *PaxiBFT.Request    // Full request data
    Digest  []byte              // Hash for verification
}

// PrePrepare represents the first phase message
type PrePrepare struct {
    Ballot      PaxiBFT.Ballot  // Current ballot number
    ID          PaxiBFT.ID      // ID of the sending node
    View        PaxiBFT.View    // Current view number
    Slot        int             // Sequence number
    Digest      []byte          // Hash of the request
    ActiveView  bool            // View status
}

func (m PrePrepare) String() string {
    return fmt.Sprintf("PrePrepare {Ballot=%v, View=%v, slot=%v}", 
        m.Ballot, m.View, m.Slot)
}

// Prepare represents the second phase message
type Prepare struct {
    Ballot  PaxiBFT.Ballot
    ID      PaxiBFT.ID
    View    PaxiBFT.View
    Slot    int
    Digest  []byte
}

func (m Prepare) String() string {
    return fmt.Sprintf("Prepare {Ballot=%v, ID=%v, View=%v, slot=%v}", 
        m.Ballot, m.ID, m.View, m.Slot)
}

// Commit represents the final phase message
type Commit struct {
    Ballot   PaxiBFT.Ballot
    ID       PaxiBFT.ID
    View     PaxiBFT.View
    Slot     int
    Digest   []byte
    Command  PaxiBFT.Command
    Request  PaxiBFT.Request
}

func (m Commit) String() string {
    return fmt.Sprintf("Commit {Ballot=%v, ID=%v, View=%v, Slot=%v, command=%v}", 
        m.Ballot, m.ID, m.View, m.Slot, m.Command)
}