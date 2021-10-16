//go:build race
// +build race

package decoder

import (
	"sync"
	"unsafe"

	"github.com/seniorGolang/json/internal/runtime"
)

var decMu sync.RWMutex

func CompileToGetDecoder(typ *runtime.Type) (Decoder, error) {

	typePtr := uintptr(unsafe.Pointer(typ))
	if typePtr > typeAddr.MaxTypeAddr {
		return compileToGetDecoderSlowPath(typePtr, typ)
	}

	index := (typePtr - typeAddr.BaseTypeAddr) >> typeAddr.AddrShift
	decMu.RLock()
	if dec := cachedDecoder[index]; dec != nil {
		decMu.RUnlock()
		return dec, nil
	}
	decMu.RUnlock()

	dec, err := compileHead(typ, map[uintptr]Decoder{})
	if err != nil {
		return nil, err
	}
	decMu.Lock()
	cachedDecoder[index] = dec
	decMu.Unlock()
	return dec, nil
}
