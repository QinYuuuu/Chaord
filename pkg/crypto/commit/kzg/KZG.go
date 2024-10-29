package kzg

import (
	"fmt"
	//"fmt"
	"github.com/Nik-U/pbc"
	"math/big"
	//"crypto/sha256"
)

func Setup(t int, pairing *pbc.Pairing, g *pbc.Element, n *big.Int) *PK {
	pk := new(PK)
	pk.pairing = pairing
	pk.g = make([]*pbc.Element, t+1)
	pk.g[0] = g
	pk.N = n
	pk.t = t
	s := pk.pairing.NewZr().Rand()
	for i := 1; i < t+1; i++ {
		pk.g[i] = pk.pairing.NewG1().PowZn(pk.g[i-1], s)
	}
	return pk
}

/*
	func Commit(pk *PK, poly1 []*pbc.Element)(*pbc.Element){
		return evaluatePolyPBC(pk, poly1)
	}
*/
func Commit(pk *PK, poly1 []*big.Int) *pbc.Element {
	return evaluatePolyBig(pk, poly1)
}

func VerifyPoly(pk *PK, c *pbc.Element, poly1 []*big.Int) bool {
	tmp := Commit(pk, poly1)
	//fmt.Printf("pk %v \n poly1 %v \n tmp %v \n c %v \n", pk.g, poly1, tmp, c)
	return c.Equals(tmp)
}

// Generate Lagrange Coefficient as PBC element
func GenerateLagrangeCoefficientPBC(pairing *pbc.Pairing, list []*pbc.Element, i int, x *pbc.Element) *pbc.Element {
	acc := pairing.NewZr().Set1()

	for j, _ := range list {
		if list[i].Equals(list[j]) {
			continue
		} else {
			tmp1 := pairing.NewZr().Sub(x, list[j])
			tmp2 := pairing.NewZr().Sub(list[i], list[j])
			acc.Mul(acc, tmp1)
			acc.Div(acc, tmp2)
		}
	}
	return acc
}

/*
// Generate Lagrange Coefficient as BigInt
func GenerateLagrangeCoefficientBigInt (list []*big.Int, i int, x *big.Int, q *big.Int) (*big.Int){
	acc, _ := new(big.Int).SetString("1", 10)

	for j, _ := range list {
		if list[i] == list[j] {
			continue
		} else{
			tmp1 := new(big.Int).Sub(x, list[j])
			tmp2 := new(big.Int).Sub(list[i], list[j])
			tmp2Inv
			acc.Mul(acc, tmp1)
			acc.Mul(acc, tmp2)
			acc.Mod(acc, q)
		}
	}
	return acc
}
*/
/*
func evaluatePolyPBC(pk *PK, poly []*pbc.Element) (*pbc.Element) {
	c := pk.pairing.NewG1().PowZn(pk.g[0], poly[0])
	for i := 1; i < len(poly); i++ {
		sp := pk.pairing.NewG1().PowZn(pk.g[i], poly[i])
		c = pk.pairing.NewG1().Mul(c, sp)
	}
	return c
}
*/
func evaluatePolyBig(pk *PK, poly []*big.Int) *pbc.Element {
	c := pk.pairing.NewG1().PowBig(pk.g[0], poly[0])
	for i := 1; i < len(poly); i++ {
		sp := pk.pairing.NewG1().PowBig(pk.g[i], poly[i])
		c = pk.pairing.NewG1().Mul(c, sp)
	}
	return c
}

// return f(i)
/*
func polynomial_getPBC(pairing *pbc.Pairing, poly1 []*pbc.Element, i *big.Int)(*pbc.Element){
	result := pairing.NewZr().Set0()
	bigi := pairing.NewZr().SetBig(i)
	power := pairing.NewZr().Set1()
	for j:=0;j<len(poly1);j++{
		tmp := pairing.NewZr().Mul(power, poly1[j])
		result.Add(result, tmp)
		power.Mul(power, bigi)
	}
	return result
}
*/
func polynomial_getBig(poly1 []*big.Int, i *big.Int, R *big.Int) *big.Int {
	result := new(big.Int).SetInt64(0)
	power := new(big.Int).SetInt64(1)
	for j := 0; j < len(poly1); j++ {
		tmp := new(big.Int).Mul(power, poly1[j])
		result.Add(result, tmp)
		power.Mul(power, i)
	}
	return new(big.Int).Mod(result, R)
}

