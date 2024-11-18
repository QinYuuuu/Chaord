package reedsolomonP

import (
	"errors"
	"fmt"
	"math/big"
	"sort"
)

// RSGFp online-error correction algorithm in modulo matrixP Field
// k required pieces and n total pieces.
type RSGFp struct {
	k         int
	n         int
	encMatrix matrixP
	p         *big.Int
}

func NewRSGFp(k, n int, p *big.Int) (*RSGFp, error) {
	if k <= 0 || n <= 0 || k > n {
		return nil, errors.New("requires 1 <= k <= n <= 256")
	}

	encMatrix, err := VandermondeP(n, k, p)
	if err != nil {
		return nil, err
	}
	fmt.Println(encMatrix.String())
	return &RSGFp{
		k:         k,
		n:         n,
		encMatrix: encMatrix,
		p:         p,
	}, nil
}

// Encode will take input data and encode to the total number of pieces n this
// *FEC is configured for.
//
// The input data must be a multiple of the required number of pieces k.
// Padding to this multiple is up to the caller.
func (fc *RSGFp) Encode(input []*big.Int) ([]Share, error) {
	if len(input) < fc.k {
		return nil, errTooFewShards
	}
	output := make([]Share, fc.n)

	for i := 0; i < fc.n; i++ {
		fecBuf := new(big.Int).SetInt64(0)
		for j := 0; j < fc.k; j++ {
			fecBuf = new(big.Int).Add(fecBuf, new(big.Int).Mul(input[j], fc.encMatrix[i][j]))
			fecBuf.Mod(fecBuf, fc.p)
		}

		output[i] = Share{
			Number: i,
			Data:   fecBuf,
		}
	}
	return output, nil
}

// Rebuild will take a list of corrected shares (pieces) and a callback output.
// output will be called k times ((*FEC).Required() times) with 1/k of the
// original data each time and the index of that data piece.
// Decode is usually preferred.
//
// Note that the data is not necessarily sent to output ordered by the piece
// number.
//
// Note that the byte slices in Shares passed to output may be reused when
// output returns.
//
// Rebuild assumes that you have already called Correct or did not need to.
func (fc *RSGFp) Rebuild(shares []Share) ([]Share, error) {
	k := fc.k
	n := fc.n
	encMatrix := fc.encMatrix

	if len(shares) < k {
		return nil, errTooFewShards
	}

	sort.Sort(byNumber(shares))
	fmt.Println(shares)

	// Initialize the decoding matrix and vectors
	var mDec matrixP
	mDec = make([][]*big.Int, k)
	for i := range mDec {
		mDec[i] = make([]*big.Int, k)
	}
	indexes := make([]int, k)
	sharesv := make([]*big.Int, k)

	// Fill the decoding matrix and vectors
	for i := 0; i < k; i++ {
		share := shares[i]
		if share.Number >= n {
			return nil, fmt.Errorf("invalid share id: %d", share.Number)
		}

		if share.Number < k {
			mDec[i][share.Number] = BigOne
		} else {
			copy(mDec[i], encMatrix[share.Number][:k])
		}

		sharesv[i] = share.Data
		indexes[i] = share.Number
	}
	fmt.Println(mDec)

	invMDec, err := mDec.Invert(fc.p)
	if err != nil {
		return nil, err
	}
	fmt.Println(invMDec)
	fmt.Println(sharesv)
	// Solve the system of linear equations to find the original data
	originalData := make([]Share, n)
	for i := 0; i < k; i++ {
		if indexes[i] < k {
			originalData[indexes[i]] = Share{
				Number: indexes[i],
				Data:   sharesv[i],
			}
		} else {
			buf := big.NewInt(0)
			for j := 0; j < k; j++ {
				product := new(big.Int).Mul(sharesv[j], invMDec[i][j])
				buf.Add(buf, product)
				buf.Mod(buf, fc.p)
			}
			originalData[i] = Share{
				Number: i,
				Data:   buf,
			}

		}
	}
	return originalData, nil
}
