package contract

import (
	"github.com/oldercn/restclient-go-sdk/mychain/mychain-sdk-go/common/codec/contract/abi"
)

const (
	MAX_N = 32
	MIN_N = 1
)

type ContractParameters struct {
	MethodSignature string
	InputParameters []interface{}
	ABI             *abi.ABI
}

//func (cp ContractParameters) GetEncodedBytes() ([]byte, error) {
//	if cp.ABI == nil {
//		return nil, errors.New("Hash not ABI information")
//	}
//	return cp.GetEncodedData()
//}

//func (cp ContractParameters) GetEncodedData() ([]byte, error) {
//	if cp.MethodSignature == "" {
//		// constructor
//		return cp.ABI.Pack("", cp.InputParameters[:]...)
//	} else {
//		return cp.ABI.Pack(cp.MethodSignature, cp.InputParameters[:]...)
//	}
//}
