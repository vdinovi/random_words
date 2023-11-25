package streams_test

import (
	"bufio"
	"regexp"
	"testing"

	streams "github.com/vdinovi/go/streams"
)

func TestShufWordsReader(t *testing.T) {
	var wordRegex = regexp.MustCompile(`^[A-Za-z\-]+$`)

	sf, err := streams.NewDefaultShufWordsReader()
	if err != nil {
		t.Fatalf("unexepected error: %s", err)

	}
	scanner := bufio.NewScanner(sf)

	n := 100
	for i := 0; i < n; i += 1 {
		if !scanner.Scan() {
			t.Fatalf("scanner ended early")
		}
		if err := scanner.Err(); err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		bytes := scanner.Bytes()
		if !wordRegex.Match(bytes) {
			t.Fatalf("expected %q to match %s but did not", string(bytes), wordRegex)
		}
	}

	if err := sf.Close(); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
}