/*
func CreateWitness(pk *PK, f []*pbc.Element, index *big.Int) (*pbc.Element, *pbc.Element){
	fi := polynomial_get(pk.pairing, f, index)
	fa := Commit(pk, f)
	gi := pk.pairing.NewG1().PowerBig(pk.g[0], index)
	gAminusI := pk.pairing.NewG1().Div(pk.g[1], gi)
	gAminusIinv := pk.pairing.NewG1().Invert(gAminusI)
	return fi, wi
}
*/
// CreateWitness
// y = p(z)
func CreateWitness(pk *PK, p []*big.Int, index, R *big.Int) (*pbc.Element, *big.Int, error) {
	y := polynomial_getBig(p, index, R)

	n := polynomialSub(p, []*big.Int{y}, R) // p-y
	// n := p // we can omit y (p(z))
	d := []*big.Int{fNeg(index, R), big.NewInt(1)} // x-z
	q, rem := polynomialDiv(n, d, R)
	fmt.Println("polydiv", q)
	if compareBigIntArray(rem, arrayOfZeroes(len(rem))) {
		return nil, nil,
			fmt.Errorf("remainder should be 0, instead is %d", rem)
	}

	// proof: e = [q(t)]₁
	w := evaluatePolyBig(pk, q)
	return w, y, nil
}

/*
	func VerifyEval(pk *PK, c *pbc.Element, index *big.Int, fi *pbc.Element, wi *pbc.Element)(bool){
		left := pk.pairing.NewGT().Pair(c, pk.g[0])
		tmp := pk.pairing.NewG1().Div(pk.g[1], pk.pairing.NewG1().PowBig(pk.g[0], index))
		right := pk.pairing.NewGT().Pair(wi, tmp)
		tmp = pk.pairing.NewGT().Pair(pk.g[0], pk.g[0])
		tmp = pk.pairing.NewGT().PowZn(tmp, fi)
		right.Mul(right, tmp)
		return right.Equals(left)
	}
*/
func VerifyEval(pk *PK, c *pbc.Element, index *big.Int, fi *big.Int, wi *pbc.Element) bool {
	left := pk.pairing.NewGT().Pair(c, pk.g[0])
	tmp := pk.pairing.NewG1().Div(pk.g[1], pk.pairing.NewG1().PowBig(pk.g[0], index))
	right := pk.pairing.NewGT().Pair(wi, tmp)
	tmp = pk.pairing.NewGT().Pair(pk.g[0], pk.g[0])
	tmp = pk.pairing.NewGT().PowBig(tmp, fi)
	right.Mul(right, tmp)
	return right.Equals(left)
}

func Batch0Witness(pk *PK, p []*big.Int, v []*big.Int) (*pbc.Element, error) {
	R := pk.N
	div := []*big.Int{big.NewInt(1)}
	for i := 0; i < len(v); i++ {
		d := []*big.Int{fNeg(v[i], R), big.NewInt(1)} // x-z
		div = polynomialMul(div, d, R)
	}
	q, rem := polynomialDiv(p, div, R)
	if compareBigIntArray(rem, arrayOfZeroes(len(rem))) {
		return nil,
			fmt.Errorf("remainder should be 0, instead is %d", rem)
	}
	w := evaluatePolyBig(pk, q)
	return w, nil
}

func Verify0Witness(pk *PK, c *pbc.Element, v []*big.Int, wi *pbc.Element) bool {
	pairing := pk.pairing
	left := pairing.NewGT().Pair(c, pk.g[0])
	n := []*big.Int{big.NewInt(1)}
	R := pk.N
	for i := 0; i < len(v); i++ {
		tmp := []*big.Int{fNeg(v[i], R), big.NewInt(1)} // x-v
		n = polynomialMul(n, tmp, R)
	}
	tmp := evaluatePolyBig(pk, n)
	right := pairing.NewGT().Pair(wi, tmp)
	return left.Equals(right)
}
