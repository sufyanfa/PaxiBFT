package pbft
import (
    "encoding/gob"
    "fmt"
    "github.com/salemmohammed/PaxiBFT"
)

// init registers message types for serialization using the gob package
// This is necessary for network communication between nodes
func init() {
    gob.Register(PrePrepare{})
    gob.Register(Prepare{})
    gob.Register(Commit{})
}

// PrePrepare represents the first phase message in PBFT protocol
// Sent by the primary/leader to all backup replicas
type PrePrepare struct {
    Ballot      PaxiBFT.Ballot  // Current ballot number for ordering
    ID          PaxiBFT.ID      // ID of the sending node
    View        PaxiBFT.View    // Current view number
    Slot        int             // Sequence number for this request
    Digest      []byte          // Hash/digest of the request
    ActiveView  bool            // Indicates if this view is active
}

// String returns a formatted string representation of PrePrepare message
// Used for logging and debugging
func (m PrePrepare) String() string {
    return fmt.Sprintf("PrePrepare {Ballot=%v , View=%v, slot=%v}", m.Ballot,m.View,m.Slot)
}

// Prepare represents the second phase message in PBFT protocol
// Sent by replicas to all other replicas to show agreement on request
type Prepare struct {
    Ballot  PaxiBFT.Ballot  // Current ballot number
    ID      PaxiBFT.ID      // ID of the sending node
    View    PaxiBFT.View    // Current view number
    Slot    int             // Sequence number for this request
    Digest  []byte          // Hash/digest of the request
}

// String returns a formatted string representation of Prepare message
// Used for logging and debugging
func (m Prepare) String() string {
    return fmt.Sprintf("Prepare {Ballot=%v, ID=%v, View=%v, slot=%v}", m.Ballot,m.ID,m.View,m.Slot)
}

// Commit represents the final phase message in PBFT protocol
// Sent by replicas to all other replicas to commit the request
type Commit struct {
    Ballot   PaxiBFT.Ballot    // Current ballot number
    ID       PaxiBFT.ID        // ID of the sending node
    View     PaxiBFT.View      // Current view number
    Slot     int               // Sequence number for this request
    Digest   []byte            // Hash/digest of the request
    Command  PaxiBFT.Command   // The actual command to be executed
    Request  PaxiBFT.Request   // Original client request
}

// String returns a formatted string representation of Commit message
// Used for logging and debugging
func (m Commit) String() string {
    return fmt.Sprintf("Commit {Ballot=%v, ID=%v, View=%v, Slot=%v, command=%v}", m.Ballot,m.ID,m.View,m.Slot, m.Command)
}