package main

import (
	"Chaord/internal/chaord"
	"math/big"
)

func main() {
	chaord.LocalTest(4, 1, 1, 1, 10, new(big.Int).SetInt64(17))
}
