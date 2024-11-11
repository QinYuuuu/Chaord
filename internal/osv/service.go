package osv

import (
	"log"
)

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
			osv.SendChan <- newMsg
		}
	}
}
