package generic

import (
	"log"
	"testing"

	"github.com/mushoffa/gorengan/generic"
)

func Test(t *testing.T) {
	nums := generic.NewSlice[int](0)

	nums2 := []int{4, 5, 6, 7, 8}

	log.Println(nums)

	nums.Add(1)
	log.Println(nums)

	nums.Append(nums2)
	log.Println(nums)

	log.Println("Size: ", nums.Size())
}
