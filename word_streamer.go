// Package streams implements a simple word randomizer based on the
// linux `shuf` command and `/usr/share/dict/*` files.
// It is a personal utility of mine and not intended for serious use
package streams

import (
	"bufio"
	"io"
	"slices"
	"strings"
	"sync"
)

// A WordStreamer provides an interface for a consumer to recieve a stream of randomized words.
type WordStreamer struct {
	source    io.ReadCloser
	words     chan string
	errors    chan error
	close     chan chan error
	closeOnce sync.Once
	transform Transform
}

// NewWordStreamer returns a new WordStreamer
// transforms are applied to each word in the order specified
func NewWordStreamer(source io.ReadCloser, transforms ...Transform) (stream *WordStreamer, err error) {
	stream = &WordStreamer{
		source:    source,
		words:     make(chan string),
		errors:    make(chan error),
		close:     make(chan chan error),
		transform: identity,
	}
	slices.Reverse(transforms)
	var iter = &stream.transform
	for _, t := range transforms {
		*iter = (*iter).compose(t)
	}
	go stream.process()
	return stream, nil
}

// Close initiates closing exactly once in which case it returns
// an unbuffered receiver channel on which it sends any errors while closing.
// If already closed, then it returns nil. Remember to check for nil
// as selecting or ranging on a nil channel blocks indefinitely
func (stream *WordStreamer) Close() <-chan error {
	var errs chan error
	stream.closeOnce.Do(func() {
		if stream.close != nil {
			errs = make(chan error)
			stream.close <- errs
		}
	})
	return errs
}

// Returns an unbuffered receiver channel on which randomized words are sent
func (stream *WordStreamer) Words() <-chan string {
	return stream.words
}

// Returns an unbuffered receiver channel on which errors are sent
//
// If EOF is reached by the underlying reader without error, then an io.EOF
// is sent on this channel before the WordStreamer is closed
func (stream *WordStreamer) Errors() <-chan error {
	return stream.errors
}

func (stream *WordStreamer) process() {
	defer func() {
		close(stream.words)
		close(stream.errors)
		close(stream.close)
		stream.close = nil
	}()
	scanner := bufio.NewScanner(stream.source)
	var err error
	for scanner.Scan() {
		text := scanner.Text()
		if text, err = stream.transform(text); err != nil {
			stream.errors <- err
			continue
		}
		select {
		case errc := <-stream.close:
			if err := stream.source.Close(); err != nil {
				errc <- err
			}
			close(errc)
			return
		case stream.words <- text:
			continue
		}
	}
	if err := stream.source.Close(); err != nil {
		stream.errors <- err
	}
	if err := scanner.Err(); err != nil {
		stream.errors <- err
	} else {
		stream.errors <- io.EOF
	}
}

// A Transform function transforms a string to another string
// may be supplied to WordStreamer which then applies all specified
// transformers to each word
type Transform func(string) (string, error)

func identity(s string) (string, error) {
	return s, nil
}

func (t Transform) compose(other Transform) Transform {
	return func(s string) (r string, err error) {
		if r, err = other(s); err != nil {
			return "", err
		}
		return t(r)
	}
}

// Transforms string to uppercase
var ToUpperTransform = func(s string) (string, error) {
	return strings.ToUpper(s), nil
}

// Transforms string to lowercase
var ToLowerTransform = func(s string) (string, error) {
	return strings.ToUpper(s), nil
}
