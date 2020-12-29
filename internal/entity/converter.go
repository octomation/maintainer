package entity

import (
	"reflect"
	"unsafe"
)

// ConvertStringsToRepositoryURN converts a list of string
// to a list of RepositoryURN but not guarantee its validity.
func ConvertStringsToRepositoryURNs(list []string) []RepositoryURN {
	header := (*reflect.SliceHeader)(unsafe.Pointer(&list))
	converted := (*[]RepositoryURN)(unsafe.Pointer(header))
	return *converted
}
