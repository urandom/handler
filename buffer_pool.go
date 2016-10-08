package handler

import (
	"bytes"
	"sync"
)

type bufPool struct {
	sync.Pool
}

var bufferPool bufPool

func (p bufPool) Get() *bytes.Buffer {
	return p.Pool.Get().(*bytes.Buffer)
}

func (p *bufPool) Put(b *bytes.Buffer) {
	b.Reset()
	p.Pool.Put(b)
}

func init() {
	bufferPool = bufPool{
		Pool: sync.Pool{
			New: func() interface{} {
				return new(bytes.Buffer)
			},
		},
	}
}
