package pdeq

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPdeq(t *testing.T) {
	tests := []struct {
		name string
		val  int
	}{
		{"31", 31},
		{"46", 46},
		{"51", 51},
		{"10", 10},
		{"30", 30},
		{"21", 21},
		{"71", 71},
		{"41", 41},
		{"11", 11},
		{"13", 13},
		{"16", 16},
		{"8", 8},
	}
	q := NewPdeq(func(a interface{}, b interface{}) bool { return a.(int) < b.(int) })
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q.Insert(tt.val)
		})
	}
	assert.Equal(t, q.PopMax(), 71)
	assert.Equal(t, q.PopMin(), 8)
	assert.Equal(t, q.PopMin(), 10)
	assert.Equal(t, q.PopMax(), 51)
	assert.Equal(t, q.PopMin(), 11)
	assert.Equal(t, q.PopMin(), 13)
	assert.Equal(t, q.PopMin(), 16)
	assert.Equal(t, q.PopMax(), 46)
	assert.Equal(t, q.PopMax(), 41)
	assert.Equal(t, q.PopMax(), 31)
	assert.Equal(t, q.PopMax(), 30)
}
