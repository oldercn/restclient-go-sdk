package abi

import (
	"errors"
)

var (
	errBadBool = errors.New("abi: improperly encoded boolean value")
)

//func formatSliceString(kind reflect.Kind, sliceSize int) string {
//	if sliceSize == -1 {
//		return fmt.Sprintf("[]%v", kind)
//	}
//	return fmt.Sprintf("[%d]%v", sliceSize, kind)
//}
//
//func typeCheck(typ Type, value reflect.Value) error {
//	if typ.T == SliceTy || typ.T == ArrayTy {
//		return sliceTypeCheck(typ, value)
//	}
//
//	if typ.Kind != value.Kind() {
//		return typeErr(typ.Kind, value.Kind())
//	} else if typ.T == FixedBytesTy && typ.Size != value.Len() {
//		return typeErr(typ.Type, value.Type())
//	} else {
//		return nil
//	}
//
//}
//
//func typeErr(expected, got interface{}) error {
//	return fmt.Errorf("abi: cannot use %v as type %v as argument", got, expected)
//}
//
//func sliceTypeCheck(typ Type, val reflect.Value) error {
//	if val.Kind() != reflect.Slice && val.Kind() != reflect.Array {
//		return typeErr(formatSliceString(typ.Kind, typ.Size), val.Type())
//	}
//
//	if typ.T == ArrayTy && val.Len() != typ.Size {
//		return typeErr(formatSliceString(typ.Elem.Kind, typ.Size), formatSliceString(val.Type().Elem().Kind(), val.Len()))
//	}
//
//	if typ.Elem.T == SliceTy {
//		if val.Len() > 0 {
//			return sliceTypeCheck(*typ.Elem, val.Index(0))
//		}
//	} else if typ.Elem.T == ArrayTy {
//		return sliceTypeCheck(*typ.Elem, val.Index(0))
//	}
//
//	if elemKind := val.Type().Elem().Kind(); elemKind != typ.Elem.Kind {
//		return typeErr(formatSliceString(typ.Elem.Kind, typ.Size), val.Type())
//	}
//	return nil
//}