package streams_test

import (
	"errors"
	"fmt"
	"io"
	"math/rand"
	"regexp"
	"testing"

	"github.com/vdinovi/go/streams"
)

var (
	lowercase = []rune("abcdefghijklmnopqrstuvwxyz")
	uppercase = []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ")
)

type mockShufWordsReader struct {
	readErr  error
	closeErr error
	eof      bool
	choices  []rune
}

func (sf *mockShufWordsReader) Read(p []byte) (n int, err error) {
	if sf.eof {
		return 0, io.EOF
	}
	if sf.readErr != nil {
		return 0, sf.readErr
	}
	buf := make([]rune, rand.Int()%(len(sf.choices)-1)+1)
	for i := range buf {
		buf[i] = sf.choices[rand.Int()%len(sf.choices)]
	}
	buf = append(buf, '\n')
	n = copy(p, []byte(string(buf)))
	return n, err
}

func (sf *mockShufWordsReader) Close() error {
	return sf.closeErr
}

func TestWordStreamer(t *testing.T) {
	sf := &mockShufWordsReader{choices: lowercase}
	stream, err := streams.NewWordStreamer(sf)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	n := 100
	wordCount := 0
	for i := 0; i < n; i += 1 {
		select {
		case err := <-stream.Errors():
			t.Errorf("unexpected error while reading: %s", err)
			i = n // break
		case <-stream.Words():
			wordCount += 1
		}
	}
	if closingErrs := stream.Close(); closingErrs != nil {
		for err := range closingErrs {
			t.Errorf("unexpected error while closing: %s", err)
		}
	}
	if wordCount != n {
		t.Errorf("expected to see %d words but only saw %d", n, wordCount)
	}
}

func TestWordStreamerClosesOnRequest(t *testing.T) {
	sf := &mockShufWordsReader{choices: lowercase}
	stream, err := streams.NewWordStreamer(sf)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if closingErrs := stream.Close(); closingErrs != nil {
		for err := range closingErrs {
			t.Errorf("unexpected error while closing: %s", err)
		}
	}
}

func TestWordStreamerClosesOnEOF(t *testing.T) {
	sf := &mockShufWordsReader{
		eof:     true,
		choices: lowercase,
	}

	stream, err := streams.NewWordStreamer(sf)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	select {
	case err := <-stream.Errors():
		if err != io.EOF {
			t.Errorf("unexpected error while reading: %s", err)
		}
	case word := <-stream.Words():
		t.Errorf("read word %q after eof", word)
	}

	if closingErrs := stream.Close(); closingErrs != nil {
		t.Error("expected close to return a nil channel but did not")
	}
}

func TestWordStreamerReadError(t *testing.T) {
	sf := &mockShufWordsReader{
		choices: lowercase,
		readErr: errors.New("read error"),
	}
	stream, err := streams.NewWordStreamer(sf)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	select {
	case err := <-stream.Errors():
		if err != sf.readErr {
			t.Errorf("unexpected error while reading: %s", err)
		}
	case word := <-stream.Words():
		t.Errorf("read word %q when expecting error", word)
	}

	if closingErrs := stream.Close(); closingErrs != nil {
		t.Error("expected close to return a nil channel but did not")
	}

}

func TestWordStreamerCloseError(t *testing.T) {
	sf := &mockShufWordsReader{
		choices:  lowercase,
		closeErr: errors.New("close error"),
	}
	// consumer close situation
	stream, err := streams.NewWordStreamer(sf)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if closingErrs := stream.Close(); closingErrs != nil {
		for err = range closingErrs {
			if err != sf.closeErr {
				t.Errorf("unexpected error while closing: %s", err)
			}
		}
	}
}

func TestWordStreamerWithTransforms(t *testing.T) {
	var regex = regexp.MustCompile(`^\"[A-Z]+\"$`)
	sf := &mockShufWordsReader{choices: lowercase}
	enquoteTransform := func(s string) (string, error) {
		return fmt.Sprintf("\"%s\"", s), nil
	}

	stream, err := streams.NewWordStreamer(sf, streams.ToUpperTransform, enquoteTransform)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	n := 100
	wordCount := 0
	for i := 0; i < n; i += 1 {
		select {
		case err := <-stream.Errors():
			t.Errorf("unexpected error while reading: %s", err)
			i = n // break
		case w := <-stream.Words():
			if !regex.Match([]byte(w)) {
				t.Fatalf("expected %q to match %s but did not", w, regex)
			}
			wordCount += 1
		}
	}
	if closingErrs := stream.Close(); closingErrs != nil {
		for err := range closingErrs {
			t.Errorf("unexpected error while closing: %s", err)
		}
	}
	if wordCount != n {
		t.Errorf("expected to see %d words but only saw %d", n, wordCount)
	}
}
