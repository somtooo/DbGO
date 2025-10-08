# Implementing the Node Format
Since we are implementing a btree that can be written to disk we first have to decide what we want the node to look like. Many dbs like postgress/mySQL have their own format which have more details but for our usecase we can use something simple
// node format:
// | type | nkeys |  pointers  |   offsets  | key-value`s |unused
// |  2B  |   2B  | nkeys * 8B | nkeys * 2B | ...         |

// key-value format:
// | klen | vlen | key | val |
// |  2B  |  2B  | ... | ... |

We simplify things in x ways:
1. Reusing the same node format for both internal and leaf nodes; an internal node will have its type byte value set to 1 while a leaf node will have its type byte value set to 2. This means internal nodes will have 0 value for both offsets and key-value pairs while leaf nodes will have zero value for pointers.

2. In a classic btree you only need n-1 keys to store n children e.g if a node has keys ["p", "q"] and its parent has a range of [a, z) then the children of that node can be broken into [a, p), [p, q), [q, z) so three childs then when you visit the node at [a,p) the range of its children will depend on this range taking into account its keys too. hence we have a recursive decomposition thats determined starting from the range of the root. But in our case we store n children for n keys so the starting range of each node is inherited directly from the parent e.g if a node has keys ["p", "q"] and its parent has range starting from [a, z) the nodes range has a lowerbound "p" so [p, q) and [q, z) and you can see if we were to implement lookup we dont need p we just need to know if its less than q or greater than or equal to q hence the first key in the node is redundant since it comes from the parent but this helps simplify code when we do lookups or splits* and makes it easy to visualize or think about a particular node since its lowebound is always stored in the node.

3. We make the Node max size the same as a page size so its easy to write the node to disk so we dont need to manage fitting multiple nodes in a page. Less efficient but simplifies stuff for us. We can also just use a free list to track deleted nodes and reuse pages
