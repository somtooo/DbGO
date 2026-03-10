package btree

import (
	"bytes"
)

// node format:
// | type | nkeys |  pointers  |   offsets  | key-value`s |unused
// |  2B  |   2B  | nkeys * 8B | nkeys * 2B | ...         |

// key-value format:
// | klen | vlen | key | val |
// |  2B  |  2B  | ... | ... |

// Todo: Can contain sibling pointers
// get, new, del are used to simulate writing to in memory and wil also be used to write to actual disk
type BTree struct {
	root uint64              // root pointer (a non zero page number)
	get  func(uint64) []byte // read data from a page number
	new  func([]byte) uint64 // allocate a new page number with data
	del  func(uint64)        // deallocate a page number
}

// Todo: Implement the btree insert/ update with tests here are some starter functions finish up
func (tree *BTree) Insert(key []byte, val []byte) {

	assert(len(key) != 0)
	assert(len(key) <= BTREE_MAX_KEY_SIZE)
	assert(len(val) <= BTREE_MAX_VAL_SIZE)

	if tree.root == 0 {
		// create the first node
		root := BNode(make([]byte, BTREE_PAGE_SIZE))
		root.setHeader(BNODE_LEAF, 2)
		// a dummy key, this makes the tree cover the whole key space.
		// thus a lookup can always find a containing node.
		nodeAppendKV(root, 0, 0, nil, nil)
		nodeAppendKV(root, 1, 0, key, val)
		tree.root = tree.new(root)
		return
	}

	node := treeInsert(tree, tree.get(tree.root), key, val)
	nsplit, split := nodeSplit3(node)
	tree.del(tree.root)
	//grow the root
	if nsplit > 1 {
		newRoot := make(BNode, BTREE_PAGE_SIZE)
		newRoot.setHeader(BNODE_INTERNAL, uint16(nsplit))
		for i := uint16(0); i < uint16(nsplit); i++ {
			nodeAppendKV(newRoot, i, tree.new(split[i]), split[i].getKey(0), nil)
		}
		tree.root = tree.new(newRoot)
	} else {
		tree.root = tree.new(split[0])
	}
}

// insert a KV into a node, the result might be split.
// the caller is responsible for deallocating the input node
// and splitting and allocating result nodes.
func treeInsert(tree *BTree, node BNode, key []byte, val []byte) BNode {
	// the result node.
	// it's allowed to be bigger than 1 page and will be split if so
	new := BNode(make([]byte, 2*BTREE_PAGE_SIZE))

	// where to insert the key?
	idx := nodeLookupLE(node, key)
	// act depending on the node type
	switch node.getNodeType() {
	case BNODE_LEAF:
		// leaf, node.getKey(idx) <= key
		if bytes.Equal(key, node.getKey(idx)) {
			// found the key, update it.
			leafUpdate(new, node, idx, key, val)
		} else {
			// insert it after the position.
			leafInsert(new, node, idx+1, key, val)
		}
	case BNODE_INTERNAL:
		// internal node recursively call into tree insert and if theres a split handle it.
		kidPtr := node.getPtr(idx)
		newNode := treeInsert(tree, tree.get(kidPtr), key, val)
		nsplit, split := nodeSplit3(newNode)
		tree.del(kidPtr)
		new.setHeader(BNODE_INTERNAL, node.getNumOfKeys()+uint16(nsplit)-1)

		// This can be made faster
		for i := range idx {
			nodeAppendKV(new, i, node.getPtr(i), node.getKey(i), node.getVal(i))
		}
		for i, knode := range split[:nsplit] {
			nodeAppendKV(new, idx+uint16(i), tree.new(knode), knode.getKey(0), nil)
		}
		for i := uint16(0); i < node.getNumOfKeys()-(idx+1); i++ {
			nodeAppendKV(new, idx+uint16(nsplit)+i, node.getPtr(idx+1+i), node.getKey(idx+1+i), node.getVal(idx+1+i))
		}

	default:
		panic("bad node!")
	}
	return new
}

func leafInsert(new, node BNode, insertPos uint16, key []byte, val []byte) {
	new.setHeader(BNODE_LEAF, node.getNumOfKeys()+1)
	nodeAppendRange(node, new, 0, 0, insertPos)
	nodeAppendKV(new, insertPos, 0, key, val)
	nodeAppendRange(node, new, insertPos, insertPos+1, node.getNumOfKeys()-insertPos)
}

