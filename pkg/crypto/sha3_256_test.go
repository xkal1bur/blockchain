/* sha3_256_test.go */
package crypto

import (
	"encoding/hex"
	"os"
	"testing"
	"golang.org/x/crypto/sha3"
)

// func TestSha3_256FromFile(t *testing.T) {
// 	filePath := "testfile.txt"
// 	expected := "71e6f5caf74b081cd1bd464755c91bcda4aaa0fa949aecf96d9198f092a9cbc4"

// 	data, err := os.ReadFile(filePath)
// 	if err != nil {
// 		t.Fatalf("Failed to read file: %v", err)
// 	}

// 	hash := Sha3_256(data)
// 	result := hex.EncodeToString(hash)

// 	if result != expected {
// 		t.Errorf("Expected %s but got %s", expected, result)
// 	}
// }


func TestSha3_256FromFileLib(t *testing.T) {
	filePath := "testfile.txt"
	expected := "71e6f5caf74b081cd1bd464755c91bcda4aaa0fa949aecf96d9198f092a9cbc4"

	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	hash := sha3.Sum256(data)
	result := hex.EncodeToString(hash[:])

	if result != expected {
		t.Errorf("Expected %s but got %s", expected, result)
	}
}
