package main

import (
	"Chaord/internal/chaord"
	"github.com/shopspring/decimal"
	"math/big"
)

func main() {

	variation := decimal.NewFromFloat(4.0)
	mean := decimal.NewFromFloat(0.0)
	p, _ := new(big.Int).SetString("7237005577332262213973186563042994240857116359379907606001950938285454250989", 10)
	chaord.LocalTest(4, 1, 96, 48, 100, p, variation, mean)

}
