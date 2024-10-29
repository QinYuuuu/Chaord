package merkle

import (
	"errors"
)

// MerkleTree struct is used to represent a merkle tree and the hash function required for computation.
type MerkleTree struct {
	root   *Node // Root of the merkle tree
	leafs  []*Node
	hasher func([]byte) []byte // hash function used for generating the tree
}

// NewMerkleTree NewTree creates a new Merkle Tree using the content cs.
func NewMerkleTree(cs [][]byte, hasher func([]byte) []byte) (*MerkleTree, error) {
	t := &MerkleTree{
		hasher: hasher,
	}
	root, leafs, err := buildWithContent(cs, hasher)
	if err != nil {
		return nil, err
	}
	t.root = root
	t.leafs = leafs
	return t, nil
}

// buildWithContent is a helper function that for a given set of Contents, generates a
// corresponding tree and returns the root node, a list of leaf nodes, and a possible error.
// Returns an error if cs contains no Contents.
func buildWithContent(cs [][]byte, hasher func([]byte) []byte) (*Node, []*Node, error) {
	if len(cs) == 0 {
		return nil, nil, errors.New("error: cannot construct tree with no content")
	}
	var leafs []*Node
	for _, c := range cs {
		hash := hasher(c)
		leafs = append(leafs, &Node{
			hash: hash,
			data: c,
			leaf: true,
			dup:  false,
		})
	}
	if len(leafs)%2 == 1 {
		duplicate := &Node{
			hash: leafs[len(leafs)-1].hash,
			data: leafs[len(leafs)-1].data,
			leaf: true,
			dup:  true,
		}
		leafs = append(leafs, duplicate)
	}
	root, err := buildIntermediate(leafs, hasher)
	if err != nil {
		return nil, nil, err
	}
	return root, leafs, nil
}

// buildIntermediate is a helper function that for a given list of leaf nodes, constructs
// the intermediate and root levels of the tree. Returns the resulting root node of the tree.
func buildIntermediate(nl []*Node, hasher func([]byte) []byte) (*Node, error) {
	var nodes []*Node
	for i := 0; i < len(nl); i += 2 {
		var left, right int = i, i + 1
		if i+1 == len(nl) {
			right = i
		}
		pdata := append(nl[left].hash, nl[right].hash...)
		phash := hasher(pdata)
		p := &Node{
			Left:   nl[left],
			Right:  nl[right],
			Parent: nil,
			hash:   phash,
			data:   pdata,
			leaf:   false,
		}
		nodes = append(nodes, p)
		nl[left].Parent = p
		nl[right].Parent = p
		if len(nl) == 2 {
			return p, nil
		}
	}
	return buildIntermediate(nodes, hasher)
}

// Root is a getter function to get the root of the merkle tree.
//
// Returns:
// - The root of the tree as node
//
// Example:
//
//	tree := MerkleTree(["1", "2", "3"], hasher.SHA256Hasher)
//	fmt.Println(tree.Root())
func (m *MerkleTree) Root() *Node {
	return m.root
}
