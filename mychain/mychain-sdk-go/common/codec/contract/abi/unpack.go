package abi

import (
	"encoding/binary"
	"fmt"
	"github.com/oldercn/restclient-go-sdk/mychain/mychain-sdk-go/common/codec"
	"github.com/oldercn/restclient-go-sdk/mychain/mychain-sdk-go/domain"
	"math/big"
	"reflect"
)

var (
	maxUint256 = big.NewInt(0).Add(
		big.NewInt(0).Exp(big.NewInt(2), big.NewInt(256), nil),
		big.NewInt(-1))
	maxInt256 = big.NewInt(0).Add(
		big.NewInt(0).Exp(big.NewInt(2), big.NewInt(255), nil),
		big.NewInt(-1))
)

func readInteger(myType byte, myKind reflect.Kind, myBytes []byte) interface{} {
	switch myKind {
	case reflect.Uint8:
		return myBytes[len(myBytes)-1]
	case reflect.Uint16:
		return binary.BigEndian.Uint16(myBytes[len(myBytes)-2:])
	case reflect.Uint32:
		return binary.BigEndian.Uint32(myBytes[len(myBytes)-4:])
	case reflect.Uint64:
		return binary.BigEndian.Uint64(myBytes[len(myBytes)-8:])
	case reflect.Int8:
		return int8(myBytes[len(myBytes)-1])
	case reflect.Int16:
		return int16(binary.BigEndian.Uint16(myBytes[len(myBytes)-2:]))
	case reflect.Int32:
		return int32(binary.BigEndian.Uint32(myBytes[len(myBytes)-4:]))
	case reflect.Int64:
		return int64(binary.BigEndian.Uint64(myBytes[len(myBytes)-8:]))
	default:
		ret := new(big.Int).SetBytes(myBytes)
		if myType == UintTy {
			return ret
		}

		if ret.Cmp(maxInt256) > 0 {
			ret.Add(maxUint256, big.NewInt(0).Neg(ret))
			ret.Add(ret, big.NewInt(1))
			ret.Neg(ret)
		}
		return ret
	}
}

func lengthPrefixPointsTo(myIndex int, inputBytets []byte) (finalStart int, finalLen int, err error) {
	bigOffsetEnd := big.NewInt(0).SetBytes(inputBytets[myIndex : myIndex+32])
	bigOffsetEnd.Add(bigOffsetEnd, codec.Big32)
	outputLength := big.NewInt(int64(len(inputBytets)))

	if bigOffsetEnd.Cmp(outputLength) > 0 {
		return 0, 0, fmt.Errorf("abi: cannot marshal in to go slice: offset %v would go over slice boundary (len=%v)",
			bigOffsetEnd, outputLength)
	}

	if bigOffsetEnd.BitLen() > 63 {
		return 0, 0, fmt.Errorf("abi offset larger than int64: %v", bigOffsetEnd)
	}

	offsetEnd := int(bigOffsetEnd.Uint64())
	lengthBig := big.NewInt(0).SetBytes(inputBytets[offsetEnd-32 : offsetEnd])

	totalSize := big.NewInt(0)
	totalSize.Add(totalSize, bigOffsetEnd)
	totalSize.Add(totalSize, lengthBig)
	if totalSize.BitLen() > 63 {
		return 0, 0, fmt.Errorf("abi: length larger than int64: %v", totalSize)
	}

	if totalSize.Cmp(outputLength) > 0 {
		return 0, 0, fmt.Errorf("abi: cannot marshal in to go type: length insufficient %v require %v",
			outputLength, totalSize)
	}
	finalStart = int(bigOffsetEnd.Uint64())
	finalLen = int(lengthBig.Uint64())
	return
}

func readFunctionType(myType Type, word []byte) (funcTy [24]byte, err error) {
	if myType.T != FunctionTy {
		return [24]byte{}, fmt.Errorf("abi: invalid type in call to make function type byte array")
	}
	if garbage := binary.BigEndian.Uint64(word[24:32]); garbage != 0 {
		err = fmt.Errorf("abi: got improperly encoded function type, got %v", word)
	} else {
		copy(funcTy[:], word[0:24])
	}
	return
}

func forTupleUnpack(t Type, output []byte) (interface{}, error) {
	retval := reflect.New(t.Type).Elem()
	virtualArgs := 0
	for index, elem := range t.TupleElems {
		marshalledValue, err := toGoType((index+virtualArgs)*32, *elem, output)
		if elem.T == ArrayTy && !isDynamicType(*elem) {
			virtualArgs += getTypeSize(*elem)/32 - 1
		} else if elem.T == TupleTy && !isDynamicType(*elem) {
			virtualArgs += getTypeSize(*elem)/32 - 1
		}
		if err != nil {
			return nil, err
		}
		retval.Field(index).Set(reflect.ValueOf(marshalledValue))
	}
	return retval.Interface(), nil
}

