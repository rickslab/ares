package util

import (
	"reflect"
)

// Copy src to dst from different struct
//  1. dst *struct/**struct
//  2. src struct/*struct
func xcopy(dstPtr reflect.Value, src reflect.Value) {
	f := src.MethodByName("Xcopy")
	if f.IsValid() {
		f.Call([]reflect.Value{dstPtr})
		return
	}

	dst := dstPtr.Elem()
	if dst.Kind() == reflect.Ptr {
		dst = reflect.New(dst.Type().Elem()).Elem()
		dstPtr.Elem().Set(dst.Addr())
	}
	if src.Kind() == reflect.Ptr {
		src = src.Elem()
	}

	srcType := src.Type()
	for i := 0; i < srcType.NumField(); i++ {
		srcFieldType := srcType.Field(i)
		tag := srcFieldType.Tag.Get("xcopy")
		if tag == "-" {
			continue
		}

		srcField := src.Field(i)
		if srcField.Kind() == reflect.Ptr {
			srcField = srcField.Elem()
		}
		if !srcField.IsValid() {
			continue
		}

		dstField := dst.FieldByName(srcFieldType.Name)
		if !dstField.IsValid() {
			continue
		}

		switch tag {
		case "int":
			dstField.SetInt(srcField.Int())
		case "":
			dstField.Set(srcField)
		default:
			vs := srcField.MethodByName(tag).Call([]reflect.Value{})
			dstField.Set(vs[0])
		}
	}
}

// Copy src to dst from slice of different element struct
//  1. dst *[]struct/*[]*struct
//  2. src []struct/*[]struct or []*struct/*[]*struct
func xcopySlice(dstPtr reflect.Value, src reflect.Value) {
	dst := dstPtr.Elem()
	if src.Kind() == reflect.Ptr {
		src = src.Elem()
	}

	srcLen := src.Len()
	dstLen := dst.Len()

	dstType := dst.Type()
	dstElemType := dstType.Elem()

	if dstElemType.Kind() == reflect.Ptr {
		for i := 0; i < srcLen; i++ {
			var dstElem reflect.Value
			if i < dstLen {
				dstElem = dst.Index(i)

				if dstElem.IsNil() {
					dstElem.Set(reflect.New(dstElemType.Elem()))
				}
			} else {
				dstElem = reflect.New(dstElemType.Elem())

				dst = reflect.Append(dst, dstElem)
				dstPtr.Elem().Set(dst)
			}

			xcopy(dstElem, src.Index(i))
		}
	} else {
		for i := 0; i < srcLen; i++ {
			var dstElem reflect.Value
			if i < dstLen {
				dstElem = dst.Index(i)

				xcopy(dstElem.Addr(), src.Index(i))
			} else {
				dstElem = reflect.New(dstElemType)

				xcopy(dstElem, src.Index(i))

				dst = reflect.Append(dst, dstElem.Elem())
				dstPtr.Elem().Set(dst)
			}
		}
	}
}

// Xcopy copy src to dst of different type
//  1. Values of different struct
//  2. Slices of different element struct
func Xcopy(dst interface{}, src interface{}) interface{} {
	if src == nil {
		return dst
	}

	srcVal := reflect.ValueOf(src)
	dstVal := reflect.ValueOf(dst)

	if srcVal.Kind() == reflect.Slice {
		xcopySlice(dstVal, srcVal)
		return dstVal.Elem().Interface()
	}

	if srcVal.Kind() == reflect.Ptr {
		if srcVal.Elem().Kind() == reflect.Slice {
			xcopySlice(dstVal, srcVal)
			return dstVal.Elem().Interface()
		}
	}

	xcopy(dstVal, srcVal)

	if dstVal.Type().Elem().Kind() == reflect.Ptr {
		return dstVal.Elem().Interface()
	}
	return dst
}
