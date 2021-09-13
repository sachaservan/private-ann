package hash

import "container/heap"

// This class implements a priority queue
// It it used to implement the priority queue for a breadth first search for the #length closest points

type DistanceSearchQueue struct {
	inserted map[int]bool
	sources  [][]*Element
	final    []*Element
	queue    []*Element
	length   int
}

type Element struct {
	coords   []int
	distance float64
}

func NewDistanceSearchQueue(length int, sources [][]*Element) *DistanceSearchQueue {
	return &DistanceSearchQueue{
		inserted: make(map[int]bool),
		queue:    make([]*Element, 0),
		final:    make([]*Element, 0),
		sources:  sources,
		length:   length,
	}
}

func (d *DistanceSearchQueue) Insert(e *Element) bool {
	if e == nil {
		return false
	}
	// don't double insert
	id := e.id(d.length)
	if d.inserted[id] {
		return false
	}
	d.inserted[id] = true
	heap.Push(d, e)
	return true
}

func (d *DistanceSearchQueue) Search() []*Element {
	baseElement := &Element{}
	for _, s := range d.sources {
		baseElement.CombineWith(s[0])
	}
	d.Insert(baseElement)
	for {
		c := d.GetNextCandidateToExpand()
		if c == nil {
			return d.GetBest()
		}
		for s := range d.sources {
			d.Insert(c.IncrementCopy(s, d.sources))
		}
	}
}

func (d *DistanceSearchQueue) GetNextCandidateToExpand() *Element {
	if len(d.final) == d.length || d.Len() <= 0 {
		return nil
	}
	return heap.Pop(d).(*Element)
}

func (d *DistanceSearchQueue) GetBest() []*Element {
	return d.final
}

func (d *DistanceSearchQueue) Len() int {
	return len(d.queue)
}

func (d *DistanceSearchQueue) Less(i, j int) bool {
	return d.queue[i].distance < d.queue[j].distance
}

func (d *DistanceSearchQueue) Swap(i, j int) {
	d.queue[i], d.queue[j] = d.queue[j], d.queue[i]
}

func (d *DistanceSearchQueue) Push(x interface{}) {
	e := x.(*Element)
	d.queue = append(d.queue, e)
}

func (d *DistanceSearchQueue) Pop() interface{} {
	e := d.queue[len(d.queue)-1]
	d.queue = d.queue[:len(d.queue)-1]
	d.final = append(d.final, e)
	return e
}

func (e *Element) id(domainSize int) int {
	id := 0
	for _, i := range e.coords {
		id *= domainSize
		id += i
	}
	return id
}

func (e *Element) CombineWith(e2 *Element) {
	e.coords = append(e.coords, e2.coords...)
	e.distance += e2.distance
}

func (e *Element) IncrementCopy(pos int, sources [][]*Element) *Element {
	e2 := &Element{coords: make([]int, len(e.coords))}
	for i := range e.coords {
		if i == pos {
			e2.coords[i] = e.coords[i] + 1
			if e2.coords[i] >= len(sources[i]) {
				return nil
			}
		} else {
			e2.coords[i] = e.coords[i]
		}
		e2.distance += sources[i][e2.coords[i]].distance
	}
	return e2
}
