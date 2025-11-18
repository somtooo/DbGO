# Tree Insertion
Now that we've designed the node format and implemented the functions that enable working on a node we can now design and implement our Btree which is just a connection of nodes. So first problem to solve is how do we insert a key-value pair into a Btree?. First lets design our Btree format.

## Btree format
So usually a tree will have a root node and an insertion is just a recursion starting from the root all the way down. So we will have something like
```
type Btree struct {
	root BNode // root node
}
```
but remember this is not an in memory btree we would be writing this to disk which would fit in a page hence, we need a way to say hey get me the root node from disk which would require something like the pointer to the page so a better data structure for our use-case would be something like
```
type Btree struct {
	root uint64 // root node page pointer
}
```

now that we have our btree storing the pointer we will need a method that can get the pointer after writing a BNode to a page something like new(node BNode) uint64 would also be helpful to have a method that can retrieve a node from a page if we know the pointer so something like get(pointer uint64) BNode, also for delete delete(pointer uint64) error. For testing purposes we wont want this methods to actually be writing to and reading from pages, it will be too expensive so we would want the ability to modify the function implementation on the fly e.g in test the new function can just write to a map and generate a uint64 number rather than actually make an OS call. Given this requirements our on disk Btree data structure can look something like

```
type Btree struct {
	root uint64              // root pointer (a non zero page number)
	get  func(uint64) []byte // read data from a page number
	new  func([]byte) uint64 // allocate a new page number with data
	del  func(uint64)        // deallocate a page number
}
```
Then we will need methods like:
```
// To create our Btree
func (tree *Btree) new() Btree {
 return Btree {
 root: x1212121,
 get: // function that reads from disk pages or just code that simulates it hence keeping it in memory
 new: // function that allocates a new page with node data or just code that simulates it hence writing it to memory
 del: // function that deletes node from a page in memory or just code that simulates it.
 }
}

// To insert a key-value pair in our Btree
func (tree *Btree) Insert(key []byte, val []byte) error
```
### Implementation breakdown
