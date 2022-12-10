package Utilities

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
)

func Contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}

func FileExists(fileName string) bool {
	if _, err := os.Stat(fileName); err == nil {
		return true

	} else if errors.Is(err, os.ErrNotExist) {
		// path/to/whatever does *not* exist
		return false

	} else {
		// Schrodinger: file may or may not exist. See err for details.
		// Therefore, do *NOT* use !os.IsNotExist(err) to test for file existence
		return false
	}
}

func ConvertHexToInt(hex string) int64 {
	// replace 0x or 0X with empty String
	hex = strings.Replace(hex, "0x", "", -1)
	hex = strings.Replace(hex, "0X", "", -1)

	deviceIndex, _ := strconv.ParseInt(hex, 16, 64)
	return deviceIndex
}

func Merge(a, b interface{}) {
	ra := reflect.ValueOf(a).Elem()
	rb := reflect.ValueOf(b).Elem()

	numFields := ra.NumField()

	for i := 0; i < numFields; i++ {
		fieldA := ra.Field(i)
		fieldB := rb.Field(i)

		switch fieldA.Kind() {
		case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
			if fieldA.IsNil() {
				fieldA.Set(fieldB)
			}
		}
	}
}

// GetFields Converts struct slice to string slice
func GetFields(i interface{}) (res []string) {
	v := reflect.ValueOf(i)
	for j := 0; j < v.NumField(); j++ {
		res = append(res, v.Field(j).String())
	}
	return
}

// Invoke any method of struct or interface
//func Invoke(any interface{}, methodName string) {
//reflect.ValueOf(any).MethodByName(methodName).Call([]reflect.Value{})
/*st := reflect.TypeOf(any)
_, ok := st.MethodByName(methodName)
if ok {
	reflect.ValueOf(any).MethodByName(methodName).Call([]reflect.Value{})
}*/
//}

// Invoke any method of struct or interface
func Invoke(any interface{}, methodName string, args ...interface{}) (reflect.Value, error) {
	method := reflect.ValueOf(any).MethodByName(methodName)
	methodType := method.Type()
	numIn := methodType.NumIn()
	if numIn > len(args) {
		return reflect.ValueOf(nil), fmt.Errorf("Method %s must have minimum %d params. Have %d", methodName, numIn, len(args))
	}
	if numIn != len(args) && !methodType.IsVariadic() {
		return reflect.ValueOf(nil), fmt.Errorf("Method %s must have %d params. Have %d", methodName, numIn, len(args))
	}
	in := make([]reflect.Value, len(args))
	for i := 0; i < len(args); i++ {
		var inType reflect.Type
		if methodType.IsVariadic() && i >= numIn-1 {
			inType = methodType.In(numIn - 1).Elem()
		} else {
			inType = methodType.In(i)
		}
		argValue := reflect.ValueOf(args[i])
		if !argValue.IsValid() {
			return reflect.ValueOf(nil), fmt.Errorf("Method %s. Param[%d] must be %s. Have %s", methodName, i, inType, argValue.String())
		}
		argType := argValue.Type()
		if argType.ConvertibleTo(inType) {
			in[i] = argValue.Convert(inType)
		} else {
			return reflect.ValueOf(nil), fmt.Errorf("Method %s. Param[%d] must be %s. Have %s", methodName, i, inType, argType)
		}
	}
	return method.Call(in)[0], nil
}

// UnmarshalJSONTuple unmarshal JSON list (tuple) into a struct.
func UnmarshalJSONTuple(text []byte, obj interface{}) (err error) {
	var list []json.RawMessage
	err = json.Unmarshal(text, &list)
	if err != nil {
		return
	}

	objValue := reflect.ValueOf(obj).Elem()
	if len(list) > objValue.Type().NumField() {
		return fmt.Errorf("tuple has too many fields (%v) for %v",
			len(list), objValue.Type().Name())
	}

	for i, elemText := range list {
		err = json.Unmarshal(elemText, objValue.Field(i).Addr().Interface())
		if err != nil {
			return
		}
	}
	return
}
