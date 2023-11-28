package streams

import (
	"math/rand"
	"time"
)

var randomStringChoices = []rune("abcdefghijklmnopqrstuvwxyz0123456789-")

type RandomStringReader struct {
	gen       *rand.Rand
	buf       []rune
	maxLength int
	minLength int
}

func NewRandomStringReader(minLength, maxLength int) (*RandomStringReader, error) {
	rs := &RandomStringReader{
		gen:       rand.New(rand.NewSource(time.Now().UnixNano())),
		buf:       make([]rune, maxLength+1),
		maxLength: maxLength,
		minLength: minLength,
	}
	return rs, nil
}

func (rs *RandomStringReader) Read(p []byte) (n int, err error) {
	length := rs.gen.Intn(rs.maxLength-rs.minLength) + rs.minLength
	for i := 0; i < length; i += 1 {
		rs.buf[i] = randomStringChoices[rs.gen.Intn(len(randomStringChoices))]
	}
	rs.buf[length] = '\n'
	n = copy(p, []byte(string(rs.buf[0:length+1])))
	return n, nil
}

func (rs *RandomStringReader) Close() error {
	return nil
}

func (rs *RandomStringReader) Seek(offset int64, whence int) (int64, error) {
	return 0, nil
}
