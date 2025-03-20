package bnode

import (
	"encoding/binary"
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
	assert(node1max <= BTREE_PAGE_SIZE) // Suppress unused warnings
}

const (
	BNODE_NODE = 1 // internal nodes without values
	BNODE_LEAF = 2 // leaf nodes with values
)

type BNode []byte // can be dumped to the disk

// Implement the functions below to implement serialization for our node format

// From the byte array deserialize the node type
func (node BNode) getNodeType() uint16 {
	return binary.LittleEndian.Uint16(node[0:2])
}

// From the byte array deserialize the number of keys
func (node BNode) getNumOfKeys() uint16 {
	return binary.LittleEndian.Uint16(node[2:4])
}

// Set the Header using little endian encoding
func (node BNode) setHeader(nodeType uint16, numKeys uint16) {
	binary.LittleEndian.PutUint16(node[0:2], nodeType)
	binary.LittleEndian.PutUint16(node[2:4], numKeys)
}

// Given a key index get its corresponding pointer
func (node BNode) getPtr(idx uint16) uint64 {
	assert(idx < node.getNumOfKeys())
	position := HEADER + idx*8
	return binary.LittleEndian.Uint64(node[position:])
}

// Given a key index set its corresponding pointer
func (node BNode) setPtr(idx uint16, value uint64) {
	assert(idx < node.getNumOfKeys())
	position := HEADER + idx*8
	binary.LittleEndian.PutUint64(node[position:], value)
}

func (node BNode) setOffset(idx uint16, value uint16) {
	assert(idx < node.getNumOfKeys())
	position := HEADER + node.getNumOfKeys()*8 + idx*2
	binary.LittleEndian.PutUint16(node[position:], value)
}

// Given an idx read the offsets array for the key
func (node BNode) getOffset(idx uint16) uint16 {
	assert(idx < node.getNumOfKeys())
	if idx == 0 {
		return 0
	}
	position := HEADER + 8*node.getNumOfKeys() + (idx-1)*2
	return binary.LittleEndian.Uint16(node[position:])
}

// Now that you can read the offset for a kv pair now get the starting position of a kv pair using the getOffset()
func (node BNode) getKvPos(idx uint16) uint16 {
	assert(idx < node.getNumOfKeys())
	l1 := 8 * node.getNumOfKeys()
	l2 := 2 * node.getNumOfKeys()
	l3 := node.getOffset(idx)
	position := HEADER + l1 + l2 + l3
	return position
}

// Now get the actual key as a byte slice
func (node BNode) getKey(idx uint16) []byte {
	assert(idx < node.getNumOfKeys())
	klen := binary.LittleEndian.Uint16(node[node.getKvPos(idx):])
	position := node.getKvPos(idx) + 4
	return node[position : klen+position]
}

// Now get the actual value as a byte slice
func (node BNode) getVal(idx uint16) []byte {
	klen := binary.LittleEndian.Uint16(node[node.getKvPos(idx):])
	vlen := binary.LittleEndian.Uint16(node[node.getKvPos(idx)+2:])
	position := node.getKvPos(idx) + 4
	return node[position+klen : position+klen+vlen]
}

// Add KV pairs or pointers to the node. don't forget to update the offset
func nodeAppendKV(new BNode, idx uint16, ptr uint64, key []byte, val []byte) {
	new.setPtr(idx, ptr)
	position := new.getKvPos(idx)
	klen := uint16(len(key))
	vlen := uint16(len(val))
	//write klen
	binary.LittleEndian.PutUint16(new[position:position+2], klen)

	// write vlen
	binary.LittleEndian.PutUint16(new[position+2:position+4], vlen)

	// copy key
	copy(new[position+4:position+4+klen], key)

	// copy val
	copy(new[position+4+klen:position+4+klen+vlen], val)

	new.setOffset(idx, (4+klen+vlen)+new.getOffset(idx))
}

func (node BNode) nbytes() uint16 {
	return node.getKvPos(node.getNumOfKeys())
}

// Assert checks if the condition is true and returns an error with detailed information if not.
// This should only be used for truly impossible situations and invariant checking.
func assert(cond bool, msg ...interface{}) {
	if !cond {
		// Get caller information
		_, file, line, _ := runtime.Caller(1)
		message := "assertion failed"
		if len(msg) > 0 {
			message = fmt.Sprint(msg...)
		}
		err := fmt.Sprintf("assertion failed at %s:%d: %s", file, line, message)
		panic(err)
	}
}
