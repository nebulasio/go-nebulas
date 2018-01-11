package stack

// Stack is a basic LIFO stack that resizes as needed.
type Stack struct {
	entries []interface{}
	limit   int
}

// NewStack returns a new stack.
// if stack is full, entries in stack will be kicked out using FIFO
func NewStack(size int) *Stack {
	return &Stack{limit: size}
}

// Len return size of stack
func (s *Stack) Len() int {
	return len(s.entries)
}

// Push adds a node to the stack.
func (s *Stack) Push(entry interface{}) {
	if s.Len() == s.limit {
		s.entries = s.entries[1:]
	}
	s.entries = append(s.entries, entry)
}

// Pop removes and returns a node from the stack in last to first order.
func (s *Stack) Pop() interface{} {
	if s.Len() == 0 {
		return nil
	}
	tail := s.Len() - 1
	entry := s.entries[tail]
	s.entries = s.entries[:tail]
	return entry
}
