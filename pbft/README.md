Introduction (pbft)
msg.go This file defines the message types and structures used in the PBFT protocol, including requests, pre-prepares, prepares, commits, and acknowledgments. It ensures consistent message handling and serialization for effective communication among replicas.
pbft.go This file implements the PBFT consensus algorithm, managing the overall operation of the protocol. It coordinates message flow, initiates consensus processes, and handles state transitions between different phases of PBFT.
replica.go This file represents a PBFT replica, encapsulating its behavior and functionalities. It processes incoming messages, manages local state updates, and responds to client requests, allowing replicas to actively participate in the consensus process.
File: msg.go
Overview
The msg.go file is integral to the PBFT (Practical Byzantine Fault Tolerance) protocol, serving as the foundation for message handling and communication between replicas. It defines the essential message types used in the protocol, ensuring that they are appropriately structured and can be serialized for efficient transmission.
Package Declaration
package pbft
This line declares that the file belongs to the pbft package, which is responsible for implementing the PBFT protocol.
Imports
import (
	"encoding/gob"
	"fmt"
	"github.com/salemmohammed/PaxiBFT"
)
encoding/gob: Provides support for encoding and decoding data in Go's binary format, allowing for efficient serialization of messages.
fmt: Offers formatted I/O functions for outputting strings and other types.
github.com/salemmohammed/PaxiBFT: Imports the PaxiBFT package, which contains fundamental types and structures used within the PBFT implementation.
Initialization Function
func init() {
	gob.Register(PrePrepare{})
	gob.Register(Prepare{})
	gob.Register(Commit{})
}
The init function registers the message types (PrePrepare, Prepare, and Commit) with the gob encoder/decoder. This registration is essential for the correct serialization and deserialization of these types when communicating between replicas.
Message Types
PrePrepare Message
type PrePrepare struct {
    Ballot      PaxiBFT.Ballot
    ID          PaxiBFT.ID
    View        PaxiBFT.View
    Slot        int
    Request     PaxiBFT.Request
    Digest      []byte
    ActiveView  bool
    Command     PaxiBFT.Command
}


Description: The PrePrepare structure represents the first message in the PBFT consensus process. It contains information about the proposed command, the ballot number, the view of the system, and a digest of the command.
Fields:
Ballot: The current ballot being proposed.
ID: The identifier of the replica sending the message.
View: The current view number of the system.
Slot: The index slot for the request.
Request: The original request from the client.
Digest: A hash digest of the request, ensuring integrity.
ActiveView: A boolean indicating if the view is active.
Command: The command associated with the request.
String Representation:
func (m PrePrepare) String() string {
    return fmt.Sprintf("PrePrepare {Ballot=%v , View=%v, slot=%v, Command=%v}", m.Ballot,m.View,m.Slot,m.Command)
}
This method provides a formatted string representation of the PrePrepare message, which is useful for debugging and logging.
Prepare Message

type Prepare struct {
    Ballot      PaxiBFT.Ballot
    ID          PaxiBFT.ID
    View        PaxiBFT.View
    Slot        int
    Digest      []byte
    Command     PaxiBFT.Command
    Request     PaxiBFT.Request
}


Description: The Prepare structure is used to signal that a replica is ready to commit a proposal. It confirms the receipt of a PrePrepare message and includes the necessary information to ensure agreement among replicas.
Fields:
Ballot, ID, View, Slot, Digest, Command, and Request: Similar to the PrePrepare structure but indicate readiness to proceed with the consensus.
String Representation:


func (m Prepare) String() string {
    return fmt.Sprintf("Prepare {Ballot=%v, ID=%v, View=%v, slot=%v, command=%v}", m.Ballot,m.ID,m.View,m.Slot,m.Command)
}
This method provides a string representation of the Prepare message.
Commit Message

type Commit struct {
    Ballot      PaxiBFT.Ballot
    ID          PaxiBFT.ID
    View        PaxiBFT.View
    Slot        int
    Digest      []byte
    Command     PaxiBFT.Command
    Request     PaxiBFT.Request
}


Description: The Commit structure is sent by replicas to indicate that a decision has been reached and that the command can be executed. It is crucial for finalizing the consensus process.
Fields: Same as the Prepare message, confirming the completion of the consensus process.
String Representation:


func (m Commit) String() string {
    return fmt.Sprintf("Commit {Ballot=%v, ID=%v, View=%v, Slot=%v, command=%v}", m.Ballot,m.ID,m.View,m.Slot, m.Command)
}
This method returns a string representation of the Commit message.


File: replica.go
Overview
The replica.go file implements the Replica structure, which represents a single node within the PBFT (Practical Byzantine Fault Tolerance) consensus protocol. This file handles the reception of requests, manages state transitions, and coordinates message handling for the PBFT process. The Replica structure plays a crucial role in ensuring that all nodes reach consensus on the execution of commands.
Package Declaration

package pbft

This line indicates that the file is part of the pbft package, which contains the implementation of the PBFT protocol.
Imports

