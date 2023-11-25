package streams_test

import (
	"fmt"
	"io"
	"os"

	streams "github.com/vdinovi/go/streams"
)

// Basic usage of streams.WordStreamer
func ExampleWordStreamer() {
	// Create a new ShufWordsReader command
	sf, err := streams.NewDefaultShufWordsReader()
	if err != nil {
		panic(err)
	}
	// Create a new WordStreamer with an uppercase transformer
	stream, err := streams.NewWordStreamer(sf, streams.ToUpperTransform)
	if err != nil {
		panic(err)
	}
	// Print out 100 words from the word stream
	func(numWords int) {
		for i := 0; i < numWords; i += 1 {
			select {
			// Listen for errors
			case err := <-stream.Errors():
				if err != io.EOF {
					fmt.Fprintf(os.Stderr, "error: %s\n", err)
				}
				return
			// Listen for words
			case word := <-stream.Words():
				fmt.Fprintf(os.Stdout, "word: %s\n", word)
			}
		}
	}(100)
	// Close and listen for errors
	if closingErrs := stream.Close(); closingErrs != nil {
		// If the returned channel is nil, it means this has already been closed
		for err := range closingErrs {
			fmt.Fprintf(os.Stderr, "error: %s\n", err)
		}
	}
}
