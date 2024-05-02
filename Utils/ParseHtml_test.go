package Utils

import (
	"testing"
)

func TestGetAndParseLinks(t *testing.T) {
	got, err := GetAndParseLinks("https://pan.baidu.com/")
	if err != nil {
		t.Fatalf("GetAndParseLinks() error = %v", err)
	}
	t.Log("Links:")
	for i, link := range got {
		t.Logf("Link %d: %s", i, link)
	}
}
