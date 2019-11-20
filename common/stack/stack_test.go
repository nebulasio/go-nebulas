package stack

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStack_Push(t *testing.T) {
	stack := NewStack(3)
	assert.Nil(t, stack.Pop())
	stack.Push(1)
	stack.Push(2)
	stack.Push(3)
	stack.Push(4)
	assert.Equal(t, stack.Pop(), 4)
	assert.Equal(t, stack.Pop(), 3)
	assert.Equal(t, stack.Pop(), 2)
	assert.Nil(t, stack.Pop())
}
