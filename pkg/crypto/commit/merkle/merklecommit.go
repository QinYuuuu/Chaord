package merkle

import (
	"bytes"
	"errors"
)

// true:hash is the left node
// false:hash is the right ndoe
type Witness struct {
	HashF [][]byte
	Left  []bool
}

type Root []byte

func (w *Witness) SetHash(hash [][]byte) {
	w.HashF = hash
}
func (w *Witness) SetPos(left []bool) {
	w.Left = left
}
func (w *Witness) Hash() [][]byte {
	return w.HashF
}
func (w *Witness) Pos() []bool { //true if
	return w.Left
}

// Commit is a function to get the hash of the root of the merkle tree.
//
// Returns:
// - The hash of the root of the tree as node
//
// Example:
//
//	tree := MerkleTree(["1", "2", "3"], hasher.SHA256Hasher)
//	fmt.Println(tree.GetCommitment())
func Commit(m *MerkleTree) Root {
	//m := NewMerkleTree(data, hasher.SHA256Hasher)
	return m.root.Hash()
}

func CreateWitness(m *MerkleTree, index int) (Witness, error) {
	witnesshash := make([][]byte, 0)
	witnesspos := make([]bool, 0)
	n := m.leafs[index]
	for n.Parent != nil {
		if n.Parent.Left == n { //n is left node
			witnesshash = append(witnesshash, n.Parent.Right.hash)
			witnesspos = append(witnesspos, false)
		} else { //n is right node
			witnesshash = append(witnesshash, n.Parent.Left.hash)
			witnesspos = append(witnesspos, true)
		}
		n = n.Parent
	}
	witness := Witness{
		HashF: witnesshash,
		Left:  witnesspos,
	}
	return witness, nil
}

func Verify(comm []byte, w Witness, content []byte, hasher func([]byte) []byte) (bool, error) {
	var contenthash []byte
	if len(w.HashF) != len(w.Left) {
		return false, errors.New("error: witness has wrong length")
	}
	for i := 0; i < len(w.HashF); i++ {
		contenthash = hasher(content)
		if w.Left[i] {
			content = append(w.HashF[i], contenthash...)
		} else {
			content = append(contenthash, w.HashF[i]...)
		}
	}
	return bytes.Equal(comm, hasher(content)), nil
}
