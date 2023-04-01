package utils

import (
	"fmt"
	"strings"
	"testing"
)

func TestCalculateHash(t *testing.T) {
	str := "hello world"
	hash := CalculateHash(strings.NewReader(str))
	fmt.Println(hash)
	if hash != "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9" {
		t.Errorf("CalculateHash() failed. Got %s, expected %s.", hash, "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9")
	}
}
