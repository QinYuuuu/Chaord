package reedsolomonP

import (
	"fmt"
	"math/big"
	"strings"
)

type matrixP [][]*big.Int

var BigZero = big.NewInt(0)
var BigOne = big.NewInt(1)
var BigTwo = big.NewInt(2)

// newMatrix returns a matrix of zeros.
func newMatrixP(rows, cols int) (matrixP, error) {
	if rows <= 0 {
		return nil, errInvalidRowSize
	}
	if cols <= 0 {
		return nil, errInvalidColSize
	}

	m := make([][]*big.Int, rows)
	for i := range m {
		m[i] = make([]*big.Int, cols)
		for j := range m[i] {
			m[i][j] = big.NewInt(0)
		}
	}
	return m, nil
}

// NewMatrixData initializes a matrix with the given row-major data.
// Note that data is not copied from input.
func newMatrixDataP(data [][]*big.Int) (matrixP, error) {
	m := matrixP(data)
	err := m.Check()
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (m matrixP) Check() error {
	rows := len(m)
	if rows == 0 {
		return errInvalidRowSize
	}
	cols := len(m[0])
	if cols == 0 {
		return errInvalidColSize
	}

	for _, col := range m {
		if len(col) != cols {
			return errColSizeMismatch
		}
	}
	return nil
}

// String returns a human-readable string of the matrix contents.
//
// Example: [[1, 2], [3, 4]]
func (m matrixP) String() string {
	rowOut := make([]string, 0, len(m))
	for _, row := range m {
		colOut := make([]string, 0, len(row))
		for _, col := range row {
			colOut = append(colOut, col.String())
		}
		rowOut = append(rowOut, "["+strings.Join(colOut, ", ")+"]")
	}
	return strings.Join(rowOut, ",\n")
}

// IdentityMatrix returns an identity matrix of the given size.
func identityMatrixP(size int) (matrixP, error) {
	m, err := newMatrixP(size, size)
	if err != nil {
		return nil, err
	}
	for i := range m {
		m[i][i] = new(big.Int).SetInt64(1)
	}
	return m, nil
}

// Multiply multiplies this matrix (the one on the left) by another
// matrix (the one on the right) and returns a new matrix with the result.
func (m matrixP) Multiply(right matrixP, p *big.Int) (matrixP, error) {
	if len(m[0]) != len(right) {
		return nil, fmt.Errorf("columns on left (%d) is different than rows on right (%d)", len(m[0]), len(right))
	}
	result, _ := newMatrixP(len(m), len(right[0]))
	for r, row := range result {
		for c := range row {
			value := new(big.Int).SetInt64(0)
			for i := range m[0] {
				tmp := new(big.Int).Mul(m[r][i], right[i][c])
				value = new(big.Int).Add(value, tmp)
			}
			result[r][c] = new(big.Int).Mod(value, p)
		}
	}
	return result, nil
}

// Augment returns the concatenation of this matrix and the matrix on the right.
func (m matrixP) Augment(right matrixP) (matrixP, error) {
	if len(m) != len(right) {
		return nil, errMatrixSize
	}

	result, _ := newMatrixP(len(m), len(m[0])+len(right[0]))
	for r, row := range m {
		for c := range row {
			result[r][c] = m[r][c]
		}
		cols := len(m[0])
		for c := range right[0] {
			result[r][cols+c] = right[r][c]
		}
	}
	return result, nil
}

func (m matrixP) SameSize(n matrixP) error {
	if len(m) != len(n) {
		return errMatrixSize
	}
	for i := range m {
		if len(m[i]) != len(n[i]) {
			return errMatrixSize
		}
	}
	return nil
}

// SubMatrix returns a part of this matrix. Data is copied.
func (m matrixP) SubMatrix(rmin, cmin, rmax, cmax int) (matrixP, error) {
	result, err := newMatrixP(rmax-rmin, cmax-cmin)
	if err != nil {
		return nil, err
	}
	// OPTME: If used heavily, use copy function to copy slice
	for r := rmin; r < rmax; r++ {
		for c := cmin; c < cmax; c++ {
			result[r-rmin][c-cmin] = m[r][c]
		}
	}
	return result, nil
}

// SwapRows Exchanges two rows in the matrix.
func (m matrixP) SwapRows(r1, r2 int) error {
	if r1 < 0 || len(m) <= r1 || r2 < 0 || len(m) <= r2 {
		return errInvalidRowSize
	}
	m[r2], m[r1] = m[r1], m[r2]
	return nil
}

// IsSquare will return true if the matrix is square, otherwise false.
func (m matrixP) IsSquare() bool {
	return len(m) == len(m[0])
}

// Invert returns the inverse of this matrix.
// Returns ErrSingular when the matrix is singular and doesn't have an inverse.
// The matrix must be square, otherwise ErrNotSquare is returned.
func (m matrixP) Invert(p *big.Int) (matrixP, error) {
	if !m.IsSquare() {
		return nil, errNotSquare
	}

	size := len(m)
	work, _ := identityMatrixP(size)

	work, _ = m.Augment(work)

	err := work.gaussianElimination(p)
	if err != nil {
		return nil, err
	}

	return work.SubMatrix(0, size, size, size*2)
}

// gaussianElimination performs Gaussian elimination on the matrix.
func (m matrixP) gaussianElimination(p *big.Int) error {
	n := len(m)
	for i := 0; i < n; i++ {
		// Find the pivot row
		pivotRow := i
		for j := i + 1; j < n; j++ {
			if (m)[j][i].Cmp(big.NewInt(0)) != 0 {
				pivotRow = j
				break
			}
		}

		if (m)[pivotRow][i].Cmp(big.NewInt(0)) == 0 {
			return errSingular
		}

		// Swap the current row with the pivot row
		(m)[i], (m)[pivotRow] = (m)[pivotRow], (m)[i]

		// Make the pivot element 1
		pivotInv, err := modInverse((m)[i][i], p)
		if err != nil {
			return err
		}
		for j := i; j < 2*n; j++ {
			(m)[i][j] = modMul((m)[i][j], pivotInv, p)
		}

		// Eliminate the other rows
		for k := 0; k < n; k++ {
			if k != i {
				factor := (m)[k][i]
				for j := i; j < 2*n; j++ {
					(m)[k][j] = modSub((m)[k][j], modMul((m)[i][j], factor, p), p)
				}
			}
		}
	}
	return nil
}

// VandermondeP creates a Vandermonde matrix, which is guaranteed to have the
// property that any subset of rows that forms a square matrix is invertible.
func VandermondeP(rows, cols int, p *big.Int) (matrixP, error) {
	result, err := newMatrixP(rows, cols)
	if err != nil {
		return nil, err
	}
	for r, row := range result {
		for c := range row {
			result[r][c] = new(big.Int).Exp(new(big.Int).SetInt64(int64(r+1)), new(big.Int).SetInt64(int64(c)), p)
		}
	}
	return result, nil
}
