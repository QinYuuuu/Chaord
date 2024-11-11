package abvss

import (
	"Chaord/pkg/crypto/elgamal"
	"errors"
	"go.dedis.ch/kyber/v4"
	"log"
	"math/big"

	"Chaord/pkg/utils/polynomial"
)

func (vss *ABVSS) DistributorInit(pk []kyber.Point, s []*big.Int) error {
	if len(pk) != vss.nodeNum {
		return errors.New("node number mismatch PK number")
	}
	if len(s) != vss.batchSize {
		return errors.New("secret number mismatch batchSize")
	}
	vss.Distributor = &Distributor{
		pk:     pk,
		secret: s,
		polyf:  make([]polynomial.Polynomial, vss.batchSize),
		polyg:  make([]polynomial.Polynomial, vss.vNum),
	}
	return nil
}

func (vss *ABVSS) SamplePoly() error {
	if vss.Distributor == nil {
		return errors.New("not a distributor")
	}
	for i := 0; i < vss.batchSize; i++ {
		poly, err := polynomial.NewRand(vss.degree, vss.p)
		if err != nil {
			return err
		}
		err = poly.SetCoefficientBig(0, vss.secret[i])
		if err != nil {
			return err
		}
		vss.polyf[i] = poly
	}
	for i := 0; i < vss.vNum; i++ {
		poly, err := polynomial.NewRand(vss.degree, vss.p)
		if err != nil {
			return err
		}
		vss.polyg[i] = poly
	}
	return nil
}

func (vss *ABVSS) GenerateShares(index int) ([]kyber.Point, []kyber.Point, []kyber.Point, []kyber.Point, error) {
	if vss.Distributor == nil {
		return nil, nil, nil, nil, errors.New("not a distributor")
	}
	zix := make([]kyber.Point, vss.batchSize)
	ziy := make([]kyber.Point, vss.batchSize)
	xix := make([]kyber.Point, vss.vNum)
	xiy := make([]kyber.Point, vss.vNum)
	for i := 0; i < vss.batchSize; i++ {
		fi := vss.polyf[i].EvalMod(new(big.Int).SetInt64(int64(index+1)), vss.p)
		tmpx, tmpy, r := elgamal.Encrypt(vss.curve, vss.pk[index], fi.Bytes())
		if len(r) > 0 {
			log.Printf("remainder %s", r)
		}
		zix[i] = tmpx
		ziy[i] = tmpy
	}
	for i := 0; i < vss.vNum; i++ {
		gi := vss.polyg[i].EvalMod(new(big.Int).SetInt64(int64(index+1)), vss.p)
		tmpx, tmpy, r := elgamal.Encrypt(vss.curve, vss.pk[index], gi.Bytes())
		if len(r) > 0 {
			log.Printf("remainder %s", r)
		}
		xix[i] = tmpx
		xiy[i] = tmpy
	}
	return zix, ziy, xix, xiy, nil
}
