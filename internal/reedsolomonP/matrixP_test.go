package reedsolomonP

import (
	"math/big"
	"testing"
)

// TestInvert_IdentityMatrix tests the inversion of an identity matrix.
func TestInvert_IdentityMatrix(t *testing.T) {
	p := big.NewInt(29)
	identity, _ := identityMatrixP(3)
	inverse, err := identity.Invert(p)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	expected, _ := identityMatrixP(3)
	for i := range expected {
		for j := range expected[i] {
			if inverse[i][j].Cmp(expected[i][j]) != 0 {
				t.Errorf("Expected inverse to be identity matrix, got: %v", inverse)
			}
		}
	}
}

// TestInvert_SmallMatrix tests the inversion of a small matrix.
func TestInvert_SmallMatrix(t *testing.T) {
	p := big.NewInt(29)
	matrix := matrixP{
		{big.NewInt(1), big.NewInt(2)},
		{big.NewInt(3), big.NewInt(4)},
	}
	inverse, err := matrix.Invert(p)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	expected := matrixP{
		{big.NewInt(11), big.NewInt(27)},
		{big.NewInt(21), big.NewInt(8)},
	}
	for i := range expected {
		for j := range expected[i] {
			if inverse[i][j].Cmp(expected[i][j]) != 0 {
				t.Errorf("Expected inverse to be %v, got: %v", expected, inverse)
			}
		}
	}
}

// TestInvert_SingularMatrix tests the inversion of a singular matrix.
func TestInvert_SingularMatrix(t *testing.T) {
	p := big.NewInt(29)
	matrix := matrixP{
		{big.NewInt(1), big.NewInt(2)},
		{big.NewInt(2), big.NewInt(4)},
	}
	_, err := matrix.Invert(p)
	if err == nil {
		t.Errorf("Expected error for singular matrix, got no error")
	}
}

// TestInvert_LargeMatrix tests the inversion of a larger matrix.
func TestInvert_LargeMatrix(t *testing.T) {
	p := big.NewInt(29)
	matrix := matrixP{
		{big.NewInt(1), big.NewInt(2), big.NewInt(3)},
		{big.NewInt(4), big.NewInt(5), big.NewInt(6)},
		{big.NewInt(7), big.NewInt(8), big.NewInt(10)},
	}
	inverse, err := matrix.Invert(p)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	expected := matrixP{
		{big.NewInt(19), big.NewInt(26), big.NewInt(2)},
		{big.NewInt(22), big.NewInt(11), big.NewInt(27)},
		{big.NewInt(2), big.NewInt(27), big.NewInt(21)},
	}
	for i := range expected {
		for j := range expected[i] {
			if inverse[i][j].Cmp(expected[i][j]) != 0 {
				t.Errorf("Expected inverse to be %v, got: %v", expected, inverse)
			}
		}
	}
}
