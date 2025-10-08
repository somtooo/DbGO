package btree

import (
	"encoding/binary"
	"testing"
)

func TestGetNodeType(t *testing.T) {
	t.Run("Reads internal node type", func(t *testing.T) {
		node := make(BNode, BTREE_PAGE_SIZE)
		binary.LittleEndian.PutUint16(node, 1)
		if got := node.getNodeType(); got != BNODE_INTERNAL {
			t.Errorf("getNodeType() = %v, want %v", got, BNODE_INTERNAL)
		}
	})

	t.Run("Reads leaf node type", func(t *testing.T) {
		node := make(BNode, BTREE_PAGE_SIZE)
		binary.LittleEndian.PutUint16(node, 2)
		if got := node.getNodeType(); got != BNODE_LEAF {
			t.Errorf("getNodeType() = %v, want %v", got, BNODE_LEAF)
		}
	})

	t.Run("Pointer Operations", func(t *testing.T) {
		node := make(BNode, BTREE_PAGE_SIZE)
		node.setHeader(BNODE_INTERNAL, 3)
		testPtr := uint64(0x1234567890ABCDEF)
		node.setPtr(0, testPtr)
		if got := node.getPtr(0); got != testPtr {
			t.Errorf("getPtr(0) = %v, want %v", got, testPtr)
		}

		// Test multiple pointers
		pointers := []uint64{0x1111111111111111, 0x2222222222222222, 0x3333333333333333}
		for i, want := range pointers {
			node.setPtr(uint16(i), want)
			if got := node.getPtr(uint16(i)); got != want {
				t.Errorf("getPtr(%d) = %v, want %v", i, got, want)
			}
		}
	})

	t.Run("getOffset()", func(t *testing.T) {
		node := make(BNode, BTREE_PAGE_SIZE)
		node.setHeader(BNODE_LEAF, 2)
		node.setOffset(0, 8)
		node.setOffset(1, 16)

		if got := node.getOffset(0); got != 0 {
			t.Errorf("getOffset(0) =%v, want %v", got, 0)
		}

		if got := node.getOffset(1); got != 8 {
			t.Errorf("getOffset(1) =%v, want %v", got, 8)
		}
	})

	t.Run("nodeAppendKv", func(t *testing.T) {
		node := make(BNode, BTREE_PAGE_SIZE)
		node.setHeader(BNODE_LEAF, 3)
		nodeAppendKV(node, 0, 0, []byte("k"), []byte("v"))
		nodeAppendKV(node, 1, 0, []byte("q1"), []byte("tq2"))
		nodeAppendKV(node, 2, 0, []byte("t11"), []byte("v3"))

		if got := node.getKey(0); string(got) != "k" {
			t.Errorf("getKey(0) =%v want %v", got, "k")
		}

		if got := node.getKey(1); string(got) != "q1" {
			t.Errorf("getKey(1) =%v want %v", got, "q1")
		}

		if got := node.getKey(2); string(got) != "t11" {
			t.Errorf("getKey(2) =%v want %v", got, "t11")
		}

		node2 := make(BNode, BTREE_PAGE_SIZE)
		node2.setHeader(BNODE_LEAF, node.getNumOfKeys()+1)
		nodeAppendKV(node2, 0, 0, node.getKey(0), node.getVal(0))
		nodeAppendKV(node2, 1, 0, node.getKey(1), node.getVal(1))
		nodeAppendKV(node2, 2, 0, node.getKey(2), node.getVal(2))
		nodeAppendKV(node2, 3, 0, []byte("a123"), []byte("v4"))

		if got := node2.getKey(3); string(got) != "a123" {
			t.Errorf("getKey(3) =%v want %v", got, "a123")
		}

		if got := node.getVal(1); string(got) != "tq2" {
			t.Errorf("getVal(1) =%v want %v", got, "tq2")
		}
	})
}
