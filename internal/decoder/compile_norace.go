//go:build !race
// +build !race

package decoder

import (
	"unsafe"

	"github.com/seniorGolang/json/internal/runtime"
)

func CompileToGetDecoder(typ *runtime.Type) (Decoder, error) {

	typePtr := uintptr(unsafe.Pointer(typ))
	if typePtr > typeAddr.MaxTypeAddr {
		return compileToGetDecoderSlowPath(typePtr, typ)
	}

	index := (typePtr - typeAddr.BaseTypeAddr) >> typeAddr.AddrShift
	if dec := cachedDecoder[index]; dec != nil {
		return dec, nil
	}

	dec, err := compileHead(typ, map[uintptr]Decoder{})
	if err != nil {
		return nil, err
	}
	cachedDecoder[index] = dec
	return dec, nil
}
