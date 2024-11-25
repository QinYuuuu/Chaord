package osv

import (
	"log"
)

func RouteLocal(osv []*Node) {
	for i := range osv {
		go func(i int) {
			for {
				msg := <-osv[i].SendChan
				osv[msg.DestID].ReceiveChan <- msg
			}
		}(i)
	}
}

func (osv *Node) Run() {
	var msg Message
	for {
		msg = <-osv.ReceiveChan
		recvMsgs, err := osv.Recv(msg)
		if err != nil {
			log.Printf("[node %v] error in osv: %s", osv.id, err.Error())
			return
		}
		for _, newMsg := range recvMsgs {
			// two int, one char
			osv.bandwidth += 9
			osv.SendChan <- newMsg
		}
	}
}
