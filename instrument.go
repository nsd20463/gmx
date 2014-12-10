package gmx

import (
	"fmt"
	"sync/atomic"
)

type Counter struct {
	value uint64
}

func (c *Counter) Inc() {
	atomic.AddUint64(&c.value, 1)
}

func (c Counter) String() string {
	return fmt.Sprintf("%u", c.value)
}

func (c Counter) Value() uint64 {
	return c.value
}

func NewCounter(name string) *Counter {
	c := new(Counter)
	Publish(name, func() interface{} {
		return c.value
	})
	return c
}

type Gauge struct {
	value int64
}

func (g *Gauge) Inc() {
	atomic.AddInt64(&g.value, 1)
}

func (g *Gauge) Dec() {
	atomic.AddInt64(&g.value, -1)
}

func (g Gauge) String() string {
	return fmt.Sprintf("%d", g.value)
}

func (g Gauge) Value() int64 {
	return g.value
}

func NewGauge(name string) *Gauge {
	g := new(Gauge)
	Publish(name, func() interface{} {
		return g.value
	})
	return g
}
