package utils

import (
	"reflect"
)

/*
 IsEmpty は任意の値が「空」であるかを判定する。空の定義は以下の通り：
 - nil
 - ゼロ値（数値なら0、文字列なら空文字、スライスやマップなら長さ0など）
*/
func IsEmpty(any interface{}) bool {
	if any == nil {
		return true
	}

	v := reflect.ValueOf(any)

	switch v.Kind() {
	case reflect.String, reflect.Array, reflect.Slice, reflect.Map:
		return v.Len() == 0
	case reflect.Ptr, reflect.Interface:
		return v.IsNil()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Struct:
		return reflect.DeepEqual(any, reflect.New(v.Type()).Elem().Interface())
	default:
		return false
	}
}

// IsSliceEmpty はスライスがnilまたは長さ0かを判定
func isSliceEmpty[T any](s []T) bool {
	return len(s) == 0
}

// IsPtrNil はポインタがnilかどうかを判定
func isPtrNil[T any](ptr *T) bool {
	return ptr == nil
}