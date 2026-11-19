package gpu

type Queue struct {
	buf             []FifoEntry
	head, tail, cnt int
}

func NewQueue(size int) *Queue   { return &Queue{buf: make([]FifoEntry, size)} }
func (q *Queue) Reset()          { q.head, q.tail, q.cnt = 0, 0, 0 }
func (q *Queue) Peek() FifoEntry { return q.buf[q.head] }
func (q *Queue) Length() int     { return int(q.cnt) }
func (q *Queue) IsEmpty() bool   { return q.cnt == 0 }
func (q *Queue) IsFull() bool    { return q.cnt == len(q.buf) }

func (q *Queue) Enqueue(v FifoEntry) bool {
	if q.IsFull() {
		return false
	}

	q.buf[q.tail] = v
	q.tail = (q.tail + 1) % len(q.buf)
	q.cnt++
	return true
}

func (q *Queue) Dequeue() FifoEntry {
	if q.IsEmpty() {
		return q.buf[q.head]
	}

	v := q.buf[q.head]
	q.head = (q.head + 1) % len(q.buf)
	q.cnt--
	return v
}