import (
	"fmt"
	"github.com/salemmohammed/PaxiBFT"
	"github.com/salemmohammed/PaxiBFT/log"
	"strconv"
	"time"
)

fmt: Provides formatted I/O functions for outputting strings and other types.
github.com/salemmohammed/PaxiBFT: Imports the PaxiBFT package, which contains fundamental types and structures for the protocol.
github.com/salemmohammed/PaxiBFT/log: Imports the logging package for logging debug messages.
strconv: Contains functions for converting string representations of numbers.
time: Provides functionality for measuring and displaying time.
Constants
go

const (
	HTTPHeaderSlot    = "Slot"
	HTTPHeaderBallot  = "Ballot"
	HTTPHeaderExecute = "Execute"
)

These constants define HTTP header fields used in the PBFT communication, providing clear identifiers for different types of messages.
Replica Structure

type Replica struct {
	PaxiBFT.Node
	*Pbft
}

The Replica struct embeds the Node structure from the PaxiBFT package and contains a pointer to the Pbft structure. This allows the Replica to utilize the node functionalities while implementing the PBFT protocol.
Constructor

func NewReplica(id PaxiBFT.ID) *Replica {
	r := new(Replica)

	r.Node = PaxiBFT.NewNode(id)
	r.Pbft = NewPbft(r)

	r.Register(PaxiBFT.Request{}, r.handleRequest)
	r.Register(PrePrepare{}, r.HandlePre)
	r.Register(Prepare{}, r.HandlePrepare)
	r.Register(Commit{}, r.HandleCommit)

	return r
}

Description: The NewReplica function creates and initializes a new Replica instance.
Parameters:
id: The identifier of the replica.
Functionality:
Initializes the embedded Node and Pbft instances.
Registers handlers for different message types, ensuring that the replica can respond appropriately to incoming requests.
Handle Request

func (p *Replica) handleRequest(m PaxiBFT.Request) {
	log.Debugf("<---------------------handleRequest------------------------->")

	p.slot++ // slot number for every request
	if p.slot%1000 == 0 {
		fmt.Print("p.slot", p.slot)
	}

Description: This method processes incoming requests and updates the internal state of the replica.
Functionality:
Increments the slot number for each request received.
Logs the current slot number at intervals of 1000 for monitoring purposes.

	e, ok := p.log[p.slot]
	if !ok {
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
		}
	}
	e = p.log[p.slot]
	e.command = m.Command
	e.request = &m

State Management:
Retrieves the log entry for the current slot. If it does not exist, a new entry is created, initializing various fields like ballot, view, command, and others.
Updates the command and request for the current log entry.

	log.Debugf("p.slot = %v ", p.slot)
	log.Debugf("Key = %v ", m.Command.Key)
	if e.commit {
		log.Debugf("Executed")
		p.exec()
	}
	if p.slot == 0 {
		fmt.Println("-------------------PBFT-------------------------")
	}

Logging: Logs the current slot and command key, helping trace the flow of requests and their handling.

	Node_ID := PaxiBFT.ID(strconv.Itoa(1) + "." + strconv.Itoa(1))
	log.Debugf("Node_ID = %v", Node_ID)
	if Node_ID == p.ID() {
		e.active = true
	}
	if e.active {
		log.Debugf("The view leader : %v ", p.ID())
		e.Leader = true
		p.ballot.Next(p.ID())
		p.view.Next(p.ID())
		p.requests = append(p.requests, &m)
		p.Pbft.HandleRequest(m, p.slot)
	}

Leader Election:
Checks if the replica is the leader for the current view and marks it as active.
Updates the ballot and view, and adds the request to the replica’s request queue.

	e.Rstatus = RECEIVED
	if e.Cstatus == COMMITTED && e.Pstatus == PREPARED && e.Rstatus == RECEIVED {
		e.commit = true
		p.exec()
	}
}

Commit Logic:
Updates the status of the request and checks if the entry is ready to be committed. If all conditions are met, the command is executed.
pbft.go File Explanation
1. Introduction
The pbft.go file is part of the implementation of the PBFT (Practical Byzantine Fault Tolerance) protocol, which is used in distributed systems to achieve consensus among nodes even in the presence of unreliable nodes. This file includes definitions for the structures and functions that support the core operations of the protocol.
2. Imported Packages

import (
	"crypto/md5"
	"github.com/salemmohammed/PaxiBFT"
	"github.com/salemmohammed/PaxiBFT/log"
	"time"
)

crypto/md5: Used to generate MD5 hashes for messages.
github.com/salemmohammed/PaxiBFT: The core package that contains definitions related to the protocol.
github.com/salemmohammed/PaxiBFT/log: A custom logging library for recording information and messages.
time: Used for handling date and time.
3. Message Status Definitions

type status int8

const (
	NONE status = iota
	PREPREPARED
	PREPARED
	COMMITTED
	RECEIVED
)

