package reedsolomonP

import (
	"errors"
	"fmt"
	"math/big"
	"sort"
)

// modAdd 在模p下执行加法
func modAdd(a, b, p *big.Int) *big.Int {
	result := new(big.Int).Add(a, b)
	result.Mod(result, p)
	return result
}

// modSub 在模p下执行减法
func modSub(a, b, p *big.Int) *big.Int {
	result := new(big.Int).Sub(a, b)
	result.Mod(result, p)
	return result

}

// modMul 在模p下执行乘法
func modMul(a, b, p *big.Int) *big.Int {
	result := new(big.Int).Mul(a, b)
	result.Mod(result, p)
	return result
}

// modInverse computes the modular multiplicative inverse of a under modulo p.
func modInverse(a, p *big.Int) (*big.Int, error) {
	inv := new(big.Int).ModInverse(a, p)
	if inv == nil {
		return nil, fmt.Errorf("no inverse exists for %s in Z_%s", a, p)
	}
	return inv, nil
}

// modPow computes a^b modPow p.
func modPow(a, b, p *big.Int) *big.Int {
	return new(big.Int).Exp(a, b, p)
}

func evalPoint(num int, p *big.Int) *big.Int {
	if num == 0 {
		return BigZero
	}
	return modPow(BigTwo, new(big.Int).SetInt64(int64(num-1)), p)
}

// invertMatrix 矩阵求逆
func invertMatrix(s, a [][]*big.Int, p *big.Int) error {
	n := len(s)
	for i := 0; i < n; i++ {
		pivot := s[i][i]
		if pivot.Sign() == 0 {
			return errors.New("matrix is singular")
		}
		pivotInv, err := modInverse(pivot, p)
		if err != nil {
			return err
		}
		for j := 0; j < n; j++ {
			s[i][j] = modMul(s[i][j], pivotInv, p)
			a[i][j] = modMul(a[i][j], pivotInv, p)
		}
		for j := 0; j < n; j++ {
			if i != j {
				factor := s[j][i]
				for k := 0; k < n; k++ {
					s[j][k] = modSub(s[j][k], modMul(factor, s[i][k], p), p)
					a[j][k] = modSub(a[j][k], modMul(factor, a[i][k], p), p)
				}
			}
		}
	}
	return nil
}

// dotProduct 计算两个向量的点积
func dotProduct(a, b []*big.Int, p *big.Int) *big.Int {
	result := big.NewInt(0)
	for i := range a {
		result.Add(result, modMul(a[i], b[i], p))
	}
	return result.Mod(result, p)
}

// evalPoly 评估多项式在某个点的值
func evalPoly(poly []*big.Int, x, p *big.Int) *big.Int {
	result := big.NewInt(0)
	for i := len(poly) - 1; i >= 0; i-- {
		result.Mul(result, x)
		result.Add(result, poly[i])
		result.Mod(result, p)
	}
	return result
}

// divPolynomials 多项式除法
func divPolynomials(A, B []*big.Int, p *big.Int) ([]*big.Int, []*big.Int, error) {
	n := len(A)
	m := len(B)
	if m == 0 {
		return nil, nil, errors.New("division by zero")
	}
	if n < m {
		return []*big.Int{big.NewInt(0)}, A, nil
	}

	Q := make([]*big.Int, n-m+1)
	R := make([]*big.Int, len(A))
	for i := range A {
		R[i] = new(big.Int).Set(A[i])
	}

	for i := n - m; i >= 0; i-- {
		q, err := modInverse(B[m-1], p)
		if err != nil {
			return nil, nil, err
		}
		q = modMul(R[i+m-1], q, p)
		Q[i] = new(big.Int).Set(q)
		for j := 0; j < m; j++ {
			R[i+j] = modSub(R[i+j], modMul(q, B[j], p), p)
		}
	}

	return Q, R, nil
}

// isZero 检查多项式是否为零
func isZero(poly []*big.Int) bool {
	for _, c := range poly {
		if c.Sign() != 0 {
			return false
		}
	}
	return true
}

