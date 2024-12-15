package pbft

import (
	"fmt"
	"github.com/salemmohammed/PaxiBFT"
	"github.com/salemmohammed/PaxiBFT/log"
	"strconv"
	"time"
	"unsafe"
)

const (
	HTTPHeaderSlot    = "Slot"
	HTTPHeaderBallot  = "Ballot"
	HTTPHeaderExecute = "Execute"
)

type Replica struct {
	PaxiBFT.Node
	*Pbft
}

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

func (p *Replica) handleRequest(m PaxiBFT.Request) {
	log.Debugf("<---------------------handleRequest------------------------->")

	p.slot++ // slot number for every request
	if p.slot%1000 == 0 {
		fmt.Print("p.slot", p.slot)
	}

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
	// Ensure that e is used after it's updated
	e.command = m.Command
	e.request = &m

	// Calculate the hash of the request
	digest := GetMD5Hash(&m)
	
	// Store the hash in the entry's Digest field
	e.Digest = digest

	log.Debugf("p.slot = %v ", p.slot)
	log.Debugf("Key = %v ", m.Command.Key)
	if e.commit{
		log.Debugf("Executed")
		p.exec()
	}
	if p.slot == 0 {
		fmt.Println("-------------------PBFT-------------------------")
	}

	// Calculate the size of the Request struct itself
	requestSize := unsafe.Sizeof(m)

	// Add the size of the Command.Value field if it's a slice
	if m.Command.Value != nil {
		requestSize += uintptr(len(m.Command.Value))
	}

	// Add the size of the Properties map
	for k, v := range m.Properties {
		requestSize += uintptr(len(k)) + uintptr(len(v))
	}

	log.Debugf("Received request of size: %d bytes", requestSize)
	//w := p.slot % e.Q1.Total() + 1
	Node_ID := PaxiBFT.ID(strconv.Itoa(1) + "." + strconv.Itoa(1))
	log.Debugf("Node_ID = %v", Node_ID)
	if Node_ID == p.ID(){
		e.active = true
	}
	if e.active{
		log.Debugf("The view leader : %v ", p.ID())
		e.Leader = true
		p.ballot.Next(p.ID())
		p.view.Next(p.ID())
		p.requests = append(p.requests, &m)
		
		// Pass only the hash to HandleRequest
		p.Pbft.HandleRequest(e.Digest, p.slot)
	}
	e.Rstatus = RECEIVED
	if e.Cstatus == COMMITTED && e.Pstatus == PREPARED && e.Rstatus == RECEIVED{
		e.commit = true
		p.exec()
	}
}