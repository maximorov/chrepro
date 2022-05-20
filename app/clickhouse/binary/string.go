package binary

import (
	"reflect"
	"unsafe"
)

// Copied from https://github.com/m3db/m3/blob/master/src/x/unsafe/string.go#L62

func unsafeStr2Bytes(str string) []byte {
	if len(str) == 0 {
		return nil
	}
	var scratch []byte
	{
		slice := (*reflect.SliceHeader)(unsafe.Pointer(&scratch))
		slice.Len = len(str)
		slice.Cap = len(str)
		slice.Data = (*reflect.StringHeader)(unsafe.Pointer(&str)).Data
	}
	return scratch
}