// BerlekampWelch corrects errors in the data using the Berlekamp-Welch algorithm.
func (fc *RSGFp) BerlekampWelch(shares []Share, e int) ([]Share, error) {
	k := fc.k // required size

	q := e + k - 1 // def of Q polynomial
	fmt.Printf("e=%v, q=%v\n", e, q)
	if e <= 0 {
		return nil, errTooFewShards
	}
	dim := q + e + 2
	// build the system of equations s * u = f
	s := make([][]*big.Int, dim)
	a := make([][]*big.Int, dim)
	for i := range s {
		s[i] = make([]*big.Int, dim)
	}
	for i := range s {
		a[i] = make([]*big.Int, dim)
	}
	f := make([]*big.Int, dim)
	u := make([]*big.Int, dim)
	for i := range f {
		f[i] = big.NewInt(0)
	}
	f[dim-1] = big.NewInt(1)
	for i := 0; i < dim-1; i++ {
		x_i := new(big.Int).SetInt64(int64(shares[i].Number + 1))
		r_i := shares[i].Data
		for j := 0; j < q+1; j++ {
			s[i][j] = modPow(x_i, big.NewInt(int64(j)), fc.p)
			//s[i][j] = modSub(BigZero, s[i][j], fc.p)
			if i == j {
				a[i][j] = big.NewInt(1)
			} else {
				a[i][j] = big.NewInt(0)
			}
		}

		for l := 0; l < e+1; l++ {
			j := l + q + 1
			s[i][j] = modMul(modPow(x_i, big.NewInt(int64(l)), fc.p), r_i, fc.p)
			if i == j {
				a[i][j] = big.NewInt(1)
			} else {
				a[i][j] = big.NewInt(0)
			}
		}
	}
	for i := 0; i < dim; i++ {
		s[dim-1][i] = BigZero
		a[dim-1][i] = BigZero
	}
	s[dim-1][dim-1] = BigOne
	a[dim-1][dim-1] = BigOne
	// invert and put the result in a
	err := invertMatrix(s, a, fc.p)
	if err != nil {
		return nil, err
	}
	// multiply the inverted matrix by the column vector
	for i := 0; i < dim; i++ {
		ri := a[i]
		u[i] = dotProduct(ri, f, fc.p)
	}

	// reverse u for easier construction of the polynomials
	for i := 0; i < len(u)/2; i++ {
		o := len(u) - i - 1
		u[i], u[o] = u[o], u[i]
	}

	qPoly := u[e+1:]
	// E(x) is monic polynomial
	ePoly := u[:e+1]

	pPoly, rem, err := divPolynomials(qPoly, ePoly, fc.p)
	if err != nil {
		return nil, err
	}

	if !isZero(rem) {
		return nil, tooManyErrors
	}

	out := make([]Share, fc.n)
	for i := 0; i < fc.n; i++ {
		fecBuf := new(big.Int).SetInt64(0)
		for j := 0; j < fc.k; j++ {
			fecBuf = new(big.Int).Add(fecBuf, new(big.Int).Mul(pPoly[k-j-1], fc.encMatrix[i][j]))
			fecBuf.Mod(fecBuf, fc.p)
		}

		out[i] = Share{
			Number: i,
			Data:   fecBuf,
		}
	}
	return out, nil
}

// Correct corrects the errors in the shares using the Berlekamp-Welch algorithm.
func (fc *RSGFp) Correct(shares []Share) ([]Share, error) {
	k := fc.k
	r := len(shares)
	if len(shares) < k {
		return nil, errTooFewShards
	}

	// Sort the shares by their number
	sort.Sort(byNumber(shares))

	e := (r - k) / 2
	// Use Berlekamp-Welch algorithm to correct errors
	for i := 0; i <= e; i++ {
		correctedShares, err := fc.BerlekampWelch(shares, i)
		if err != nil {
			continue
		}
		//fmt.Printf("correctedData: %v\n", correctedShares)
		return correctedShares, nil
	}
	return nil, tooManyErrors
}

// Decode will take a list of shares and decode the original data.
func (fc *RSGFp) Decode(shares []Share) ([]*big.Int, error) {
	k := fc.k

	if len(shares) < k {
		return nil, errTooFewShards
	}

	// Correct any errors in the shares
	correctedShares, err := fc.Correct(shares)
	if err != nil {
		return nil, err
	}

	// Sort the shares by their number
	sort.Sort(byNumber(correctedShares))
	s := make([][]*big.Int, k)
	a := make([][]*big.Int, k)
	f := make([]*big.Int, k)
	output := make([]*big.Int, k)
	for i := 0; i < k; i++ {
		s[i] = fc.encMatrix[i][:k]
		a[i] = make([]*big.Int, k)
		f[i] = shares[i].Data
		for j := 0; j < k; j++ {
			if i == j {
				a[i][j] = big.NewInt(1)
			}
		}
	}
	err = invertMatrix(s, a, fc.p)
	if err != nil {
		return nil, err
	}
	for i := 0; i < k; i++ {
		ri := a[i]
		output[i] = dotProduct(ri, f, fc.p)
	}
	return output, nil
}