status: Represents the status of messages in the PBFT protocol. It includes states like:
NONE: No specific status.
PREPREPARED: The message is in the pre-preparation stage.
PREPARED: The message has been prepared.
COMMITTED: The message has been committed.
RECEIVED: The message has been received.
4. Entry Structure Definition

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
	Pstatus    status
	Cstatus    status
	Rstatus    status
}

entry: Structure representing an entry in the log, including:
ballot: The ballot number.
view: The view number.
command: The command being executed.
commit: A flag indicating whether the entry has been committed.
active: A flag indicating whether the entry is active.
Leader: A flag indicating whether the node is the leader.
request: The request associated with the entry.
timestamp: The time the entry was created.
Digest: The hash of the entry.
Q1, Q2, Q3, Q4: Quorum sets used in the protocol.
Pstatus, Cstatus, Rstatus: Various statuses related to the entry.
5. PBFT Structure Definition

type Pbft struct {
	PaxiBFT.Node

	config          []PaxiBFT.ID
	N               PaxiBFT.Config
	log             map[int]*entry       // log ordered by slot

	slot            int                  // highest slot number
	view            PaxiBFT.View            // view number
	ballot          PaxiBFT.Ballot          // highest ballot number
	execute         int                  // next execute slot number
	requests        []*PaxiBFT.Request
	quorum          *PaxiBFT.Quorum // phase 1 quorum
	ReplyWhenCommit bool
	RecivedReq      bool
	Member          *PaxiBFT.Memberlist
}

Pbft: Structure representing the PBFT protocol implementation, including:
config: A list of node IDs in the system.
N: Protocol configuration settings.
log: Log of entries ordered by slot number.
slot: The highest slot number.
view: The current view number.
ballot: The highest ballot number.
execute: The next execute slot number.
requests: A list of requests.
quorum: Quorum for phase 1.
ReplyWhenCommit: Flag to indicate if a reply should be sent upon commit.
RecivedReq: Flag to indicate if a request has been received.
Member: The member list.
6. Creating a New PBFT Instance

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

NewPbft: Function to create a new instance of the Pbft structure. It accepts a node as input and additional options for customizing the instance settings.
7. Generating Message Hash

func GetMD5Hash(r *PaxiBFT.Request) []byte {
	hasher := md5.New()
	hasher.Write([]byte(r.Command.Value))
	return []byte(hasher.Sum(nil))
}

GetMD5Hash: Function to generate an MD5 hash for a specific command from a request.
8. Handling Requests

func (p *Pbft) HandleRequest(r PaxiBFT.Request, s int) {
	log.Debugf("<--------------------HandleRequest------------------>")

	e := p.log[s]
	e.Digest = GetMD5Hash(&r)
	log.Debugf("[p.ballot.ID %v, p.ballot %v ]", p.ballot.ID(), p.ballot)
	log.Debugf("PrePrepare will be called")
	p.PrePrepare(&r, &e.Digest, s)
}

HandleRequest: Function to handle a request. It updates the hash of the request and calls the PrePrepare function.
9. Pre-Preparation Phase

func (p *Pbft) PrePrepare(r *PaxiBFT.Request, s *[]byte, slt int) {
	log.Debugf("<--------------------PrePrepare------------------>")

	_, ok := p.log[p.slot]
	if !ok {
		p.log[p.slot] = &entry{
			ballot:    p.ballot,
			view:      p.view,
			command:   r.Command,
			commit:    false,
			active:    false,
			Leader:    false,
			request:   r,
			timestamp: time.Now(),
			Digest:    GetMD5Hash(r),
			Q1:        PaxiBFT.NewQuorum(),
			Q2:        PaxiBFT.NewQuorum(),
			Q3:        PaxiBFT.NewQuorum(),
			Q4:        PaxiBFT.NewQuorum(),
		}
	}

	p.Broadcast(PrePrepare{
		Ballot:     p.ballot,
		ID:         p.ID(),
		View:       p.view,
		Slot:       slt,
		Request:    *r,
		Digest:     *s,
		Command:    r.Command,
	})
	log.Debugf("++++++ PrePrepare Done ++++++")
}

PrePrepare: Function to initiate the first phase of the protocol. It sends a PrePrepare message to the other nodes.
10. Handling PrePrepare Messages

func (p *Pbft) HandlePre(m PrePrepare) {
	log.Debugf("<--------------------HandlePre------------------>")

	log.Debugf(" Sender  %v ", m.ID)

	log.Debugf(" m.Slot  %v ", m.Slot)

	if m.Ballot > p.ballot {
		log.Debugf("m.Ballot > p.ballot")
		p.ballot = m.Ballot
		p.view = m.View
	}

	e, ok := p.log[m.Slot]
	if !ok {
		log.Debugf("Create a log")
		p.log[m.Slot] = &entry{
			ballot:    p.ballot,
			view:      p.view,

HandlePre: Function to handle incoming PrePrepare messages. It checks the ballot number and updates the log accordingly.