package util

import (
	"fmt"
	"math"
	"math/rand"
	"testing"
	"time"
)

func Uint64() uint64 {
	return uint64(rand.Uint32())<<32 + uint64(rand.Uint32())
}

func TestPartialPriorityTree(t *testing.T) {
	tree := InitPartialPriorityTree()

	for i := 0; i < 100; i++ {
		r := Uint64()
		tree.Insert(r)
	}

	min := tree.PopMin()
	fmt.Println("Min popped:", min)

	fmt.Println(tree.Str())

	fmt.Println("Popping 10")
	for i := 0; i < 99; i++ {
		tree.PopMin()
	}
	fmt.Println(tree.Str())

	for i := 0; i < 100; i++ {
		r := Uint64()
		if r > math.MaxUint64/2 {
			tree.Insert(r)
		}
	}

	min = tree.PopMin()
	fmt.Println("Min popped:", min)

	fmt.Println(tree.Str())
	fmt.Println("Tree size:", tree.NodeCount())

	fmt.Println("Popping 10 nearmax")
	for i := 0; i < 10; i++ {
		tree.PopNearMax()
	}
	fmt.Println(tree.Str())

	total := 1000
	begin := uint64(time.Now().UnixNano())
	for i := 0; i < total; i++ {
		r := Uint64()
		tree.Insert(r)
	}
	end := uint64(time.Now().UnixNano())

	duration := end - begin
	fmt.Printf("Insert %d in %d\n", total, duration)

	for j := 0; j < 10; j++ {
		min_dropped := uint64(math.MaxUint64)
		max_popped := uint64(0)
		begin = uint64(time.Now().UnixNano())
		for i := 0; i < total; i++ {
			for n := 0; n < 100; n++ {
				tree.Insert(Uint64())
			}
			for n := 0; n < 96; n++ {
				nm := tree.PopNearMax()
				if nm < min_dropped {
					min_dropped = nm
				}
			}
			for n := 0; n < 4; n++ {
				pm := tree.PopMin()
				if pm > max_popped {
					max_popped = pm
				}
			}
		}
		end = uint64(time.Now().UnixNano())
		duration = end - begin
		fmt.Printf("Iteration %d: %d in %d\n", j, total, duration)
		fmt.Println("Tree size:", tree.NodeCount(), tree.Size())
		fmt.Printf("Min %.1f, max %.1f\n", float32(max_popped)/float32(math.MaxUint64), float32(min_dropped)/float32(math.MaxUint64))
	}

}
