package abi

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

const (
	IntTy byte = iota
	UintTy
	BoolTy
	StringTy
	SliceTy
	ArrayTy
	TupleTy
	IdentityTy
	FixedBytesTy
	BytesTy
	HashTy
	FixedPointTy
	FunctionTy
)

type Type struct {
	Elem *Type
	Kind reflect.Kind
	Type reflect.Type
	Size int
	T    byte

	stringKind string

	TupleElems    []*Type
	TupleRawNames []string
}

var (
	typeRegex = regexp.MustCompile("([a-zA-Z]+)(([0-9]+)(x([0-9]+))?)?")
)

func NewType(myType string, myArgs []ArgumentMarshaling) (finalType Type, err error) {
	if strings.Count(myType, "[") != strings.Count(myType, "]") {
		return Type{}, fmt.Errorf("invalid arg type in abi")
	}

	finalType.stringKind = myType

	if strings.Count(myType, "[") != 0 {
		i := strings.LastIndex(myType, "[")
		embeddedType, err := NewType(myType[:i], myArgs)
		if err != nil {
			return Type{}, err
		}
		sliced := myType[i:]
		re := regexp.MustCompile("[0-9]+")
		mySliceIndex := re.FindAllString(sliced, -1)

		if len(mySliceIndex) == 1 {
			finalType.T = ArrayTy
			finalType.Kind = reflect.Array
			finalType.Elem = &embeddedType
			finalType.Size, err = strconv.Atoi(mySliceIndex[0])
			if err != nil {
				return Type{}, fmt.Errorf("abi: error parsing variable size: %v", err)
			}
			finalType.Type = reflect.ArrayOf(finalType.Size, embeddedType.Type)
			if embeddedType.T == TupleTy {
				finalType.stringKind = embeddedType.stringKind + sliced
			}
		} else if len(mySliceIndex) == 0 {
			finalType.T = SliceTy
			finalType.Kind = reflect.Slice
			finalType.Elem = &embeddedType
			finalType.Type = reflect.SliceOf(embeddedType.Type)
			if embeddedType.T == TupleTy {
				finalType.stringKind = embeddedType.stringKind + sliced
			}
		} else {
			return Type{}, fmt.Errorf("invalid formatting of array type")
		}
		return finalType, err
	}
	myRegexIndex := typeRegex.FindAllStringSubmatch(myType, -1)
	if len(myRegexIndex) == 0 {
		return Type{}, fmt.Errorf("invalid type '%v'", myType)
	}
	parsedType := myRegexIndex[0]

	var varSize int
	if len(parsedType[3]) > 0 {
		var err error
		varSize, err = strconv.Atoi(parsedType[2])
		if err != nil {
			return Type{}, fmt.Errorf("abi: error parsing variable size: %v", err)
		}
	} else {
		if parsedType[0] == "uint" || parsedType[0] == "int" {
			return Type{}, fmt.Errorf("unsupported arg type: %s", myType)
		}
	}
	switch myParsedType := parsedType[1]; myParsedType {
	case "uint":
		finalType.Kind, finalType.Type = reflectIntKindAndType(true, varSize)
		finalType.T = UintTy
		finalType.Size = varSize
	case "bool":
		finalType.Kind = reflect.Bool
		finalType.T = BoolTy
		finalType.Type = reflect.TypeOf(bool(false))
	case "string":
		finalType.Kind = reflect.String
		finalType.T = StringTy
		finalType.Type = reflect.TypeOf("")
	case "int":
		finalType.Kind, finalType.Type = reflectIntKindAndType(false, varSize)
		finalType.T = IntTy
		finalType.Size = varSize
	case "bytes":
		if varSize == 0 {
			finalType.T = BytesTy
			finalType.Kind = reflect.Slice
			finalType.Type = reflect.SliceOf(reflect.TypeOf(byte(0)))
		} else {
			finalType.T = FixedBytesTy
			finalType.Kind = reflect.Array
			finalType.Size = varSize
			finalType.Type = reflect.ArrayOf(varSize, reflect.TypeOf(byte(0)))
		}
	case "function":
		finalType.Kind = reflect.Array
		finalType.T = FunctionTy
		finalType.Size = 24
		finalType.Type = reflect.ArrayOf(24, reflect.TypeOf(byte(0)))
	case "identity":
		finalType.Kind = reflect.Array
		finalType.Type = identityT
		finalType.T = IdentityTy
		finalType.Size = 32
	case "tuple":
		var (
			fields     []reflect.StructField
			elems      []*Type
			names      []string
			expression string
		)
		expression += "("
		for idx, c := range myArgs {
			cType, err := NewType(c.Type, c.Components)
			if err != nil {
				return Type{}, err
			}
			if ToCamelCase(c.Name) == "" {
				return Type{}, errors.New("abi: purely anonymous or underscored field is not supported")
			}
			fields = append(fields, reflect.StructField{
				Name: ToCamelCase(c.Name),
				Type: cType.Type,
				Tag:  reflect.StructTag("json:\"" + c.Name + "\""),
			})
			elems = append(elems, &cType)
			names = append(names, c.Name)
			expression += cType.stringKind
			if idx != len(myArgs)-1 {
				expression += ","
			}
		}
		expression += ")"
		finalType.Kind = reflect.Struct
		finalType.Type = reflect.StructOf(fields)
		finalType.TupleElems = elems
		finalType.TupleRawNames = names
		finalType.T = TupleTy
		finalType.stringKind = expression
	default:
		return Type{}, fmt.Errorf("unsupported arg type: %s", myType)
	}

	return
}