func tuplePointsTo(index int, output []byte) (start int, err error) {
	offset := big.NewInt(0).SetBytes(output[index : index+32])
	outputLen := big.NewInt(int64(len(output)))

	if offset.Cmp(big.NewInt(int64(len(output)))) > 0 {
		return 0, fmt.Errorf("abi: cannot marshal in to go slice: offset %v would go over slice boundary (len=%v)", offset, outputLen)
	}
	if offset.BitLen() > 63 {
		return 0, fmt.Errorf("abi offset larger than int64: %v", offset)
	}
	return int(offset.Uint64()), nil
}

func readFixedBytes(myType Type, word []byte) (interface{}, error) {
	if myType.T != FixedBytesTy {
		return nil, fmt.Errorf("abi: invalid type in call to make fixed byte array")
	}
	array := reflect.New(myType.Type).Elem()

	reflect.Copy(array, reflect.ValueOf(word[0:myType.Size]))
	return array.Interface(), nil

}

func readBool(word []byte) (bool, error) {
	for _, b := range word[:31] {
		if b != 0 {
			return false, errBadBool
		}
	}
	switch word[31] {
	case 1:
		return true, nil
	case 0:
		return false, nil
	default:
		return false, errBadBool
	}
}

func toGoType(index int, myType Type, output []byte) (interface{}, error) {
	if index+32 > len(output) {
		return nil, fmt.Errorf("abi: cannot marshal in to go type: length insufficient %d require %d", len(output), index+32)
	}
	var (
		returnOutput  []byte
		begin, length int
		err           error
	)
	if myType.requiresLengthPrefix() {
		begin, length, err = lengthPrefixPointsTo(index, output)
		if err != nil {
			return nil, err
		}
	} else {
		returnOutput = output[index : index+32]
	}
	switch myType.T {
	case TupleTy:
		if isDynamicType(myType) {
			begin, err := tuplePointsTo(index, output)
			if err != nil {
				return nil, err
			}
			return forTupleUnpack(myType, output[begin:])
		} else {
			return forTupleUnpack(myType, output[index:])
		}
	case SliceTy:
		return forEachUnpack(myType, output[begin:], 0, length)
	case ArrayTy:
		if isDynamicType(*myType.Elem) {
			offset := int64(binary.BigEndian.Uint64(returnOutput[len(returnOutput)-8:]))
			return forEachUnpack(myType, output[offset:], 0, myType.Size)
		}
		return forEachUnpack(myType, output[index:], 0, myType.Size)
	case BoolTy:
		return readBool(returnOutput)
	case IntTy, UintTy:
		return readInteger(myType.T, myType.Kind, returnOutput), nil
	case FixedBytesTy:
		return readFixedBytes(myType, returnOutput)
	case HashTy:
		return domain.BytesToHash(returnOutput), nil
	case IdentityTy:
		return domain.BytesToIdentity(returnOutput), nil
	case BytesTy:
		return output[begin : begin+length], nil
	case FunctionTy:
		return readFunctionType(myType, returnOutput)
	case StringTy:
		return string(output[begin : begin+length]), nil
	default:
		return nil, fmt.Errorf("abi: unknown type %v", myType.T)
	}
}

func forEachUnpack(myType Type, myOutputBytes []byte, myStart, size int) (interface{}, error) {
	if size < 0 {
		return nil, fmt.Errorf("cannot marshal input to array, size is negative (%d)", size)
	}
	if myStart+32*size > len(myOutputBytes) {
		return nil, fmt.Errorf("abi: cannot marshal in to go array: offset %d would go over slice boundary (len=%d)", len(myOutputBytes), myStart+32*size)
	}

	var refSlice reflect.Value

	if myType.T == SliceTy {
		refSlice = reflect.MakeSlice(myType.Type, size, size)
	} else if myType.T == ArrayTy {
		refSlice = reflect.New(myType.Type).Elem()
	} else {
		return nil, fmt.Errorf("abi: invalid type in array/slice unpacking stage")
	}

	elemSize := getTypeSize(*myType.Elem)

	for i, j := myStart, 0; j < size; i, j = i+elemSize, j+1 {
		inter, err := toGoType(i, *myType.Elem, myOutputBytes)
		if err != nil {
			return nil, err
		}

		refSlice.Index(j).Set(reflect.ValueOf(inter))
	}

	return refSlice.Interface(), nil
}