package streams_test

import (
	"bufio"
	"io"
	"regexp"
	"testing"

	streams "github.com/vdinovi/go/streams"
)

func TestRandomStringReader(t *testing.T) {
	var wordRegex = regexp.MustCompile(`^[a-z0-9\-]{3,8}$`)
	rs, err := streams.NewRandomStringReader(3, 8)
	if err != nil {
		t.Fatalf("unexepected error: %s", err)
	}
	scanner := bufio.NewScanner(rs)

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

	if _, err := rs.Seek(0, io.SeekStart); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if err := rs.Close(); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
}
