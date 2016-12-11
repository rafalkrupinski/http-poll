package uints64

import "testing"

func TestMax(t *testing.T) {
	max := Max(1, 0)
	if max != 1 {
		t.Error("Expected 1 got", max)
	}

}

func TestMax2(t *testing.T) {
	max := Max(0, 1)
	if max != 1 {
		t.Error("Expected 1 got", max)
	}

}
