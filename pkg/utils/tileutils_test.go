package utils

import "testing"

func TestMakeTile(t *testing.T) {
	testTile := Tile{X: 1, Y: 2, Z: 3}
	verifyTile := MakeTile(testTile.X, testTile.Y, testTile.Z)
	if verifyTile != testTile {
		t.Error("MakeTile did not return the correct tile")
	}
}

func TestGetPath(t *testing.T) {
	testTile := Tile{X: 1, Y: 2, Z: 3}
	path := testTile.GetPath()

	if path != "3/1/2" {
		t.Errorf("Expected path to be %s, but got %s", "3/1/2", path)
	}
}

func TestPathToTile(t *testing.T) {
	testTile := Tile{X: 1, Y: 2, Z: 3}
	testBasepath := "/path/to/tile"
	path := testBasepath + "/" + testTile.GetPath()

	verifyTile, verifyBasepath := PathToTile(path)
	if verifyTile != testTile {
		t.Error("PathToTile did not return the correct tile")
	}
	if verifyBasepath != testBasepath {
		t.Errorf("Expected basepath to be %s, but got %s", testBasepath, verifyBasepath)
	}

}

func TestNewPoint(t *testing.T) {
	testPoint := Point{X: 1, Y: 2}
	verifyPoint, _ := NewPoint("1", "2")

	if verifyPoint != testPoint {
		t.Error("NewPoint did not return the correct point")
	}

	_, err := NewPoint("bob", "sammy")

	if err == nil {
		t.Error("NewPoint should have returned an error")
	}

}
