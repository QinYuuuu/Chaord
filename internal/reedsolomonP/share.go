package reedsolomonP

import "math/big"

// A Share represents a piece of the FEC-encoded data.
// Both fields are required.
type Share struct {
	Number int
	Data   *big.Int
}

type byNumber []Share

func (b byNumber) Len() int               { return len(b) }
func (b byNumber) Less(i int, j int) bool { return b[i].Number < b[j].Number }
func (b byNumber) Swap(i int, j int)      { b[i], b[j] = b[j], b[i] }
