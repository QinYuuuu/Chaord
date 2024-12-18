package osv

import (
	"errors"
)

const Echo string = "E"
const Vote string = "V"

type Message struct {
	FromID int
	DestID int
	Mtype  string
}

func (m Message) Dest() int {
	return m.DestID
}

func (m Message) From() int {
	return m.FromID
}

func (m Message) Type() string {
	return m.Mtype
}

type Node struct {
	n        int
	t        int
	id       int
	echosNum int
	votesNum int
	nVotes   []bool
	acquired bool
	voted    bool
	done     bool
	OutPut   chan bool

	bandwidth int

	SendChan    chan Message
	ReceiveChan chan Message
}

func NewOSV(n, t, id int) *Node {
	return &Node{
		n:           n,
		t:           t,
		id:          id,
		echosNum:    0,
		votesNum:    0,
		nVotes:      make([]bool, n),
		acquired:    false,
		voted:       false,
		done:        false,
		OutPut:      make(chan bool),
		SendChan:    make(chan Message, 100),
		ReceiveChan: make(chan Message, 100),
	}
}

func (osv *Node) InitSync() []Message {
	var msgs []Message
	for i := 0; i < osv.n; i++ {
		msg := Message{}
		msg.FromID = osv.id
		msg.DestID = i
		msg.Mtype = Echo
		if i == osv.id {
			newmsgs := osv.Loop(msg)
			msgs = append(msgs, newmsgs...)
		} else {
			msgs = append(msgs, msg)
		}
	}
	return msgs
}

func (osv *Node) Init() {
	//log.Printf("node %v osv init", osv.id)
	var msgs []Message
	for i := 0; i < osv.n; i++ {
		msg := Message{}
		msg.FromID = osv.id
		msg.DestID = i
		msg.Mtype = Echo
		if i == osv.id {
			newmsgs := osv.Loop(msg)
			msgs = append(msgs, newmsgs...)
		} else {
			msgs = append(msgs, msg)
		}
	}
	for _, msg := range msgs {
		osv.bandwidth += 9
		osv.SendChan <- msg
	}
}

func (osv *Node) GetBandwidth() int {
	return osv.bandwidth
}

func (osv *Node) Done() bool {
	return osv.done
}

func (osv *Node) Loop(m Message) []Message {
	msgs, _ := osv.Recv(m)
	i := 0
	flag := len(msgs)
	for i < flag {
		if msgs[i].Dest() == osv.id {
			newmsgs, _ := osv.Recv(msgs[i])
			msgs = append(msgs[:i], newmsgs...)
			i = 0
			flag = len(msgs)
		}
		i++
	}
	return msgs
}

func (osv *Node) Recv(m Message) ([]Message, error) {
	var msgs []Message
	if m.DestID != osv.id {
		return nil, errors.New("wrong destination id")
	}
	if m.Mtype == Echo {
		//osv.handleEcho(m)
		//log.Printf("[node %v] received ECHO from node %v", osv.id, m.FromID)
		osv.echosNum += 1
	}
	if m.Mtype == Vote {
		//osv.handleVote(m)
		if osv.nVotes[m.FromID] {
			//log.Printf("node %v has already voted", m.fromID)
			return nil, nil
		}
		//log.Printf("[node %v] received VOTE from node %v", osv.id, m.FromID)
		osv.votesNum += 1
		osv.nVotes[m.FromID] = true
	}
	if osv.echosNum >= osv.n-osv.t && !osv.voted {
		for i := 0; i < osv.n; i++ {
			msg := Message{}
			msg.FromID = osv.id
			msg.DestID = i
			msg.Mtype = Vote
			if i == osv.id {
				newMsgs := osv.Loop(msg)
				msgs = append(msgs, newMsgs...)
			} else {
				msgs = append(msgs, msg)
			}
		}
		osv.voted = true
		return msgs, nil
	}
	if osv.votesNum >= osv.t+1 && !osv.voted {
		for i := 0; i < osv.n; i++ {
			msg := Message{}
			msg.FromID = osv.id
			msg.DestID = i
			msg.Mtype = Vote
			if i == osv.id {
				newMsgs := osv.Loop(msg)
				msgs = append(msgs, newMsgs...)
			} else {
				msgs = append(msgs, msg)
			}
		}
		osv.voted = true
		return msgs, nil
	}
	if osv.votesNum >= osv.n-osv.t && osv.voted {
		if osv.done == false {
			osv.done = true
			osv.OutPut <- true
			//log.Printf("[node %v] output %v", osv.id, osv.done)
		}
		return nil, nil
	}
	return nil, nil
}
