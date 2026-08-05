package connection

import "time"

type Backoff struct {
	current time.Duration
	max     time.Duration
}

func NewBackoff() *Backoff {
	return &Backoff{
		current: 5 * time.Second,
		max:     60 * time.Second,
	}
}

func (b *Backoff) Next() time.Duration {

	delay := b.current

	b.current *= 2

	if b.current > b.max {
		b.current = b.max
	}

	return delay
}

func (b *Backoff) Reset() {
	b.current = 5 * time.Second
}
