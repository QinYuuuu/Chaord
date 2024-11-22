package chaord

import (
	"log"
	"math/big"
)

var BigTwo = new(big.Int).SetInt64(2)
var BigOne = new(big.Int).SetInt64(1)

func LocalTest(nodeNum, degree, dataScale, sampleScale, distributeParam int, p *big.Int) {
	//curve := edwards25519.NewBlakeSHA256Ed25519()
	//fmt.Print(curve.ScalarLen())
	param := NewParam(nodeNum, degree, dataScale, sampleScale, p)
	owner := NewOwner(param, false)
	consumer := NewConsumer(param, false)
	nodes := make([]*Node, nodeNum)
	for i := 0; i < nodeNum; i++ {
		nodes[i] = NewNode(param, distributeParam, false)
	}

	R, rShares, rSquareShares := batchDDGOffline(nodeNum, degree, p)

	//var ownerTime, consumerTime, nodeTime time.Duration
	//var start, end time.Time
	index := make([]*big.Int, nodeNum)
	for node := 0; node < nodeNum; node++ {
		index[node] = new(big.Int).SetInt64(int64(node + 1))
	}

	// step1 and step2 refactor from DGG_local_20241022.py
	step1 := func() (*big.Int,[]*big.Int) {
		// owner select random value on Z_2, run secret sharing on Z_p among nodes
		bo, boShares := owner.batchDDGInit()
		boAddRShares := make([]*big.Int, nodeNum)

		// node verify s^2-s
		for node := 0; node < nodeNum; node++ {
			boAddRShares[node] = new(big.Int).Add(boShares[node], rShares[node])
		}
		boAddR := reconstruct(index, boAddRShares, p)
		value := new(big.Int).Mul(boAddR, boAddR)
		boAddRSquareShares := shareSecretBig(value, nodeNum, degree, p)
		boAddRSquare := reconstruct(index, boAddRSquareShares, p)
		tmp := new(big.Int).Add(bo, R)
		tmp2 := new(big.Int).Exp(tmp, BigTwo, p)
		if boAddR.Cmp(tmp) != 0 {
			log.Printf("Chaord - bo_add_R error: want %v, get %v\n ", tmp, boAddR)
		}
		if boAddRSquare.Cmp(tmp2) != 0 {
			log.Printf("Chaord - bo_add_R_square error: want %v, get %v\n ", tmp2, boAddRSquare)
		}
		mpcShares := make([]*big.Int, nodeNum)
		for node := 0; node < nodeNum; node++ {
			mid := new(big.Int).Mul(rShares[node], boAddR)
			mid.Mul(mid, BigTwo)
			mid.Add(mid, boShares[node])
			mid.Sub(boAddRSquareShares[node], mid)
			mid.Add(mid, rSquareShares[node])
			mpcShares[node] = mid.Mod(mid, p)

		}

		mpcResult := reconstruct(index, mpcShares, p)
		if mpcResult.Int64() != 0 {
			log.Printf("Chaord - MPC error, %v\n", mpcResult)
		}
		return bo, boShares
	}
	bo, boShares := step1()

	step2 := func() {
		// consumer select random value on Z_2, run secret sharing on Z_p among nodes
		bc, bcShares := consumer.batchDDGInit()
		bcReconstruct := reconstruct(index, bcShares, p)
		if bc.Cmp(bcReconstruct) != 0 {
			log.Printf("Chaord - bc_reconstruct error, get %v, want %v\n", bcReconstruct, bc)
		}
		bcAddboShares := make([]*big.Int, nodeNum)
		// get 1 unbias bit share
		for i := 0; i < nodeNum; i++ {
			if bcReconstruct.Int64() == 0 {
				bcAddboShares[i] = boShares[i]
			} else {
				bcAddboShares[i] = new(big.Int).Sub(BigOne, boShares[i])
				bcAddboShares[i].Mod(bcAddboShares[i], p)
			}
		}
		bcAddbo := reconstruct(index, bcAddboShares, p)
		tmp := new(big.Int).Add(bc, bo)
		tmp.Mod(tmp, BigTwo)
		if bcAddbo.Cmp(tmp) != 0{

		}
	}
	step2()

	bUnbiasShares := make([][]*big.Int, nodeNum)

	step3 := func() {

	}
}
