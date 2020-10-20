// Package math provides integer math utilities.
package math

import (
	"fmt"
	"math/big"
)


const (
	MaxInt8   = 1<<7 - 1
	MinInt8   = -1 << 7
	MaxInt16  = 1<<15 - 1
	MinInt16  = -1 << 15
	MaxInt32  = 1<<31 - 1
	MinInt32  = -1 << 31
	MaxInt64  = 1<<63 - 1
	MinInt64  = -1 << 63
	MaxUint8  = 1<<8 - 1
	MaxUint16 = 1<<16 - 1
	MaxUint32 = 1<<32 - 1
	MaxUint64 = 1<<64 - 1
)

var (
	tt255     = BigPow(2, 255)
	tt256     = BigPow(2, 256)
	tt256m1   = new(big.Int).Sub(tt256, big.NewInt(1))
	tt63      = BigPow(2, 63)
	MaxBig256 = new(big.Int).Set(tt256m1)
	MaxBig63  = new(big.Int).Sub(tt63, big.NewInt(1))
)

const (
	wordBits = 32 << (uint64(^big.Word(0)) >> 63)
	wordBytes = wordBits / 8
)

type HexOrDecimal256 big.Int


func (i *HexOrDecimal256) UnmarshalText(myInputBytes []byte) error {
	myBigInt, ok := ParseBig256(string(myInputBytes))
	if !ok {
		return fmt.Errorf("invalid hex or decimal integer %q", myInputBytes)
	}
	*i = HexOrDecimal256(*myBigInt)
	return nil
}

func ParseBig256(myStr string) (*big.Int, bool) {
	if myStr == "" {
		return new(big.Int), true
	}
	var bigint *big.Int
	var isOK bool
	if len(myStr) >= 2 && (myStr[:2] == "0x" || myStr[:2] == "0X") {
		bigint, isOK = new(big.Int).SetString(myStr[2:], 16)
	} else {
		bigint, isOK = new(big.Int).SetString(myStr, 10)
	}
	if isOK && bigint.BitLen() > 256 {
		bigint, isOK = nil, false
	}
	return bigint, isOK
}

func (i *HexOrDecimal256) MarshalText() ([]byte, error) {
	if i == nil {
		return []byte("0x0"), nil
	}
	return []byte(fmt.Sprintf("%#x", (*big.Int)(i))), nil
}

func BigPow(aValue, bValue int64) *big.Int {
	myResult := big.NewInt(aValue)
	return myResult.Exp(myResult, big.NewInt(bValue), nil)
}

func PaddedBigBytes(bigint *big.Int, n int) []byte {
	if bigint.BitLen()/8 >= n {
		return bigint.Bytes()
	}
	ret := make([]byte, n)
	ReadBits(bigint, ret)
	return ret
}

func bigEndianByteAt(bigint *big.Int, n int) byte {
	words := bigint.Bits()
	i := n / wordBytes
	if i >= len(words) {
		return byte(0)
	}
	word := words[i]
	shift := 8 * uint(n%wordBytes)

	return byte(word >> shift)
}

func Byte(bigint *big.Int, padlength, n int) byte {
	if n >= padlength {
		return byte(0)
	}
	return bigEndianByteAt(bigint, padlength-1-n)
}

func ReadBits(bigint *big.Int, buf []byte) {
	i := len(buf)
	for _, d := range bigint.Bits() {
		for j := 0; j < wordBytes && i > 0; j++ {
			i--
			buf[i] = byte(d)
			d >>= 8
		}
	}
}

func U256(x *big.Int) *big.Int {
	return x.And(x, tt256m1)
}

func S256(x *big.Int) *big.Int {
	if x.Cmp(tt255) < 0 {
		return x
	}
	return new(big.Int).Sub(x, tt256)
}

func Exp(base, exponent *big.Int) *big.Int {
	result := big.NewInt(1)

	for _, word := range exponent.Bits() {
		for i := 0; i < wordBits; i++ {
			if word&1 == 1 {
				U256(result.Mul(result, base))
			}
			U256(base.Mul(base, base))
			word >>= 1
		}
	}
	return result
}
