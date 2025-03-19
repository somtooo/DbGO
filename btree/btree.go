package btree

import (
	"fmt"
	"runtime"
)

// node format:
// | type | nkeys |  pointers  |   offsets  | key-values
// |  2B  |   2B  | nkeys * 8B | nkeys * 2B | ...

// key-value format:
// | klen | vlen | key | val |
// |  2B  |  2B  | ... | ... |

const HEADER = 4

const BTREE_PAGE_SIZE = 4096
const BTREE_MAX_KEY_SIZE = 1000
const BTREE_MAX_VAL_SIZE = 3000

func init() {
	node1max := HEADER + 8 + 2 + 4 + BTREE_MAX_KEY_SIZE + BTREE_MAX_VAL_SIZE
	err := assert(node1max <= BTREE_PAGE_SIZE) // Suppress unused warnings
	if err != nil {
		panic("Node size Error")
	}
}

const (
	BNODE_NODE = 1 // internal nodes without values
	BNODE_LEAF = 2 // leaf nodes with values
)

type BNode []byte // can be dumped to the disk

// Implement the functions below to implement serialization for our node format

// From the byte array deserialize the node type
func (node BNode) getNodeType() uint16 {
	return 1
}

// From the byte array deserialize the number of keys
func (node BNode) getNumOfKeys() uint16 {
	return 1
}

// Set the Header using little endian encoding
func (node BNode) setHeader(nodeType uint16, numKeys uint16) {
}

// Given a key index get its corresponding pointer
func (node BNode) getPtr(idx uint16) uint64 {
	return 1
}

// Given a key index set its corresponding pointer
func (node BNode) setPtr(idx uint16, value uint64) {

}

// Given an idx read the offsets array for the key
func (node BNode) getOffset(idx uint16) uint16 {
	return 1
}

// Now that you can read the offset for a kv pair now get the starting position of a kv pair using the getOffset()
func (node BNode) getKvPos(idx uint16) uint16 {
	return 1
}

// Now get the actual key as a byte slice
func (node BNode) getKey(idx uint16) []byte {
	return []byte("1")
}

// Now get the actual value as a byte slice
func (node BNode) getVal(idx uint16) []byte {
	return []byte("1")
}

// Add KV pairs or pointers to the node. don't forget to update the offset
func nodeAppendKV(new BNode, idx uint16, ptr uint64, key []byte, val []byte) {

}

func (node BNode) nbytes() uint16 {
	return node.getKvPos(node.getNumOfKeys())
}

// Assert checks if the condition is true and returns an error with detailed information if not.
// This should only be used for truly impossible situations and invariant checking.
func assert(cond bool, msg ...interface{}) error {
	if !cond {
		// Get caller information
		_, file, line, _ := runtime.Caller(1)
		message := "assertion failed"
		if len(msg) > 0 {
			message = fmt.Sprint(msg...)
		}
		return fmt.Errorf("assertion failed at %s:%d: %s", file, line, message)
	}
	return nil
}
