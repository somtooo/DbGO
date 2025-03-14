package btree

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
	assert(node1max <= BTREE_PAGE_SIZE)
}

const (
	BNODE_NODE = 1 // internal nodes without values
	BNODE_LEAF = 2 // leaf nodes with values
)

type BNode []byte // can be dumped to the disk

// Implement the functions below to implement serialization for our node format

// //////////////////// HEADER /////////////////////
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

//////////////////////// HEADER END ////////////////////

////////////////// CHILD POINTERS /////////////////////

// Given a key index get its corresponding pointer
func (node BNode) getPtr(idx uint16) uint64 {
	return 1
}

// Given a key index set its corresponding pointer
func (node BNode) setPtr(idx uint16, value uint64) {

}

////////////////// CHILD POINTERS END /////////////////////

func assert(cond bool) {
	if !cond {
		panic("assertion failure")
	}
}
