package sets_test

import (
	"testing"

	"github.com/tiggoins/ingress-validator/sets"
	"github.com/tiggoins/ingress-validator/client"
)

func TestIngressHost(t *testing.T) {
	ih := sets.NewIngressHost()

	// Test Add
	ih.Add("www.example.com")
	if !ih.Has("www.example.com") {
		t.Error("Add method failed: host not found")
	}

	// Test Delete
	ih.Delete("www.example.com")
	if ih.Has("www.example.com") {
		t.Error("Delete method failed: host still exists")
	}

	// Test Has
	ih.Add("www.example.com")
	if !ih.Has("www.example.com") {
		t.Error("Has method failed: host not found")
	}


	ih.Add(client.EmptyHost)
	ih.Add("www.example.com")
	ih.Add("www.example.com")
	ih.Add("www.example.com")

	// Test Len
	expectedLen := 2
	if ih.Len() != expectedLen {
		t.Errorf("Len method failed: expected %d, got %d", expectedLen, ih.Len())
	}

	// Test List (verify output manually)
	ih.List()
}
