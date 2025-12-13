package main

import "testing"

func TestAlwaysTrue(t *testing.T) {
	if 1 != 1 {
		t.Error("expected 1 == 1")
	}
}
