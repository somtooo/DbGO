package btree

import (
	"bytes"
)

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
	if nsplit > 1 {

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
	// case BNODE_INTERNAL:
	// 	// internal node, insert it to a kid node.
	// 	nodeInsert(tree, new, node, idx, key, val)
	default:
		panic("bad node!")
	}
	return new
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

		if cmp > 0 {
			return i - 1
		}
	}

	return i - 1
}

// Split an oversized node into two
func nodeSplit2(left BNode, right BNode, old BNode) {
	numOfKeys := old.getNumOfKeys()
	midPoint := numOfKeys / 2

	leftBytes := func() uint16 {
		return (HEADER + midPoint*8 + midPoint*2) + old.getOffset(midPoint)
	}

	for leftBytes() > uint16(BTREE_PAGE_SIZE) {
		midPoint--
	}

	assert(midPoint >= 1)
	rightBytes := func() uint16 {
		return old.nbytes() - leftBytes() + HEADER
	}

	for rightBytes() > uint16(BTREE_PAGE_SIZE) {
		midPoint++
	}

	assert(midPoint < old.getNumOfKeys())

	left.setHeader(old.getNodeType(), midPoint)
	rightNumOfKeys := numOfKeys - midPoint
	right.setHeader(old.getNodeType(), rightNumOfKeys)

	for i := uint16(0); i < midPoint; i++ {
		nodeAppendKV(left, i, old.getPtr(i), old.getKey(i), old.getVal(i))
	}

	for i := uint16(0); i < rightNumOfKeys; i++ {
		nodeAppendKV(right, i, old.getPtr(midPoint+i), old.getKey(midPoint+i), old.getVal(midPoint+i))
	}
	assert(right.nbytes() <= uint16(BTREE_PAGE_SIZE))
}

// Can split a node into two or three. After splitting a node into two, the left half may still be too large, because while fitting the right half, the
// left half size grows. This can happen if there is a big key in the middle, requiring another split. So thats why the node can be split into three
func nodeSplit3(old BNode) (int, [3]BNode) {
	if old.nbytes() < uint16(BTREE_PAGE_SIZE) {
		old = old[:BTREE_PAGE_SIZE]
		return 1, [3]BNode{old}
	}

	left := make(BNode, 2*BTREE_PAGE_SIZE)
	right := make(BNode, BTREE_PAGE_SIZE)

	nodeSplit2(left, right, old)

	if left.nbytes() < uint16(BTREE_PAGE_SIZE) {
		left = left[:BTREE_PAGE_SIZE]
		return 2, [3]BNode{left, right}
	}

	left1 := make(BNode, BTREE_PAGE_SIZE)
	right1 := make(BNode, BTREE_PAGE_SIZE)

	nodeSplit2(left1, right1, left)
	return 3, [3]BNode{left1, right1, right}
}
