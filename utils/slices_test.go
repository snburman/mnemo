package utils

import "testing"

func TestReduceSlice(t *testing.T) {
	s := []int{1, 2, 3, 4, 5}
	f := func(a int, b int) int {
		return a + b
	}
	result := ReduceSlice(&s, f)
	if *result != 15 {
		t.Errorf("Expected 15, got %v", *result)
	}
}

func TestFilterSlice(t *testing.T) {
	s := []int{1, 2, 3, 4, 5}
	f := func(a int) bool {
		return a%2 == 0
	}
	result := FilterSlice(&s, f)
	if len(result) != 2 || result[0] != 2 || result[1] != 4 {
		t.Errorf("Expected [2, 4], got %v", result)
	}
}

func TestFind(t *testing.T) {
	s := []int{1, 2, 3, 4, 5}
	f := func(a int) bool {
		return a == 3
	}
	result := Find(&s, f)
	if result != 3 {
		t.Errorf("Expected 3, got %v", result)
	}
}
