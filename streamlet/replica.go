package streamlet

import (
	"fmt"
	"github.com/salemmohammed/PaxiBFT"
	"github.com/salemmohammed/PaxiBFT/log"
	"strconv"
	"sync"
	"time"
)

type Replica struct {
	PaxiBFT.Node
	*Streamlet
	mux sync.Mutex
}
const (
	HTTPHeaderSlot       = "Slot"
	HTTPHeaderBallot     = "Ballot"
	HTTPHeaderExecute    = "Execute"
	HTTPHeaderInProgress = "Inprogress"
)
var t int
func NewReplica(id PaxiBFT.ID) *Replica {
	log.Debugf("Replica started \n")
	r := new(Replica)
	r.Node = PaxiBFT.NewNode(id)
	r.Streamlet = NewStreamlet(r)
	r.Register(PaxiBFT.Request{}, r.handleRequest)
	r.Register(Propose{},         r.handlePropose)
	r.Register(Vote{},  		  r.HandleVote)
	r.Register(RoundRobin{},       r.handleRound)

	return r
}
func (p *Replica) handleRequest(m PaxiBFT.Request) {
	log.Debugf("\n<-----------handleRequest----------->\n")
	if p.slot <= 0 {
		fmt.Print("-------------------streamlet-------------------------")
	}
	p.slot++
	if p.slot % 1000 == 0 {fmt.Print("p.slot", p.slot)}
	//p.Requests = append(p.Requests, &m)
	e, ok := p.log[p.slot]
	if !ok{
		p.log[p.slot] = &entry{
			Ballot:    	p.ballot,
			request:   	&m,
			Timestamp: 	time.Now(),
			Q1:			PaxiBFT.NewQuorum(),
			Q2: 		PaxiBFT.NewQuorum(),
			Q3: 		PaxiBFT.NewQuorum(),
			active:     false,
			Leader:		false,
			commit:    	false,
			Sent:       false,
			MyTurn:     false,
		}
	}
	e = p.log[p.slot]
	t = e.Q1.Total()
	log.Debugf("e.request= %v" , e.request)
	e.request = &m
	w := p.slot % e.Q1.Total() + 1
	p.Node_ID = PaxiBFT.ID(strconv.Itoa(1) + "." + strconv.Itoa(w))

	log.Debugf("Node_ID = %v", p.Node_ID)

	if p.Node_ID == p.ID(){
		log.Debugf("leader")
		e.active = true
		e.Leader = true
		p.ballot.Next(p.ID())
		log.Debugf("p.ballot %v ", p.ballot)
		e.Ballot = p.ballot
		e.Pstatus = PREPARED
	}

	if (p.ID() == p.Node_ID){

		log.Debugf("p.slot module e.Q2.Total1() == 0 = %v ", (p.slot % e.Q2.Total1()))
		if p.slot % e.Q2.Total1() == 0 && e.Sent == false{
			time.Sleep(50 * time.Millisecond)
			e.Sent = true
			log.Debugf("slot = %v", p.slot)
			log.Debugf("t = %v", t)
			p.HandleRequest(m,p.slot,t)
		}
		log.Debugf(" value = %v, e.MyTurn = %v ", e.slot , e.MyTurn)
		if p.slot > 0 && e.MyTurn == true && e.Sent == false{
			e.Sent = true
			time.Sleep(50 * time.Millisecond)
			log.Debugf("slot = %v", p.slot)
			p.HandleRequest(m,p.slot,t)
		}
	}
	e.Rstatus = RECEIVED
	log.Debugf("e.Pstatus = %v", e.Pstatus)
	if e.Cstatus == COMMITTED && e.Pstatus == PREPARED && e.Rstatus == RECEIVED{
		log.Debug("late call")
		e.commit = true
		p.exec()
	}
}
func (p *Streamlet) handleRound(m RoundRobin) {
	log.Debugf("\n<-----------handleRound----------->\n")
	log.Debugf("p.requests = %v ", p.Requests)
	log.Debugf("m.Slot = %v ", m.Slot)
	log.Debugf("p.id = %v ", m.Id)
	if p.slot >= m.Slot {
		log.Debugf("p.slot >= m.Slot")
		e, ok := p.log[m.Slot]
		if !ok {
			log.Debugf("!ok")
		}else{
			log.Debugf("e.Sent %v ", e.Sent)
			if e.Sent == false {
				p.HandleRequest(*e.request, m.Slot, t)
			}
		}
	}else {
		log.Debugf("p.slot < m.Slot")
		_, ok := p.log[m.Slot]
		if !ok {
			log.Debugf("created")
			p.log[m.Slot] = &entry{
				Ballot:    	p.ballot,
				request:   	&m.Request,
				Timestamp: 	time.Now(),
				Q1:			PaxiBFT.NewQuorum(),
				Q2: 		PaxiBFT.NewQuorum(),
				Q3: 		PaxiBFT.NewQuorum(),
				active:     false,
				Leader:		false,
				commit:    	false,
				Sent:      false,
				MyTurn:    true,
			}
		}
	}
}