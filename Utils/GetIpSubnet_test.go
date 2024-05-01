package Utils

import (
	"testing"
)

func TestGetIpSubnetFromFile(t *testing.T) {
	expectedLength := 5
	file, err := GetIpSubnetFromFile("F:\\Golang\\HttpBenchmark\\all_cn_cidr.txt", expectedLength)
	if err != nil {
		t.Fatalf("error opening file: %v", err)
	}

	if len(file) != expectedLength {
		t.Errorf("Expected length of file to be %d, but got %d", expectedLength, len(file))
	}
}
