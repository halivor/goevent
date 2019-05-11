package eventpool

import (
	"fmt"
	"time"
)

type Expire interface {
	ExpAt() int64 // ms
	Event
}

type minHeap struct {
	ei map[Expire]int
	es []Expire
	a  int
}

func newHeap() *minHeap {
	return &minHeap{
		ei: make(map[Expire]int),
		es: make([]Expire, maxConns),
		a:  0,
	}
}

func (mh *minHeap) ExpireAfter() int {
	if mh.a == 0 {
		return 1000
	}
	return int(mh.es[0].ExpAt() - time.Now().UnixNano()/(1000*1000))
}

func (mh *minHeap) Push(ce Expire) {
	mh.up(mh.a, ce)
	mh.a++
}
func (mh *minHeap) up(ci int, ce Expire) {
	for pi := 0; ci > 0; ci = pi {
		pi = (ci - 1) / 2
		pe := mh.es[pi]
		if pe.ExpAt() <= ce.ExpAt() {
			break
		}
		mh.es[ci] = pe
		mh.ei[pe] = ci
	}
	mh.ei[ce] = ci
	mh.es[ci] = ce
}
func (mh *minHeap) Pop() Expire {
	mh.a--
	return mh.down(0)
}
func (mh *minHeap) down(pi int) Expire {
	if mh.a >= 0 {
		e := mh.es[pi]
		pe := mh.es[mh.a]
		for ci := 2 * (pi + 1); ci <= mh.a; ci = 2 * (pi + 1) {
			if ci == mh.a || mh.es[ci].ExpAt() > mh.es[ci-1].ExpAt() {
				ci--
			}
			ce := mh.es[ci]
			if ce.ExpAt() >= pe.ExpAt() {
				break
			}
			mh.es[pi] = ce
			mh.ei[ce] = pi
			pi = ci
		}
		mh.es[pi] = pe
		return e
	}
	return nil
}

func (mh *minHeap) Erase(e Expire) {
	if ci, ok := mh.ei[e]; ok {
		mh.a--
		pi := (ci - 1) / 2
		ce := mh.es[mh.a]
		pe := mh.es[pi]
		if pi > 0 && pe.ExpAt() > ce.ExpAt() {
			mh.up(ci, ce)
		} else {
			mh.down(ci)
		}
	}
}

func (mh *minHeap) Top() int64 {
	if mh.a == 0 {
		return int64(uint64(mh.a-1) >> 1)
	}
	return mh.es[0].ExpAt()
}

func (mh *minHeap) String() string {
	str := ""
	for i, j := 0, 0; i < mh.a; i++ {
		str += fmt.Sprintf("%d ", mh.es[i].ExpAt())
		if i == j {
			str += fmt.Sprintf("\n")
			j = 2 * (i + 1)
		}
	}
	str += fmt.Sprintf("\n")
	str += fmt.Sprintf("\n")
	return str
}
