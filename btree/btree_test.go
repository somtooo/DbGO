package btree

import (
	"testing"
)

func TestBNode(t *testing.T) {
	// Create a new node with enough space for header and 3 key-value pairs
	node := make(BNode, BTREE_PAGE_SIZE)

	t.Run("Header Operations", func(t *testing.T) {
		node.setHeader(BNODE_LEAF, 3)
		if got := node.getNodeType(); got != BNODE_LEAF {
			t.Errorf("getNodeType() = %v, want %v", got, BNODE_LEAF)
		}
		if got := node.getNumOfKeys(); got != 3 {
			t.Errorf("getNumOfKeys() = %v, want %v", got, 3)
		}
	})

	t.Run("Pointer Operations", func(t *testing.T) {
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

	t.Run("Size Constraints", func(t *testing.T) {
		node1max := HEADER + 8 + 2 + 4 + BTREE_MAX_KEY_SIZE + BTREE_MAX_VAL_SIZE
		if node1max > BTREE_PAGE_SIZE {
			t.Error("Node with single key-value pair doesn't fit in page size")
		}
	})
}
