package libxdr

import (
	"bytes"
	"io"

	"github.com/groundhogchain/lib-go/libmem"
	"github.com/xdrpp/goxdr/xdr"
)

func MarshalXdrWriter(w io.Writer, value xdr.XdrType) {
	out := xdr.XdrOut{Out: w}
	out.Marshal("", value)
}

type XdrMarshaler struct {
	bpool *libmem.BufferPool
}

func NewXdrMarshaler() XdrMarshaler {
	return XdrMarshaler{bpool: libmem.NewBufferPool()}
}

func (x XdrMarshaler) Borrow() libmem.BorrowedBuffer {
	return x.bpool.Get()
}

func (x XdrMarshaler) MarshalXdr(value xdr.XdrType) libmem.BorrowedBuffer {
	bb := x.bpool.Get()
	MarshalXdrWriter(bb, value)
	return bb
}

func UnmarshalXdr(val xdr.XdrType, in []byte) (e error) {
	return UnmarshalXdrReader(val, bytes.NewBuffer(in))
}

func UnmarshalXdrReader(val xdr.XdrType, r io.Reader) (e error) {
	defer func() {
		switch i := recover().(type) {
		case nil:
		case xdr.XdrError:
			e = i
		default:
			panic(i)
		}
	}()
	val.XdrMarshal(xdr.XdrIn{In: r}, "")
	return nil
}
