package libxdr

import (
	"crypto/sha256"
	"encoding/hex"

	"github.com/xdrpp/goxdr/xdr"
)

func HashXdr(m XdrMarshaler, value xdr.XdrType) [32]byte {
	bb := m.MarshalXdr(value)
	defer bb.Recycle()
	return sha256.Sum256(bb.Bytes())
}

func HashString(h [32]byte) string {
	return hex.EncodeToString(h[:])[:12]
}

func HomoHashString(h [2048]byte) string {
	return hex.EncodeToString(h[:])[:12]
}
