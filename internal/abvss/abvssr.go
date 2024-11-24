package abvss

import (
	"Chaord/pkg/crypto/elgamal"
	"errors"
	"go.dedis.ch/kyber/v4"
	"log"
	"math/big"
	"math/rand"

	"Chaord/pkg/utils"
	"Chaord/pkg/utils/polynomial"
)

const ReceiverRandSeed = 10

func (vss *ABVSS) ReceiverInit(sk kyber.Scalar) {
	vss.Receiver = &Receiver{
		sk:           sk,
		fShares:      make([]*big.Int, vss.batchSize),
		gShares:      make([]*big.Int, vss.vNum),
		xix:          make([][]kyber.Point, vss.nodeNum),
		xiy:          make([][]kyber.Point, vss.nodeNum),
		zix:          make([][]kyber.Point, vss.nodeNum),
		ziy:          make([][]kyber.Point, vss.nodeNum),
		Received:     make(chan bool),
		randomBeacon: rand.New(rand.NewSource(ReceiverRandSeed)),
	}
}

func (vss *ABVSS) SetShares(fShares, gShares []*big.Int) {
	vss.fShares = fShares
	vss.gShares = gShares
}

func (vss *ABVSS) GetShares() ([]*big.Int, []*big.Int) {
	return vss.fShares, vss.gShares
}

func (vss *ABVSS) obtainShares(zix, ziy, xix, xiy []kyber.Point, index int) error {
	if vss.Receiver == nil {
		return errors.New("not a receiver")
	}
	if len(zix) != vss.batchSize || len(ziy) != vss.batchSize {
		return errors.New("insufficient zi")
	}
	if len(xix) != vss.vNum || len(xix) != vss.vNum {
		return errors.New("insufficient xi")
	}
	vss.zix[index] = zix
	vss.ziy[index] = ziy
	vss.xix[index] = xix
	vss.xiy[index] = xiy
	if index == vss.nodeID {
		for i := 0; i < vss.batchSize; i++ {
			tmp, err := elgamal.Decrypt(vss.curve, vss.sk, zix[i], ziy[i])

			if err != nil {
				/*
					log.Printf("wrong zi %v", zi[i])
					return errors.Join(errors.New("decrypt zi failed"), err)*/
				vss.fShares[i] = utils.RandomNum(vss.p)
			} else {
				vss.fShares[i] = new(big.Int).SetBytes(tmp)
			}
			log.Printf("vss.fShares[i] %v", vss.fShares[i])
		}
		for i := 0; i < vss.vNum; i++ {

			tmp, err := elgamal.Decrypt(vss.curve, vss.sk, xix[i], xiy[i])

			if err != nil {
				/*
					log.Printf("wrong xi %v", xi[i])
					return errors.Join(errors.New("decrypt xi failed"), err)*/
				vss.gShares[i] = utils.RandomNum(vss.p)
			} else {
				vss.gShares[i] = new(big.Int).SetBytes(tmp)
			}
			//vss.gShares[i] = xi[i]
		}
		vss.Received <- true
	}
	return nil
}

func (vss *ABVSS) GetLCM() LcmTuple {
	lcm, err := vss.constructLCM()
	if err != nil {
		log.Printf("vss.constructLCM err %v", err)
	}
	return lcm
}

func (vss *ABVSS) constructLCM() (LcmTuple, error) {
	if vss.Receiver == nil {
		return LcmTuple{}, errors.New("not a receiver")
	}
	lcm := make([]*big.Int, vss.vNum)
	r := make([][]*big.Int, vss.vNum)
	for i := 0; i < vss.vNum; i++ {
		r[i] = make([]*big.Int, vss.batchSize)
		for j := 0; j < vss.batchSize; j++ {
			r[i][j] = new(big.Int).Mod(new(big.Int).SetInt64(vss.randomBeacon.Int63()), vss.p)
			//fmt.Printf("%v %v %v\n", i, j, r[i][j])
		}
	}
	for i := 0; i < vss.vNum; i++ {
		//fmt.Println(r[i])
		tmp, err := utils.DotProduct(vss.fShares, r[i])
		//fmt.Printf("node %v get fShares %v\n", vss.nodeID, vss.fShares)
		if err != nil {
			return LcmTuple{}, err
		}
		lcm[i] = new(big.Int).Mod(new(big.Int).Add(tmp, vss.gShares[i]), vss.p)
		//fmt.Printf("node %v %v li:%v\n", vss.nodeID, i, lcm[i])
	}
	tuple := LcmTuple{
		index: vss.nodeID,
		lcm:   lcm,
	}
	return tuple, nil
}

func (vss *ABVSS) getRecoverShares(sk kyber.Scalar, index int, r [][]*big.Int) error {
	fj := make([]*big.Int, vss.batchSize)
	for i := 0; i < vss.batchSize; i++ {
		tmp, err := elgamal.Decrypt(vss.curve, sk, vss.zix[index][i], vss.ziy[index][i])
		if err != nil {
			return err
		}
		fj[i] = new(big.Int).SetBytes(tmp)
	}
	gj := make([]*big.Int, vss.vNum)
	for i := 0; i < vss.vNum; i++ {
		tmp, err := elgamal.Decrypt(vss.curve, sk, vss.xix[index][i], vss.xiy[index][i])
		if err != nil {
			return err
		}
		fj[i] = new(big.Int).SetBytes(tmp)
	}
	lcm := make([]*big.Int, vss.vNum)
	for i := 0; i < vss.vNum; i++ {
		tmp, err := utils.DotProduct(fj, r[i])
		if err != nil {
			return err
		}
		lcm[i] = new(big.Int).Add(tmp, gj[i])
	}
	if true {
		vss.qlist[index] = fj
	}
	return nil
}

func (vss *ABVSS) sendComplain() error {

	return errors.New("do not need to complain")
}

func (vss *ABVSS) handleComplain(sk kyber.Scalar, index int) error {

	return errors.New("do not need to complain")
}

func (vss *ABVSS) shareRecovery() error {
	if vss.Receiver == nil {
		return errors.New("not a receiver")
	}
	if !vss.complain {
		return errors.New("not a complain node")
	}
	if len(vss.qlist) < vss.degree+1 {
		return errors.New("invalid Q list")
	}
	xlist := make([]*big.Int, len(vss.qlist))
	ylist := make([][]*big.Int, vss.batchSize)
	for i, index := range vss.jlist {
		xlist[i] = new(big.Int).SetInt64(int64(index))
		ylist[i] = vss.qlist[index]
	}
	for i := 0; i < vss.batchSize; i++ {
		f, err := polynomial.LagrangeInterpolation(xlist, ylist[i], vss.p)
		if err != nil {
			return err
		}
		vss.fShares[i] = f.EvalMod(new(big.Int).SetInt64(int64(vss.nodeID)), vss.p)
	}
	return nil
}
