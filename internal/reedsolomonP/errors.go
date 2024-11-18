package reedsolomonP

import "errors"

// errTooFewShards is returned if too few shards where given to
// Encode/Verify/Reconstruct/Update. It will also be returned from Reconstruct
// if there were too few shards to reconstruct the missing data.
var errTooFewShards = errors.New("too few shards given")

// errInvalidRowSize will be returned if attempting to create a matrix with negative or zero row number.
var errInvalidRowSize = errors.New("invalid row size")

// errInvalidColSize will be returned if attempting to create a matrix with negative or zero column number.
var errInvalidColSize = errors.New("invalid column size")

// errColSizeMismatch is returned if the size of matrix columns mismatch.
var errColSizeMismatch = errors.New("column size is not the same for all rows")

// errMatrixSize is returned if matrix dimensions are doesn't match.
var errMatrixSize = errors.New("matrix sizes do not match")

// errNotSquare is returned if matrix dimensions are doesn't match.
var errNotSquare = errors.New("matrix is not square")

var errSingular = errors.New("matrix is singular")

var tooManyErrors = errors.New("too many errors to reconstruct")
