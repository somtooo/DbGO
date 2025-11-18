package btree

import (
	"bytes"
	"fmt"
	"testing"
)

// buildNode builds a BNode with the given keys and vals. It sets the header
// to the final number of keys before appending, because nodeAppendKV expects
// the header to reflect the total key count.
func buildNode(keys [][]byte, vals [][]byte) BNode {
	numKeys := len(keys)
	// create a large backing buffer to avoid bounds issues when writing pointers,
	// offsets and kv data. Use several pages to be safe.
	buf := make(BNode, 4*BTREE_PAGE_SIZE)
	buf.setHeader(BNODE_LEAF, uint16(numKeys))
	for i := 0; i < numKeys; i++ {
		nodeAppendKV(buf, uint16(i), 0, keys[i], vals[i])
	}
	return buf
}

// collectParts extracts all keys and values from the returned parts and
// validates each part's nbytes() does not exceed BTREE_PAGE_SIZE.
func collectParts(t *testing.T, nsplit int, parts [3]BNode) ([][]byte, [][]byte) {
	var gotKeys [][]byte
	var gotVals [][]byte
	total := 0
	for i := 0; i < nsplit; i++ {
		part := parts[i]
		if part.nbytes() > uint16(BTREE_PAGE_SIZE) {
			t.Errorf("part %d too large: %d bytes", i, part.nbytes())
		}
		nk := int(part.getNumOfKeys())
		total += nk
		for j := 0; j < nk; j++ {
			gotKeys = append(gotKeys, part.getKey(uint16(j)))
			gotVals = append(gotVals, part.getVal(uint16(j)))
		}
	}
	// sanity: the caller should compare lengths with expected count
	_ = total
	return gotKeys, gotVals
}

// TestNodeSplit3_OrderAndSizes ensures nodeSplit3 preserves key/value order,
// preserves total count and that resulting parts fit into a page.
func TestNodeSplit3_OrderAndSizes(t *testing.T) {
	numKeys := 60
	keys := make([][]byte, numKeys)
	vals := make([][]byte, numKeys)
	for i := 0; i < numKeys; i++ {
		// Use moderately large keys so the node becomes oversized
		keys[i] = bytes.Repeat([]byte{byte('a' + (i % 26))}, 120)
		vals[i] = []byte(fmt.Sprintf("v%d", i))
	}

	old := buildNode(keys, vals)
	if old.nbytes() <= uint16(BTREE_PAGE_SIZE) {
		t.Fatalf("test setup failed: node not oversized, nbytes=%d", old.nbytes())
	}

	nsplit, parts := nodeSplit3(old)
	if nsplit < 2 || nsplit > 3 {
		t.Fatalf("unexpected nsplit %d, want 2 or 3", nsplit)
	}

	gotKeys, gotVals := collectParts(t, nsplit, parts)
	if len(gotKeys) != numKeys {
		t.Fatalf("total keys mismatch: got %d want %d", len(gotKeys), numKeys)
	}

	for i := 0; i < numKeys; i++ {
		if !bytes.Equal(gotKeys[i], keys[i]) {
			t.Fatalf("key %d mismatch", i)
		}
		if !bytes.Equal(gotVals[i], vals[i]) {
			t.Fatalf("val %d mismatch", i)
		}
	}
}

// TestNodeSplit3_LargeMiddleKey stresses splitting when a very large key is
// in the middle; the function may return 2 or 3 parts. We only assert the
// invariants (order, size, total count).
func TestNodeSplit3_LargeMiddleKey(t *testing.T) {
	numKeys := 120
	keys := make([][]byte, numKeys)
	vals := make([][]byte, numKeys)
	for i := 0; i < numKeys; i++ {
		if i == numKeys/2 {
			// Very large key to force additional splits
			keys[i] = bytes.Repeat([]byte{'Z'}, 2000)
		} else {
			keys[i] = bytes.Repeat([]byte{byte('k' + (i % 26))}, 40)
		}
		vals[i] = []byte(fmt.Sprintf("val-%d", i))
	}

	old := buildNode(keys, vals)
	if old.nbytes() <= uint16(BTREE_PAGE_SIZE) {
		t.Fatalf("test setup failed: node not oversized, nbytes=%d", old.nbytes())
	}

	nsplit, parts := nodeSplit3(old)
	if nsplit < 2 || nsplit > 3 {
		t.Fatalf("unexpected nsplit %d (want 2 or 3)", nsplit)
	}

	gotKeys, gotVals := collectParts(t, nsplit, parts)
	if len(gotKeys) != numKeys {
		t.Fatalf("total keys mismatch: got %d want %d", len(gotKeys), numKeys)
	}

	for i := 0; i < numKeys; i++ {
		if !bytes.Equal(gotKeys[i], keys[i]) {
			t.Fatalf("key %d mismatch", i)
		}
		if !bytes.Equal(gotVals[i], vals[i]) {
			t.Fatalf("val %d mismatch", i)
		}
	}
}
