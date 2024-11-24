package chaord

import (
	"Chaord/pkg/utils/polynomial"
	"math/big"
	"math/rand"
)

func shareSecret(secret, n, t int, p *big.Int) []*big.Int {
	poly, _ := polynomial.New(t)
	poly.Rand(p)
	err := poly.SetCoefficient(0, int64(secret))
	if err != nil {
		return nil
	}
	shares := make([]*big.Int, n)
	for i := 0; i < n; i++ {
		shares[i] = poly.EvalMod(new(big.Int).SetInt64(int64(i+1)), p)
	}
	return shares
}

func shareSecretBig(secret *big.Int, n, t int, p *big.Int) []*big.Int {
	poly, _ := polynomial.New(t)
	poly.Rand(p)
	err := poly.SetCoefficientBig(0, secret)
	if err != nil {
		return nil
	}
	shares := make([]*big.Int, n)
	for i := 0; i < n; i++ {
		shares[i] = poly.EvalMod(new(big.Int).SetInt64(int64(i+1)), p)
	}
	return shares
}

func reconstruct(x, y []*big.Int, p *big.Int) *big.Int {
	poly, err := polynomial.LagrangeInterpolation(x, y, p)
	if err != nil {
		return nil
	}
	s, _ := poly.GetCoefficient(0)
	return s
}

func batchDDGOffline(nodeNum, degree int, p *big.Int) (*big.Int, []*big.Int, []*big.Int) {
	// generate R and R^2 share
	R := rand.Int() % 2
	RShares := shareSecret(R, nodeNum, degree, p)
	R2Shares := shareSecret(R*R, nodeNum, degree, p)
	return new(big.Int).SetInt64(int64(R)), RShares, R2Shares
}
