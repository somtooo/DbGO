package bnode

import (
	"encoding/binary"
	"fmt"
	"runtime"
)

// node format:
// | type | nkeys |  pointers  |   offsets  | key-value`s |unused
// |  2B  |   2B  | nkeys * 8B | nkeys * 2B | ...         |

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

// the node type. 1 = internal node and 2 = leaf node
const (
	BNODE_INTERNAL = 1 // internal nodes without values
	BNODE_LEAF     = 2 // leaf nodes with values
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

// Given a key index get its corresponding pointer assume idx starts from zero
func (node BNode) getPtr(idx uint16) uint64 {
	startIndex := HEADER + 8*(idx)
	return binary.LittleEndian.Uint64(node[startIndex:])
}

// Given a key index set its corresponding pointer assume idx starts from zero
func (node BNode) setPtr(idx uint16, value uint64) {
	startIndex := HEADER + 8*(idx)
	binary.LittleEndian.PutUint64(node[startIndex:], value)
}

// The offset array stores the starting point of each key value pair. e.g offset(0) -> 0 offset(1) -> 8 (if the total bytes taken up by the first key value pair is 8), offset(2) -> 19 (if the total bytes taken up by the first + second key-value pair is 19 so 8b + 11b)
func (node BNode) getOffset(idx uint16) uint16 {
	if idx == 0 {
		return 0
	}

	startIndex := (HEADER + node.getNumOfKeys()*8) + (idx-1)*2
	return binary.LittleEndian.Uint16(node[startIndex:])
}

func (node BNode) setOffset(idx uint16, value uint16) {
	startIndex := (HEADER + node.getNumOfKeys()*8) + idx*2
	binary.LittleEndian.PutUint16(node[startIndex:], value)
}

// Now get the starting position of a kv pair using the getOffset()
func (node BNode) getKvPos(idx uint16) uint16 {
	// HEADER + POINTERS + OFFSETS
	numOfKeys := node.getNumOfKeys()
	return (HEADER + numOfKeys*8 + numOfKeys*2) + node.getOffset(idx)
}

// Now get the actual key as a byte slice because key can be of any comparable type
func (node BNode) getKey(idx uint16) []byte {
	var kvSize uint16 = 4
	kvStartPosition := node.getKvPos(idx)
	keyLen := binary.LittleEndian.Uint16(node[kvStartPosition:])
	startIndex := kvStartPosition + kvSize
	result := make([]byte, keyLen)
	copy(result, node[startIndex:startIndex+keyLen])
	return result
}

// Now get the actual value as a byte slice
func (node BNode) getVal(idx uint16) []byte {
	var kvSize uint16 = 4
	kvStartPosition := node.getKvPos(idx)
	keyLen := binary.LittleEndian.Uint16(node[kvStartPosition:])
	valLenStartPosition := kvStartPosition + 2
	valLen := binary.LittleEndian.Uint16(node[valLenStartPosition:])
	startIndex := kvStartPosition + kvSize + keyLen
	result := make([]byte, valLen)
	copy(result, node[startIndex:startIndex+valLen])
	return result

}

// // node format:
// | type | nkeys |  pointers  |   offsets  | key-value`s |unused
// |  2B  |   2B  | nkeys * 8B | nkeys * 2B | ...         |

// // key-value format:
// | klen | vlen | key | val |
// |  2B  |  2B  | ... | ... |

// Add KV pairs or pointers to the node. don't forget to update the offset
// Current implementation assumes idx starts from zero
func nodeAppendKV(new BNode, idx uint16, ptr uint64, key []byte, val []byte) {
	new.setPtr(idx, ptr)

	keyLen := uint16(len(key))
	valLen := uint16(len(val))

	startPosition := new.getKvPos(idx)
	binary.LittleEndian.PutUint16(new[startPosition:], keyLen)
	binary.LittleEndian.PutUint16(new[startPosition+2:], valLen)

	copy(new[startPosition+4:], key)
	copy(new[startPosition+4+keyLen:], val)

	new.setOffset(idx, new.getOffset(idx)+4+keyLen+valLen)
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
