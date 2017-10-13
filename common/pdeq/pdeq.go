package pdeq

import (
	"math"
)

// Less is compare function for elements in queue
type Less func(a interface{}, b interface{}) bool

// Pdeq is a priority double-ended queue
type Pdeq struct {
	less Less
	heap []interface{}
}

// LvTy is type of level
type LvTy int

// LvMin & LvMax in priority double-ended queue
const (
	LvMin LvTy = iota
	LvMax
)

// NewPdeq create a new Pdeq
func NewPdeq(less Less) *Pdeq {
	return &Pdeq{
		less: less,
	}
}

// Len return the length of the priority deque
func (q *Pdeq) Len() int {
	return len(q.heap)
}

// Insert an item into priority deque
func (q *Pdeq) Insert(ele interface{}) {
	q.heap = append(q.heap, ele)
	q.bubbleUp(q.Len() - 1)
}

// PopMax pop the max value in priority deque
func (q *Pdeq) PopMax() interface{} {
	heap := q.heap
	pos := 0
	switch q.Len() {
	case 0:
		return nil
	case 1:
		break
	case 2:
		pos = 1
		break
	default:
		pos = 1
		if q.less(heap[1], heap[2]) {
			pos = 2
		}
		break
	}
	tx := heap[pos]
	q.deleteAt(pos)
	return tx
}

// PopMin pop the min value in priority deque
func (q *Pdeq) PopMin() interface{} {
	if q.Len() > 0 {
		tx := q.heap[0]
		q.deleteAt(0)
		return tx
	}
	return nil
}

func (q *Pdeq) deleteAt(pos int) {
	heap := q.heap
	size := len(heap)
	heap[pos] = heap[size-1]
	q.heap = heap[0 : size-1]
	q.trickleDown(pos)
}

func level(pos int) LvTy {
	level := (int)(math.Floor(math.Log2((float64)(pos + 1))))
	if level%2 == 0 {
		return LvMin
	}
	return LvMax
}

func parent(pos int) int {
	return (pos - 1) / 2
}

func leftChildren(pos int) int {
	return pos*2 + 1
}

func rightChildren(pos int) int {
	return pos*2 + 2
}

func (q *Pdeq) swap(i int, j int) {
	q.heap[i], q.heap[j] = q.heap[j], q.heap[i]
}

func (q *Pdeq) bubbleUp(pos int) {
	heap := q.heap
	switch level(pos) {
	case LvMin:
		if pos > 0 {
			if q.less(heap[parent(pos)], heap[pos]) {
				q.swap(pos, parent(pos))
				q.bubbleUpMax(parent(pos))
			} else {
				q.bubbleUpMin(pos)
			}
		}
		break
	case LvMax:
		if pos > 0 {
			if q.less(heap[pos], heap[parent(pos)]) {
				q.swap(pos, parent(pos))
				q.bubbleUpMin(parent(pos))
			} else {
				q.bubbleUpMax(pos)
			}
		}
		break
	}
}

func (q *Pdeq) bubbleUpMin(pos int) {
	heap := q.heap
	grandParent := parent(parent(pos))
	if pos > 2 {
		if q.less(heap[pos], heap[grandParent]) {
			q.swap(pos, grandParent)
			q.bubbleUpMin(grandParent)
		}
	}
}

func (q *Pdeq) bubbleUpMax(pos int) {
	heap := q.heap
	grandParent := parent(parent(pos))
	if pos > 2 {
		if q.less(heap[grandParent], heap[pos]) {
			q.swap(pos, grandParent)
			q.bubbleUpMax(grandParent)
		}
	}
}

func (q *Pdeq) trickleDown(pos int) {
	switch level(pos) {
	case LvMin:
		q.trickleDownMin(pos)
		break
	case LvMax:
		q.trickleDownMax(pos)
		break
	}
}

func (q *Pdeq) children(parents []int) []int {
	heap := q.heap
	size := len(heap)
	res := []int{}
	for _, pos := range parents {
		if leftChildren(pos) < size {
			res = append(res, leftChildren(pos))
		}
		if rightChildren(pos) < size {
			res = append(res, rightChildren(pos))
		}
	}
	return res
}

func (q *Pdeq) sort(items []int) []int {
	heap := q.heap
	for i := 0; i < len(items); i++ {
		for j := i + 1; j < len(items); j++ {
			if q.less(heap[items[j]], heap[items[i]]) {
				items[i], items[j] = items[j], items[i]
			}
		}
	}
	return items
}

func (q *Pdeq) trickleDownMin(pos int) {
	heap := q.heap
	children := q.children([]int{pos})
	if len(children) > 0 {
		grandchild := q.children(children)
		if len(grandchild) > 0 {
			children = append(children, grandchild...)
			opts := q.sort(children)
			opt := opts[0]
			if q.less(heap[opt], heap[pos]) {
				q.swap(opt, pos)
				if q.less(heap[parent(opt)], heap[pos]) {
					q.swap(opt, parent(opt))
				}
				q.trickleDownMin(opt)
			}
		} else {
			opts := q.sort(children)
			opt := opts[0]
			if q.less(heap[opt], heap[pos]) {
				q.swap(opt, pos)
			}
		}
	}
}

func (q *Pdeq) trickleDownMax(pos int) {
	heap := q.heap
	children := q.children([]int{pos})
	if len(children) > 0 {
		grandchild := q.children(children)
		if len(grandchild) > 0 {
			children = append(children, grandchild...)
			opts := q.sort(children)
			opt := opts[len(opts)-1]
			if q.less(heap[pos], heap[opt]) {
				q.swap(opt, pos)
				if q.less(heap[opt], heap[parent(opt)]) {
					q.swap(opt, parent(opt))
				}
				q.trickleDownMax(opt)
			}
		} else {
			opts := q.sort(children)
			opt := opts[len(opts)-1]
			if q.less(heap[pos], heap[opt]) {
				q.swap(opt, pos)
			}
		}
	}
}
