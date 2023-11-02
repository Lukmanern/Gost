package experiment

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

func ImplementExample() {
	permissionIDs := []uint8{
		1, 8, 10, 20, 24, 30, 40, 50, 60, 70, 80, 180,
	}
	// 81 -> 11, 1, 0b_0010
	// func
	// input : id-int
	// output : group, bit-Nya

	// checkgroup
	for _, id := range permissionIDs {
		fmt.Println(id, "- group :", (id-1)/8+1)
		fmt.Printf("%08b\n", uint8(math.Pow(2, float64(id%8)-1)))
	}
}

func ImplementExample2() {
	var i1 uint8 = 0b00000001
	var i2 uint8 = 0b00000011
	fmt.Println(i1 | i2)
	fmt.Println(3 | 5)
}

func PermissionBitGroup(i int) int {
	return (i-1)/8 + 1
}

func Test_PermissionBitGroup(t *testing.T) {
	d := 8
	testCases := []struct {
		input  int
		result map[int]int
	}{
		{
			input: d,
			result: map[int]int{
				1: int(math.Pow(2, 7)),
			},
		},
		{
			input: 10 * d,
			result: map[int]int{
				10: int(math.Pow(2, 7)),
			},
		},
		{
			input: d + 7,
			result: map[int]int{
				2: int(math.Pow(2, 6)),
			},
		},
		{
			input: d,
			result: map[int]int{
				1: int(math.Pow(2, 7)),
			},
		},
	}

	for _, tc := range testCases {
		result := BuildBitGroups(tc.input)
		if !reflect.DeepEqual(result, tc.result) {
			t.Error("should same, but got", result, "want", tc.result)
		}
	}

	permIDs := make([]int, 0)
	for i := 1; i < 90; i++ {
		if i%2 != 0 {
			continue
		}
		permIDs = append(permIDs, i)
	}

	result := BuildBitGroups(permIDs...)
	for group, bits := range result {
		fmt.Printf("%d : %08b\n", group, bits)
	}
}

func Test_CheckHasPermission(t *testing.T) {
	// user perms
	permIDs := make([]int, 0)
	for i := 1; i <= 19; i++ {
		permIDs = append(permIDs, i)
	}

	bitGroups := BuildBitGroups(permIDs...)
	for i := 1; i <= 30; i++ {
		fmt.Println(i, ":", CheckHasPermission(i, bitGroups))
	}
}
