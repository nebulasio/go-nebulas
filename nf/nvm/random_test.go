package nvm

import (
	"crypto/md5"
	"encoding/binary"
	"io"
	"math/rand"
	"testing"
)

func TestRandom(t *testing.T) {
	m := md5.New()
	io.WriteString(m, "blockseed")
	io.WriteString(m, "txhash")

	seed := int64(binary.BigEndian.Uint64(m.Sum(nil)))

	r := rand.New(rand.NewSource(seed))
}
