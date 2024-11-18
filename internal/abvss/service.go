package abvss

import (
	"Chaord/pkg/core"
	"Chaord/pkg/protobuf"
	"go.dedis.ch/kyber/v4"
	"log"
	"math/big"
)

func (vss *ABVSS) Run() {
	go vss.Receive()
}

func (vss *ABVSS) Receive() {
	for {
		//log.Printf("node %v waiting", dkg.id)
		msg := <-vss.ReceiveChan
		log.Printf("node %v handle msg: %v from node %v", vss.nodeID, msg.GetType(), msg.Sender)
		go func(msg *protobuf.Message) {
			msgType := msg.GetType()
			if msgType == "Shares" {
				newmsg := core.Decapsulation(msgType, msg).(*protobuf.SharesMsg)
				zixBytes := newmsg.GetZix()
				ziyBytes := newmsg.GetZiy()
				xiyBytes := newmsg.GetXiy()
				xixBytes := newmsg.GetXix()
				zix := make([]kyber.Point, len(zixBytes))
				ziy := make([]kyber.Point, len(ziyBytes))
				xix := make([]kyber.Point, len(xixBytes))
				xiy := make([]kyber.Point, len(xiyBytes))
				for i := range zixBytes {
					zix[i] = vss.curve.Point()
					ziy[i] = vss.curve.Point()
					_ = zix[i].UnmarshalBinary(zixBytes[i])
					_ = ziy[i].UnmarshalBinary(ziyBytes[i])
				}
				for i := range xixBytes {
					xix[i] = vss.curve.Point()
					xiy[i] = vss.curve.Point()
					_ = xix[i].UnmarshalBinary(xixBytes[i])
					_ = xiy[i].UnmarshalBinary(xiyBytes[i])
				}
				err := vss.ObtainShares(zix, ziy, xix, xiy, int(newmsg.GetIndex()))
				if err != nil {
					log.Printf("node %v receive shares from node %v error: %v", vss.GetNodeID(), newmsg.GetFromID(), err)
				}
				//log.Printf("node %v receive shares %v from node %v in instance %v", vss.GetNodeID(), newmsg.GetIndex(), newmsg.GetFromID(), newmsg.GetInstanceID())
			}
			if msgType == "LCM" {
				newmsg := core.Decapsulation(msgType, msg).(*protobuf.LCMMsg)
				vss.ReceiveLCM(newmsg)
			}
			if msgType == "SK" {

			}
		}(msg)

	}
}

func (vss *ABVSS) ReceiveLCM(lcmmsg *protobuf.LCMMsg) {
	lcmBytes := lcmmsg.GetLcmi()
	lcm := make([]*big.Int, len(lcmBytes))
	for i := range lcmBytes {
		lcm[i] = new(big.Int).SetBytes(lcmBytes[i])
	}
	tuple := LcmTuple{
		index: int(lcmmsg.GetFromID()),
		lcm:   lcm,
	}
	vss.shareCh <- tuple
	//log.Printf("node %v receive lcm from node %v", vss.GetNodeID(), lcmmsg.FromID)
	return
}

func (vss *ABVSS) SecretSharing(pk []kyber.Point, s []*big.Int) {
	err := vss.DistributorInit(pk, s)
	if err != nil {
		log.Printf("init error: %v", err)
	}
	err = vss.SamplePoly()
	if err != nil {
		log.Printf("sample poly error: %v", err)
	}
	for i := 0; i < vss.nodeNum; i++ {
		go func(i int) {
			zix, ziy, xix, xiy, err := vss.GenerateShares(i)
			if err != nil {
				log.Printf("generate shares error: %v", err)
			}

			zixBytes := make([][]byte, len(zix))
			ziyBytes := make([][]byte, len(ziy))
			xixBytes := make([][]byte, len(xix))
			xiyBytes := make([][]byte, len(xiy))
			for j := range zix {
				zixBytes[j], _ = zix[j].MarshalBinary()
				ziyBytes[j], _ = ziy[j].MarshalBinary()
			}
			for j := range xix {
				xixBytes[j], _ = xix[j].MarshalBinary()
				xiyBytes[j], _ = xiy[j].MarshalBinary()
			}
			sharesmsg := &protobuf.SharesMsg{
				FromID: int64(vss.nodeID),
				Index:  int64(i),
				Zix:    zixBytes,
				Ziy:    ziyBytes,
				Xix:    xixBytes,
				Xiy:    xiyBytes,
			}

			m := core.Encapsulation("Shares", nil, uint32(vss.nodeID), sharesmsg)

			//ctx, cancel := context.WithCancel(context.Background())
			//defer cancel()
			for j := 0; j < vss.nodeNum; j++ {
				if j == vss.nodeID {
					err := vss.ObtainShares(zix, ziy, xix, xiy, i)
					if err != nil {
						log.Printf("obtain shares error: %v", err)
					}
					continue
				}
				//log.Printf("node %v send shares to node %v", vss.nodeID, j)
				vss.SendChan[j] <- m
			}
		}(i)

	}
}

func (vss *ABVSS) BroadcastLCM() {
	tuple, err := vss.ConstructLCM()
	if err != nil {
		log.Printf("construct lcm error: %v", err)
	}
	lcm := tuple.lcm
	lcmBytes := make([][]byte, len(lcm))
	for i := range lcm {
		lcmBytes[i] = lcm[i].Bytes()
	}

	for i := 0; i < vss.nodeNum; i++ {
		lcmmsg := &protobuf.LCMMsg{
			FromID: int64(vss.nodeID),
			DestID: int64(i),
			Lcmi:   lcmBytes,
		}
		if i == vss.nodeID {
			err := vss.VerifyLCM(tuple)
			if err != nil {
				log.Printf("VerifyLCM error: %v", err)
			}
			continue
		}
		m := core.Encapsulation("LCM", nil, uint32(vss.nodeID), lcmmsg)
		vss.SendChan[i] <- m
	}
}