func leafUpdate(new, node BNode, idx uint16, key []byte, val []byte) {
	new.setHeader(BNODE_LEAF, node.getNumOfKeys())
	nodeAppendRange(node, new, 0, 0, idx)
	nodeAppendKV(new, idx, 0, key, val)
	nodeAppendRange(node, new, idx+1, idx+1, node.getNumOfKeys()-(idx+1))
}

// returns the first kid node whose range intersects the key. (kid[i] <= key)
func nodeLookupLE(node BNode, key []byte) (idx uint16) {
	nKeys := node.getNumOfKeys()
	var i uint16
	for i = 0; i < nKeys; i++ {
		cmp := bytes.Compare(key, node.getKey(i))
		if cmp == 0 {
			return i
		}

		if cmp < 0 {
			return i - 1
		}
	}

	return i - 1
}

// Split an oversized node into two. Remember big key in the middle means we dont know on which side it will end up.
func nodeSplit2(left BNode, right BNode, old BNode) {
	nleft := old.getNumOfKeys() / 2

	leftSize := HEADER + 8*nleft + 2*nleft + old.getOffset(nleft)

	for leftSize > uint16(BTREE_PAGE_SIZE) {
		nleft--
		leftSize = HEADER + 8*nleft + 2*nleft + old.getOffset(nleft)
	}

	rightSize := old.nbytes() - leftSize

	for rightSize > uint16(BTREE_PAGE_SIZE) {
		nleft++
		leftSize = HEADER + 8*nleft + 2*nleft + old.getOffset(nleft)
		rightSize = old.nbytes() - leftSize
	}

	left.setHeader(old.getNodeType(), nleft)
	nodeAppendRange(old, left, 0, 0, nleft)

	nright := old.getNumOfKeys() - nleft
	right.setHeader(old.getNodeType(), nright)
	nodeAppendRange(old, right, nleft, 0, nright)
}

func nodeAppendRange(old BNode, new BNode, oldStartIdx uint16, newStartIdx uint16, numOfKVPairs uint16) {
	assert(oldStartIdx+numOfKVPairs <= old.getNumOfKeys())
	assert(newStartIdx+numOfKVPairs <= new.getNumOfKeys())
	if numOfKVPairs == 0 {
		return
	}

	for i := uint16(0); i < numOfKVPairs; i++ {
		new.setPtr(newStartIdx+i, old.getPtr(i+oldStartIdx))
	}

	for i := uint16(1); i <= numOfKVPairs; i++ {
		offset := (old.getOffset(i+oldStartIdx) - old.getOffset(oldStartIdx-1+i)) + new.getOffset(i-1+newStartIdx)

		new.setOffset(i+newStartIdx, offset)
	}

	oldKvPosStart := old.getKvPos(oldStartIdx)
	oldKvPosEnd := old.getKvPos(oldStartIdx + numOfKVPairs)
	newKvPos := new.getKvPos(newStartIdx)

	copy(new[newKvPos:], old[oldKvPosStart:oldKvPosEnd])
}

// Can split a node into two or three. After splitting a node into two, the left half may still be too large, because while fitting the right half, the
// left half size grows. This can happen if there is a big key in the middle, requiring another split. So that's why the node can be split into three
func nodeSplit3(old BNode) (int, [3]BNode) {
	if old.nbytes() > uint16(BTREE_PAGE_SIZE) {
		left := make(BNode, 2*BTREE_PAGE_SIZE)
		right := make(BNode, BTREE_PAGE_SIZE)
		nodeSplit2(left, right, old)
		if left.nbytes() > uint16(BTREE_PAGE_SIZE) {
			newLeft := make(BNode, BTREE_PAGE_SIZE)
			middle := make(BNode, BTREE_PAGE_SIZE)
			nodeSplit2(newLeft, middle, left)
			return 3, [3]BNode{newLeft, middle, right}
		}

		return 2, [3]BNode{left[:BTREE_PAGE_SIZE], right}
	}

	return 1, [3]BNode{old[:BTREE_PAGE_SIZE]}

}