func (t Type) String() (out string) {
	return t.stringKind
}

//func (t Type) pack(v reflect.Value) ([]byte, error) {
//	v = indirect(v)
//	if err := typeCheck(t, v); err != nil {
//		return nil, err
//	}
//
//	switch t.T {
//	case SliceTy, ArrayTy:
//		var ret []byte
//
//		if t.requiresLengthPrefix() {
//			ret = append(ret, packNum(reflect.ValueOf(v.Len()))...)
//		}
//
//		offset := 0
//		offsetReq := isDynamicType(*t.Elem)
//		if offsetReq {
//			offset = getTypeSize(*t.Elem) * v.Len()
//		}
//		var tail []byte
//		for i := 0; i < v.Len(); i++ {
//			val, err := t.Elem.pack(v.Index(i))
//			if err != nil {
//				return nil, err
//			}
//			if !offsetReq {
//				ret = append(ret, val...)
//				continue
//			}
//			ret = append(ret, packNum(reflect.ValueOf(offset))...)
//			offset += len(val)
//			tail = append(tail, val...)
//		}
//		return append(ret, tail...), nil
//	case TupleTy:
//		fieldmap, err := mapArgNamesToStructFields(t.TupleRawNames, v)
//		if err != nil {
//			return nil, err
//		}
//		offset := 0
//		for _, elem := range t.TupleElems {
//			offset += getTypeSize(*elem)
//		}
//		var ret, tail []byte
//		for i, elem := range t.TupleElems {
//			field := v.FieldByName(fieldmap[t.TupleRawNames[i]])
//			if !field.IsValid() {
//				return nil, fmt.Errorf("field %s for tuple not found in the given struct", t.TupleRawNames[i])
//			}
//			val, err := elem.pack(field)
//			if err != nil {
//				return nil, err
//			}
//			if isDynamicType(*elem) {
//				ret = append(ret, packNum(reflect.ValueOf(offset))...)
//				tail = append(tail, val...)
//				offset += len(val)
//			} else {
//				ret = append(ret, val...)
//			}
//		}
//		return append(ret, tail...), nil
//
//	default:
//		return packElement(t, v), nil
//	}
//}

func (t Type) requiresLengthPrefix() bool {
	return t.T == StringTy || t.T == BytesTy || t.T == SliceTy
}

func isDynamicType(t Type) bool {
	if t.T == TupleTy {
		for _, elem := range t.TupleElems {
			if isDynamicType(*elem) {
				return true
			}
		}
		return false
	}
	return t.T == StringTy || t.T == BytesTy || t.T == SliceTy || (t.T == ArrayTy && isDynamicType(*t.Elem))
}

func getTypeSize(t Type) int {
	if t.T == ArrayTy && !isDynamicType(*t.Elem) {
		if t.Elem.T == ArrayTy {
			return t.Size * getTypeSize(*t.Elem)
		}
		return t.Size * 32
	} else if t.T == TupleTy && !isDynamicType(t) {
		total := 0
		for _, elem := range t.TupleElems {
			total += getTypeSize(*elem)
		}
		return total
	}
	return 32
}
