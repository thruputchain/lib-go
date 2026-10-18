package libflow

import (
	"context"
	"time"

	"github.com/thruputchain/lib-go/libtime"
	"github.com/thruputchain/lib-go/must"
)

type Flush[T any] func(context.Context, []T)
type SizeOf[T any] func(T) int

type BatchingBuffer[T any] struct {
	time    libtime.Time
	flush   Flush[T]
	maxSize int
	maxWait time.Duration
	sizeOf  SizeOf[T]
	ch      chan T
}

func NewBatchingBuffer[T any](
	ctx context.Context,
	time libtime.Time,
	flush Flush[T],
	maxSize int,
	maxWait time.Duration,
	sizeOf SizeOf[T],
) *BatchingBuffer[T] {
	must.Assert(ctx, maxSize > 0)
	must.Assert(ctx, maxWait > 0)
	buf := &BatchingBuffer[T]{
		time:    time,
		flush:   flush,
		maxSize: maxSize,
		maxWait: maxWait,
		sizeOf:  sizeOf,
		ch:      make(chan T, maxSize),
	}
	go buf.loop(ctx)
	return buf
}

func (buf *BatchingBuffer[T]) Submit(ctx context.Context, xs T) {
	buf.ch <- xs
}

func (buf *BatchingBuffer[T]) loop(ctx context.Context) {
	for {
		var pending []T
		wake := buf.time.After(buf.maxWait)
	fill:
		for buf.SizeOf(pending) < buf.maxSize {
			select {
			case x := <-buf.ch:
				pending = append(pending, x)
			case <-ctx.Done():
				return
			case <-wake:
				break fill
			}
		}

		if buf.SizeOf(pending) > 0 {
			buf.flush(ctx, pending)
			pending = pending[:0]
		}
	}
}

func (buf *BatchingBuffer[T]) SizeOf(ts []T) int {
	n := 0
	for _, t := range ts {
		n += buf.sizeOf(t)
	}
	return n
}
