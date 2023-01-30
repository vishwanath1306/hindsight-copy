package util

import (
	"fmt"
	"math"
	"sort"
	"strings"
)

/*
A partial priority tree is used for priority ordering in Hindsight

It supports three operations:
Insert(int)
Remove(int)
PopMin(int)
PopNearMax(int)

PopMin always returns the min element
PopNearMax returns an element that is not the max element, but is large

Hindsight uses this because eviction only happens during overload, and it's
OK to evict things that are not going to be reported

TODO: convert to allow arbitrary elements and comparators. for now hard-coded to int elements

*/

// Invariants:
type TreeNode struct {
	expand   int
	collapse int
	low      uint64 // lower bound inclusive
	mid      uint64 // midpoint
	high     uint64 // upper bound inclusive
	elements []uint64
	sorted   bool
	isleaf   bool
	size     int
	left     *TreeNode
	right    *TreeNode
}

func InitPartialPriorityTree() *TreeNode {
	node := initTreeNode(0, math.MaxUint64)
	return node
}

func initTreeNode(low uint64, high uint64) *TreeNode {
	var n TreeNode
	n.expand = 20
	n.collapse = 6
	n.low = low
	n.mid = low + (high-low)/2 + 1
	n.high = high
	n.sorted = true
	n.isleaf = true
	n.size = 0
	return &n
}

func (n *TreeNode) Insert(elem uint64) {
	n.elements = append(n.elements, elem)
	n.sorted = false
	n.size += 1
}

func (n *TreeNode) Remove(elem uint64) {
	// hmm TODO
}

func (n *TreeNode) distributeElementsToLeaves() {
	for _, e := range n.elements {
		if e < n.mid {
			n.left.Insert(e)
		} else {
			n.right.Insert(e)
		}
	}
	n.elements = nil
	n.sorted = true
}

func (n *TreeNode) collapseIfEmpty() {
	if n.size == 0 {
		n.left = nil
		n.right = nil
		n.isleaf = true
		n.sorted = true
	}
}

// Invariants:
//    pre: size > 0
//    post: len(elements) < 10 && sorted = true
func (n *TreeNode) PopMin() uint64 {
	if n.isleaf {
		if n.size < 10 {
			/* Can sort and return an element; nothing fancy */
			if !n.sorted {
				sort.Slice(n.elements, func(i, j int) bool { return n.elements[i] < n.elements[j] })
				n.sorted = true
			}
			v := n.elements[0]
			n.elements = n.elements[1:]
			n.size -= 1
			return v
		}

		/* Expand the node */
		n.left = initTreeNode(n.low, n.mid-1)
		n.right = initTreeNode(n.mid, n.high)
		n.isleaf = false
	}

	/* Distribute elements to leaves */
	n.distributeElementsToLeaves()

	/* Pop from one of the children */
	var v uint64
	if n.left.Size() > 0 {
		v = n.left.PopMin()
	} else {
		v = n.right.PopMin()
	}
	n.size -= 1

	/* Drop children if they are both empty */
	n.collapseIfEmpty()

	return v
}

func (n *TreeNode) Size() int {
	return n.size
}

func (n *TreeNode) NodeCount() int {
	if n.isleaf {
		return 1
	} else {
		return n.left.NodeCount() + n.right.NodeCount() + 1
	}
}

func (n *TreeNode) PopNearMax() uint64 {
	if n.isleaf {
		/* Leaf nodes pick an arbitrary element */
		v := n.elements[0]
		n.elements = n.elements[1:]
		n.size -= 1
		return v
	}

	var v uint64

	/* Go right if possible */
	if n.right.Size() > 0 {
		v = n.right.PopNearMax()
	} else {

		n.distributeElementsToLeaves()

		if n.right.Size() > 0 {
			v = n.right.PopNearMax()
		} else {
			v = n.left.PopNearMax()
		}
	}

	n.size -= 1
	n.collapseIfEmpty()
	return v
}

func (n *TreeNode) Str() string {
	return n.str(0)
}

func (n *TreeNode) str(indent int) string {
	var b strings.Builder
	for i := 0; i < indent; i++ {
		fmt.Fprintf(&b, " ")
	}
	fmt.Fprintf(&b, "[%d, %d] %d elements (%d total)", n.low, n.high, len(n.elements), n.size)
	if n.isleaf {
		fmt.Fprintf(&b, "(leaf)")
	} else {
		fmt.Fprintf(&b, ":\n%s\n%s", n.left.str(indent+1), n.right.str(indent+1))
	}
	return b.String()
}
