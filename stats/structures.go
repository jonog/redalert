package stats

import "time"

type counter struct {
	count int
}

func newCounter() *counter {
	return &counter{}
}

func (c *counter) Inc() int {
	c.count++
	return c.count
}

func (c *counter) Reset() {
	c.count = 0
}

type occurrence struct {
	t time.Time
}

func newOccurrence() *occurrence {
	return &occurrence{}
}

func (o *occurrence) Mark() {
	o.t = time.Now()
}
