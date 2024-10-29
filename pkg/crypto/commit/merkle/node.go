package merkle

// Node represent a node in a Merkle Tree.
// It contains information including the key, hash of data, pointers to left child,
// right child and parent node, level of the node and index of the node in tree.
type Node struct {
	data   []byte // Data of the node
	hash   []byte // Hash of the node
	Left   *Node  // Left child of the node
	Right  *Node  // Right chiild of the node
	Parent *Node  // Parent of the node
	leaf   bool   // if a leaf node
	dup    bool   // if is duplicate leaf node
}

// Hash is a getter function to get the hash of the node.
//
// Returns:
// - The hash of the node as byte slice
//
// Example:
//
//	node := NewNode(byte("H"), nil, false)
//	hash := node.Hash()
//	fmt.Println(hash)
func (n *Node) Hash() []byte {
	return n.hash
}

// Key is a getter function to get the key of the node.
//
// Returns:
// - The key of the node as string
//
// Example:
//
//	node := NewNode("H", nil, false)
//	key := node.Key()
//	fmt.Println(key)
func (n *Node) Data() []byte {
	return n.data
}
