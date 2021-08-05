package utils

import "testing"

func TestMin(t *testing.T) {
	if Min(1, 2) != 1 {
		t.Error("Min(1, 2) != 1")
	}
	if Min(2, 1) != 1 {
		t.Error("Min(2, 1) != 1")
	}
	if Min(-1, 1) != -1 {
		t.Error("Min(-1, 1) != 1")
	}
	if Min(1, -1) != -1 {
		t.Error("Min(1, -1) != 1")
	}
	if Min(1, 1) != 1 {
		t.Error("Min(1, 1) != 1 ?")
	}
}

func TestMax(t *testing.T) {
	if Max(1, 2) != 2 {
		t.Error("Max(1, 2) != 2")
	}
	if Max(2, 1) != 2 {
		t.Error("Max(2, 1) != 1")
	}
	if Max(-1, 1) != 1 {
		t.Error("Max(-1, 1) != -1")
	}
	if Max(1, -1) != 1 {
		t.Error("Max(1, -1) != -1")
	}
	if Max(1, 1) != 1 {
		t.Error("Max(1, 1) != 1 ?")
	}
}

func TestIntRange(t *testing.T) {
	testIntRange := IntRange(-5, 5)
	verifyIntRange := [11]int{-5, -4, -3, -2, -1, 0, 1, 2, 3, 4, 5}
	for i, v := range verifyIntRange {
		if testIntRange[i] != v {
			t.Errorf("IntRange(%d, %d)[%d] != %d", -5, 5, i, v)
		}
	}


	testIntRange2 := IntRange(5, -5)
	verifyIntRange2 := [11]int{5, 4, 3, 2, 1, 0, -1, -2, -3, -4, -5}
	for i, v := range testIntRange2 {
		if v != verifyIntRange2[i] {
			t.Errorf("IntRange(%d, %d)[%d] != %d", 5, -5, i, v)
		}
	}
}

func TestAbsInt(t *testing.T) {
	if AbsInt(1) != 1 {
		t.Error("AbsInt(1) != 1")
	}
	if AbsInt(-1) != 1 {
		t.Error("AbsInt(-1) != 1")
	}
	if AbsInt(0) != 0 {
		t.Error("AbsInt(0) != 0")
	}
}
