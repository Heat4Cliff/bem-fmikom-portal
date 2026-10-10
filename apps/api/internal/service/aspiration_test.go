package service

import "testing"

func TestAspirationStatusTransitionsAreStrictlyForward(t *testing.T) {
	if !allowedAspirationTransition("PENDING", "READ") || !allowedAspirationTransition("READ", "RESOLVED") {
		t.Fatal("expected the documented forward transitions to be allowed")
	}
	for _, pair := range [][2]string{{"PENDING", "RESOLVED"}, {"RESOLVED", "READ"}, {"RESOLVED", "PENDING"}, {"READ", "PENDING"}, {"PENDING", "PENDING"}} {
		if allowedAspirationTransition(pair[0], pair[1]) {
			t.Fatalf("unexpected transition %s -> %s", pair[0], pair[1])
		}
	}
}
