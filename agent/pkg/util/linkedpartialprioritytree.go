package util

// import (
// 	"fmt"
// 	"math"
// 	"strings"
// )

// /*
// A partial priority tree is used for priority ordering in Hindsight

// It supports three operations:
// Insert(int)
// Remove(int)
// PopMin(int)
// PopNearMax(int)

// PopMin always returns the min element
// PopNearMax returns an element that is not the max element, but is large

// Hindsight uses this because eviction only happens during overload, and it's
// OK to evict things that are not going to be reported

// TODO: convert to allow arbitrary elements and comparators. for now hard-coded to int elements

// */

// // Invariants:
// type TreeNode struct {
// 	expand        int
// 	collapse      int
// 	low           uint64 // lower bound inclusive
// 	mid           uint64 // midpoint
// 	high          uint64 // upper bound inclusive
// 	head          *TreeElement
// 	tail          *TreeElement
// 	element_count int // The number of elements directly attached to this node
// 	sorted        bool
// 	isleaf        bool
// 	size          int // The size of this node plus all child nodes
// 	left          *TreeNode
// 	right         *TreeNode
// 	parent        *TreeNode
// }

// type TreeElement struct {
// 	valid bool
// 	value uint64
// 	node  *TreeNode
// 	next  *TreeElement
// 	prev  *TreeElement
// }

// func InitPartialPriorityTree() *TreeNode {
// 	node := initTreeNode(nil, 0, math.MaxUint64)
// 	return node
// }

// func initTreeNode(parent *TreeNode, low uint64, high uint64) *TreeNode {
// 	var n TreeNode
// 	n.expand = 20
// 	n.collapse = 6
// 	n.low = low
// 	n.mid = low + (high-low)/2 + 1
// 	n.high = high
// 	n.sorted = true
// 	n.isleaf = true
// 	n.size = 0
// 	n.parent = parent
// 	return &n
// }

// func (n *TreeNode) Insert(value uint64) *TreeElement {
// 	var e TreeElement
// 	e.value = value
// 	e.valid = true
// 	n.attach(&e)
// 	return &e
// }

// func (n *TreeNode) attach(e *TreeElement) {
// 	e.node = n
// 	e.next = nil
// 	e.prev = n.tail
// 	n.tail = e
// 	if e.prev == nil {
// 		n.head = e
// 	} else {
// 		e.prev.next = e
// 	}
// 	n.sorted = false
// 	n.element_count += 1
// 	n.size += 1
// }

// func (e *TreeElement) detach() {
// 	n := e.node

// 	if e.prev != nil {
// 		e.prev.next = e.next
// 	} else {
// 		n.head = e.next
// 	}
// 	if e.next != nil {
// 		e.next.prev = e.prev
// 	} else {
// 		n.tail = e.prev
// 	}

// 	n.element_count -= 1
// 	for n != nil {
// 		n.size -= 1
// 		n = n.parent
// 	}

// 	e.valid = false
// 	e.next = nil
// 	e.prev = nil
// 	e.node = nil
// }

// func (n *TreeNode) Remove(e *TreeElement) {
// 	if !e.valid {
// 		return
// 	}

// 	e.detach()
// }

// func (n *TreeNode) distributeElementsToLeaves() {
// 	current := n.head
// 	for current != nil {
// 		next := current.next
// 		if current.value < n.mid {
// 			n.left.attach(current)
// 		} else {
// 			n.right.attach(current)
// 		}
// 		current = next
// 	}
// 	n.head = nil
// 	n.tail = nil
// 	n.element_count = 0
// 	n.sorted = true
// }

// func (n *TreeNode) collapseIfEmpty() {
// 	if n.size == 0 {
// 		n.left = nil
// 		n.right = nil
// 		n.isleaf = true
// 		n.sorted = true
// 	}
// }

// func (n *TreeNode) sortElements() {
// 	if n.sorted {
// 		return
// 	}
// 	// TODO)
// 	// // TODO: sort linked list
// 	// sort.Slice(n.elements, func(i, j int) bool { return n.elements[i] < n.elements[j] })
// 	// n.sorted = true

// 	n.sorted = true
// }

// // Invariants:
// //    pre: size > 0
// //    post: len(elements) < 10 && sorted = true
// func (n *TreeNode) PopMin() uint64 {
// 	if n.isleaf {
// 		if n.size < 10 {
// 			/* Can sort and return an element; nothing fancy */
// 			// For now instead of sorting, just search for the min

// 			// n.sortElements()
// 			// e := n.head
// 			// n.head = e.next
// 			// if n.head == nil {
// 			// 	n.tail = nil
// 			// } else {
// 			// 	n.head.prev = nil
// 			// }
// 			// n.size -= 1
// 			// n.element_count -= 1

// 			// // Might be external references to e; make sure to null pointers
// 			// e.next = nil
// 			// e.prev = nil
// 			// e.node = nil
// 			// e.valid = false

// 			min := n.head
// 			next := min.next
// 			for next != nil {
// 				if next.value < min.value {
// 					min = next
// 				}
// 				next = next.next
// 			}

// 			min.detach()

// 			return min.value
// 		}

// 		/* Expand the node */
// 		n.left = initTreeNode(n, n.low, n.mid-1)
// 		n.right = initTreeNode(n, n.mid, n.high)
// 		n.isleaf = false
// 	}

// 	/* Distribute elements to leaves */
// 	n.distributeElementsToLeaves()

// 	/* Pop from one of the children */
// 	var v uint64
// 	if n.left.Size() > 0 {
// 		v = n.left.PopMin()
// 	} else {
// 		v = n.right.PopMin()
// 	}

// 	/* Drop children if they are both empty */
// 	n.collapseIfEmpty()

// 	return v
// }

// func (n *TreeNode) Size() int {
// 	return n.size
// }

// func (n *TreeNode) NodeCount() int {
// 	if n.isleaf {
// 		return 1
// 	} else {
// 		return n.left.NodeCount() + n.right.NodeCount() + 1
// 	}
// }

// func (n *TreeNode) PopNearMax() uint64 {
// 	if n.isleaf {
// 		/* Leaf nodes pick an arbitrary element */
// 		e := n.head
// 		e.detach()
// 		return e.value
// 	}

// 	var v uint64

// 	/* Go right if possible */
// 	if n.right.Size() > 0 {
// 		v = n.right.PopNearMax()
// 	} else {

// 		n.distributeElementsToLeaves()

// 		if n.right.Size() > 0 {
// 			v = n.right.PopNearMax()
// 		} else {
// 			v = n.left.PopNearMax()
// 		}
// 	}

// 	// Size is updated by detached node
// 	n.collapseIfEmpty()
// 	return v
// }

// func (n *TreeNode) Str() string {
// 	return n.str(0)
// }

// func (n *TreeNode) str(indent int) string {
// 	var b strings.Builder
// 	for i := 0; i < indent; i++ {
// 		fmt.Fprintf(&b, " ")
// 	}
// 	fmt.Fprintf(&b, "[%d, %d] %d elements (%d total)", n.low, n.high, n.element_count, n.size)
// 	if n.isleaf {
// 		fmt.Fprintf(&b, "(leaf)")
// 	} else {
// 		fmt.Fprintf(&b, ":\n%s\n%s", n.left.str(indent+1), n.right.str(indent+1))
// 	}
// 	return b.String()
// }
