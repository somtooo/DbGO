package btree

// get, new, del are used to simulate writing to in memory and wil also be used to write to actual disk
type Btree struct {
	root uint64              // root pointer (a non zero page number)
	get  func(uint64) []byte // read data from a page number
	new  func([]byte) uint64 // allocate a new page number with data
	del  func(uint64)        // deallocate a page number
}

// Implement the btree insert
func (tree *Btree) Insert(key []byte, val []byte) error {

	return nil
}
