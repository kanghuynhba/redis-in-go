package main

type Deque struct {
	buf   []string
	head  int
	tail  int
	count int
}

func NewDeque() *Deque {
	return &Deque{
		buf: make([]string, 8),
	}
}

func (d *Deque) Len() int {
	return d.count
}

func (d *Deque) grow() {
	newBuf := make([]string, d.Len()*2)
	for i := 0; i < d.Len(); i++ {
		newBuf[i] = d.buf[(d.head+i)%len(d.buf)]
	}
	d.buf = newBuf
	d.head = 0
	d.tail = d.Len()
}

func (d *Deque) PushBack(value string) {
	if d.Len() == len(d.buf) {
		d.grow()
	}

	d.buf[d.tail] = value
	d.tail = (d.tail + 1) % len(d.buf)
	d.count++
}

func (d *Deque) PushFront(value string) {
	if d.Len() == len(d.buf) {
		d.grow()
	}
	d.head = (d.head - 1 + len(d.buf)) % len(d.buf)
	d.buf[d.head] = value
	d.count++
}

func (d *Deque) PushMultipleValues(values []string, isBack bool) {
	if isBack {
		for _, val := range values {
			d.PushBack(val)
		}
		return
	}

	for _, val := range values {
		d.PushFront(val)
	}
}

func (d *Deque) PopBack() (string, bool) {
	if d.Len() == 0 {
		return "", false
	}
	d.tail = (d.tail - 1 + len(d.buf)) % len(d.buf)
	val := d.buf[d.tail]
	d.buf[d.tail] = ""
	d.count--
	return val, true
}

func (d *Deque) PopFront() (string, bool) {
	if d.Len() == 0 {
		return "", false
	}
	val := d.buf[d.head]
	d.buf[d.head] = ""
	d.head = (d.head + 1) % len(d.buf)
	d.count--
	return val, true
}

func (d *Deque) Range(start, stop int) []string {
	values := []string{}

	if start > stop || d.count == 0 {
		return values
	}

	for i := start; i <= stop; i++ {
		physicalIdx := (d.head + i) % len(d.buf)
		values = append(values, d.buf[physicalIdx])
	}

	return values
}
