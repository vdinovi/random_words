package streams

import (
	"fmt"
	"io"
	"os"
	"os/exec"
)

// A ShufWordsReader is an io.ReadCloser+io.Seeker which wraps an exec.Cmd and it's stdout stream
// for use as a source in a WordStreamer. This has been seperated out of WordStream
// primarily for testing purposes.
type ShufWordsReader struct {
	cmd    *exec.Cmd
	stdout io.ReadCloser
}

const (
	defaultShuf  = "shuf"
	defaultWords = "/usr/share/dict/words"
)

// Returns a new ShufWordsReader with the default `shuf` command and words file
func NewDefaultShufWordsReader() (sf *ShufWordsReader, err error) {
	return NewShufWordsReader(defaultShuf, defaultWords)
}

// Returns a new ShufWordsReader with the specified `shuf` command and words file
func NewShufWordsReader(shuf, words string) (sf *ShufWordsReader, err error) {
	shuf, words, err = makeArgs(shuf, words)
	if err != nil {
		return nil, err
	}
	cmd, stdout, err := makeCmd(shuf, words)
	if err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	sf = &ShufWordsReader{
		cmd:    cmd,
		stdout: stdout,
	}
	return sf, nil
}

// Reads from the stdout of the cmd
func (sf *ShufWordsReader) Read(p []byte) (n int, err error) {
	return sf.stdout.Read(p)
}

// Kills the cmd and closes it's stdout
func (sf *ShufWordsReader) Close() error {
	if err := sf.stdout.Close(); err != nil {
		return fmt.Errorf("error closing ShufWordsReader stdout: %s", err)
	}
	if err := sf.cmd.Process.Kill(); err != nil {
		return fmt.Errorf("error killing ShufWordsReader: %s", err)
	}
	sf.cmd.Wait()
	return nil
}

func (sf *ShufWordsReader) Seek(offset int64, whence int) (int64, error) {
	if whence != io.SeekStart || offset != 0 {
		return 0, fmt.Errorf("ShufWordsReader only supports seeking to SeekStart + 0")
	}
	if err := sf.Close(); err != nil {
		return 0, err
	}
	if err := sf.reset(); err != nil {
		return 0, err
	}
	return 0, nil
}

func (sf *ShufWordsReader) reset() (err error) {
	cmd := exec.Command(sf.cmd.Path, sf.cmd.Args[1:]...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	sf.cmd = cmd
	sf.stdout = stdout
	return nil
}

func makeArgs(shuf, words string) (string, string, error) {
	var err error
	if shuf, err = exec.LookPath(shuf); err != nil {
		return "", "", err
	}
	if _, err := os.Stat(words); err != nil {
		return "", "", err
	}
	return shuf, words, nil
}

func makeCmd(shuf, words string) (*exec.Cmd, io.ReadCloser, error) {
	cmd := exec.Command(shuf, words)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, nil, err
	}
	return cmd, stdout, nil
}
