package chaord

import (
	"Chaord/pkg/utils"
	"bytes"
	"fmt"
	"github.com/shopspring/decimal"
	"log"
	"math/big"
	"math/rand"
	"os"
	"time"
)

var BigTwo = new(big.Int).SetInt64(2)
var BigOne = new(big.Int).SetInt64(1)
var decimalTwo = decimal.NewFromInt(2)

func mod_decimal_normal(a, b decimal.Decimal) decimal.Decimal {
	a = a.Mod(b)
	if a.Cmp(b.Div(decimalTwo)) == 1 {
		return a.Sub(b)
	} else if a.Cmp(decimal.Zero.Sub(b.Div(decimalTwo))) == -1 {
		return a.Add(b)
	} else {
		return a
	}
}

func LocalTest(nodeNum, degree, dataScale, sampleScale, distributeParam int, p *big.Int, variation, mean decimal.Decimal) {
	//curve := edwards25519.NewBlakeSHA256Ed25519()
	//fmt.Print(curve.ScalarLen())
	param := NewParam(nodeNum, degree, dataScale, sampleScale, p)
	owner := NewOwner(param, false)
	consumer := NewConsumer(param, false)
	nodes := make([]*Node, nodeNum)
	for i := 0; i < nodeNum; i++ {
		nodes[i] = NewNode(param, i, distributeParam, false)
	}

	{
		data := make([]*big.Int, dataScale)
		for i := 0; i < dataScale; i++ {
			data[i] = utils.RandomNum(p)
		}
		err := owner.SetData(data)
		if err != nil {
			return
		}
	}

	R, rShares, rSquareShares := batchDDGOffline(nodeNum, degree, p)

	var ownerTime, consumerTime, nodeTime time.Duration = 0, 0, 0
	var start, end time.Time
	var ownerStart, ownerEnd time.Time
	var nodeStart, nodeEnd time.Time

	index := make([]*big.Int, nodeNum)
	for node := 0; node < nodeNum; node++ {
		index[node] = new(big.Int).SetInt64(int64(node + 1))
	}

	// protocol 1
	// as owner
	ownerStart = time.Now()
	cShares, vShares, r := owner.step1()

	nodeStart = time.Now()
	for i := 0; i < nodeNum; i++ {
		tmp1 := make([]*big.Int, dataScale)
		for j := 0; j < dataScale; j++ {
			tmp1[j] = cShares[j][i]
		}
		tmp2 := make([]*big.Int, nodeNum)
		for j := 0; j < nodeNum; j++ {
			tmp2[j] = vShares[j][i]
		}
		nodes[i].step1(tmp1, tmp2, r)
	}

	// step1, step2, step3 refactor from DGG_local_20241022.py
	BDDGStep1 := func() (*big.Int, []*big.Int) {
		// owner select random value on Z_2, run secret sharing on Z_p among nodes
		bo, boShares := owner.batchDDGInit()
		boAddRShares := make([]*big.Int, nodeNum)

		// node verify s^2-s
		for node := 0; node < nodeNum; node++ {
			boAddRShares[node] = new(big.Int).Add(boShares[node], rShares[node])
			nodes[node].bandwidth += len(boAddRShares[node].Bytes())
		}
		boAddR := reconstruct(index, boAddRShares, p)

		for i := 0; i < nodeNum; i++ {
			nodes[i].bandwidth += len(boAddRShares[i].Bytes())
		}

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

		for i := 0; i < nodeNum; i++ {
			nodes[i].bandwidth += len(mpcShares[i].Bytes())
		}

		if mpcResult.Int64() != 0 {
			log.Printf("Chaord - MPC error, %v\n", mpcResult)
		}
		return bo, boShares
	}
	// Po done
	bo, boShares := BDDGStep1()

	// protocol 2
	var flag []int
	var cHat [][]*big.Int

	for i := 0; i < nodeNum; i++ {
		cHat, flag = nodes[i].step2Sample(cShares)
	}
	bHat, pHat := owner.step2(flag)

	for i := 0; i < nodeNum; i++ {
		nodes[i].step2Forward(bHat, pHat)
	}

	consumer.bandwidthRecv += dataScale * len(bHat[0].Bytes())
	consumer.bandwidthRecv += dataScale * nodeNum * len(cHat[0][0].Bytes())
	start = time.Now()
	consumer.step2(index, cHat, bHat)
	end = time.Now()
	consumerTime += end.Sub(start)

	BDDGStep2 := func() []*big.Int {
		// consumer select random value on Z_p, run secret sharing on Z_p among nodes
		start = time.Now()
		bc, bcShares := consumer.batchDDGInit()
		end = time.Now()
		consumerTime += end.Sub(start)

		bcReconstruct := reconstruct(index, bcShares, p)

		for i := 0; i < nodeNum; i++ {
			nodes[i].bandwidth += len(bcShares[i].Bytes())
		}

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
		for i := 0; i < nodeNum; i++ {
			nodes[i].bandwidth += len(bcAddboShares[i].Bytes())
		}

		tmp := new(big.Int).Add(bc, bo)
		tmp.Mod(tmp, BigTwo)
		if bcAddbo.Cmp(tmp) != 0 {
			log.Printf("Chaord - bc_add_bo bit reconstruct error, get %v, want %v", bcAddbo, tmp)

		}
		return bcAddboShares
	}

	// protocol 3
	b := owner.step3()
	ownerEnd = time.Now()

	bcAddboShares := BDDGStep2()

	consumer.bandwidthRecv += dataScale * len(b[0].Bytes())

	start = time.Now()
	r1 := consumer.step3(b)
	end = time.Now()
	consumerTime += end.Sub(start)

	//log.Printf("consumer time after step3:%v\n", consumerTime.Milliseconds())
	// node verify root then vote in ABA
	if bytes.Compare(r, r1) != 0 {
		log.Printf("Chaord - root error, get %v, want %v\n", r, r1)
	}

	BDDGStep3 := func() []*big.Int {
		// coin is generated offline
		coin := make([]int, distributeParam)
		for i := 0; i < distributeParam; i++ {
			coin[i] = rand.Int() % 2
		}

		// get k unbias bit shares
		bUnbiasShares := make([][]*big.Int, distributeParam)
		for i := 0; i < distributeParam; i++ {
			bUnbiasShares[i] = make([]*big.Int, nodeNum)
			for node := 0; node < nodeNum; node++ {

				if coin[i] == 0 {
					bUnbiasShares[i][node] = bcAddboShares[node]
				} else {
					bUnbiasShares[i][node] = new(big.Int).Sub(BigOne, bcAddboShares[node])
				}
			}
		}
		gaussianNonNormalShares := make([]*big.Int, nodeNum)
		// add shares to get gaussian shares
		for i := 0; i < nodeNum; i++ {
			tmp := new(big.Int).SetInt64(0)
			for j := 0; j < distributeParam; j++ {
				tmp.Add(tmp, bUnbiasShares[j][i])
			}
			gaussianNonNormalShares[i] = tmp.Mod(tmp, p)
		}
		// reconstruct(index, gaussianNonNormalShares, p)

		dP := decimal.NewFromInt(int64(distributeParam))
		v1 := decimal.NewFromInt(4).Mul(variation).Div(dP)
		v1 = v1.Mul(v1)
		v2 := mean.Sub(v1.Mul(dP).Div(decimal.NewFromInt(2)))
		v2Shares := shareSecretBig(v2.BigInt(), nodeNum, degree, p)
		gaussianShares := make([]*big.Int, nodeNum)
		for i := 0; i < nodeNum; i++ {
			gaussianValue := decimal.NewFromBigInt(gaussianNonNormalShares[i], 0)
			v2Value := decimal.NewFromBigInt(v2Shares[i], 0)
			value := v1.Mul(mod_decimal_normal(gaussianValue.Sub(v2Value), decimal.NewFromBigInt(p, 0)))
			gaussianShares[i] = value.BigInt()
		}
		return gaussianShares
	}

	ABA1(nodeNum, degree, nodes)

	gaussianShares := BDDGStep3()
	cBarShares := make([][]*big.Int, dataScale)
	for i := 0; i < dataScale; i++ {
		cBarShares[i] = make([]*big.Int, nodeNum)
		for j := 0; j < nodeNum; j++ {
			cBarShares[i][j] = new(big.Int).Add(cShares[i][j], gaussianShares[j])
			cBarShares[i][j].Mod(cBarShares[i][j], p)
		}
	}
	ABAtx(nodeNum, degree, nodes)

	nodeEnd = time.Now()

	//consumer.bandwidthRecv += dataScale * nodeNum * len(cBarShares[0])

	start = time.Now()
	consumer.step4(index, cBarShares)
	end = time.Now()
	consumerTime += end.Sub(start)

	ownerTime = ownerEnd.Sub(ownerStart)
	nodeTime = nodeEnd.Sub(nodeStart)

	for i := 0; i < nodeNum; i++ {
		nodes[i].bandwidth += nodes[i].osvTX.GetBandwidth()
		nodes[i].bandwidth += nodes[i].osv1.GetBandwidth()
	}

	logFile, err := os.OpenFile("output.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		fmt.Println("open log file failed, err:", err)
		return
	}
	log.SetOutput(logFile)
	log.Printf("setting: node number-%v, threshold-%v, distribution parameter-%v, data scale-%v\n", nodeNum, degree, distributeParam, dataScale)
	log.Printf("owner time:%v ms\n", ownerTime.Milliseconds())
	log.Printf("owner Band:%v Byte\n", owner.bandwidth)
	log.Printf("node time:%v ms\n", nodeTime.Milliseconds())
	log.Printf("node Band:%v Byte\n", nodes[0].bandwidth)
	log.Printf("consumer time:%v ms\n", consumerTime.Milliseconds())

	consumerBand := consumer.bandwidthRecv
	if consumer.bandwidthRecv < consumer.bandwidthSend {
		consumerBand = consumer.bandwidthSend
	}

	log.Printf("consumer Band:%v Byte\n", consumerBand)
	log.Printf("==========================================")
}
