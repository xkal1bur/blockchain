package crypto

import (
	"os"
	"testing"
	"fmt"
)

func TestSha256EduFromExistingFile(t *testing.T) {
	data, err := os.ReadFile("testfile.txt")
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	expected := "4a79aed64097a0cd9e87f1e88e9ad771ddb5c5d762b3c3bbf02adf3112d5d375"
	hash := Sha256Edu(data)
	result := fmt.Sprintf("%x", hash)

	if result != expected {
		t.Errorf("Expected %s but got %s", expected, result)
	}
}
