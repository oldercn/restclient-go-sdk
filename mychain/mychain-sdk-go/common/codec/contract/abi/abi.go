package abi

import (
	"encoding/json"
	"fmt"
	"io"
)

type ABI struct {
	Constructor Method
	Methods     map[string]Method
	Events      map[string]Event
}

//func (abi ABI) UnpackIntoMap(valueMap map[string]interface{}, typeName string, originalData []byte) (err error) {
//	if len(originalData) == 0 {
//		return fmt.Errorf("abi: unmarshalling empty output")
//	}
//	if method, ok := abi.Methods[typeName]; ok {
//		if len(originalData)%32 != 0 {
//			return fmt.Errorf("abi: improperly formatted output")
//		}
//		return method.Outputs.UnpackIntoMap(valueMap, originalData)
//	}
//	if event, ok := abi.Events[typeName]; ok {
//		return event.Inputs.UnpackIntoMap(valueMap, originalData)
//	}
//	return fmt.Errorf("abi: could not locate named method or event")
//}

//func (abi ABI) Pack(typeName string, args ...interface{}) ([]byte, error) {
//	if typeName == "" {
//		arguments, err := abi.Constructor.Inputs.Pack(args...)
//		if err != nil {
//			return nil, err
//		}
//		return arguments, nil
//	}
//	method, exist := abi.Methods[typeName]
//	if !exist {
//		return nil, fmt.Errorf("method '%s' not found", typeName)
//	}
//	arguments, err := method.Inputs.Pack(args...)
//	if err != nil {
//		return nil, err
//	}
//	return append(method.Id(), arguments...), nil
//}
//
//func (abi *ABI) MethodById(sigdata []byte) (*Method, error) {
//	if len(sigdata) < 4 {
//		return nil, fmt.Errorf("data too short (%d bytes) for abi method lookup", len(sigdata))
//	}
//	for _, method := range abi.Methods {
//		if bytes.Equal(method.Id(), sigdata[:4]) {
//			return &method, nil
//		}
//	}
//	return nil, fmt.Errorf("no method with id: %#x", sigdata[:4])
//}


func JSON(reader io.Reader) (ABI, error) {
	dec := json.NewDecoder(reader)

	var abi ABI
	if err := dec.Decode(&abi); err != nil {
		return ABI{}, err
	}

	return abi, nil
}

func (abi ABI) Unpack(v interface{}, name string, data []byte) (err error) {
	if len(data) == 0 {
		return fmt.Errorf("abi: unmarshalling empty output")
	}
	if method, ok := abi.Methods[name]; ok {
		if len(data)%32 != 0 {
			return fmt.Errorf("abi: improperly formatted output: %s - Bytes: [%+v]", string(data), data)
		}
		return method.Outputs.Unpack(v, data)
	}
	if event, ok := abi.Events[name]; ok {
		return event.Inputs.Unpack(v, data)
	}
	return fmt.Errorf("abi: could not locate named method or event")
}

func (abi *ABI) UnmarshalJSON(originalData []byte) error {
	var fields []struct {
		Type      string
		Name      string
		Constant  bool
		Anonymous bool
		Inputs    []Argument
		Outputs   []Argument
	}

	if err := json.Unmarshal(originalData, &fields); err != nil {
		return err
	}

	abi.Methods = make(map[string]Method)
	abi.Events = make(map[string]Event)

	for _, field := range fields {
		switch field.Type {
		case "function", "":
			abi.Methods[field.Name] = Method{
				Name:    field.Name,
				Const:   field.Constant,
				Inputs:  field.Inputs,
				Outputs: field.Outputs,
			}
		case "constructor":
			abi.Constructor = Method{
				Inputs: field.Inputs,
			}
		case "event":
			abi.Events[field.Name] = Event{
				Name:      field.Name,
				Anonymous: field.Anonymous,
				Inputs:    field.Inputs,
			}
		}
	}

	return nil
}