package Utilities

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"io"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unicode"
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

func DrawRect(imgOrig image.Image, boundingBoxes [][]int, thickness int, color color.Color) draw.Image {
	// convert as usable image
	b := imgOrig.Bounds()
	img := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	draw.Draw(img, img.Bounds(), imgOrig, b.Min, draw.Src)

	for _, singleBoundingBox := range boundingBoxes {
		x1, y1, x2, y2 := singleBoundingBox[0], singleBoundingBox[1], singleBoundingBox[2], singleBoundingBox[3]

		for t := 0; t < thickness; t++ {
			// draw horizontal lines
			for x := x1; x <= x2; x++ {
				img.Set(x, y1+t, color)
				img.Set(x, y2-t, color)
			}
			// draw vertical lines
			for y := y1; y <= y2; y++ {
				img.Set(x1+t, y, color)
				img.Set(x2-t, y, color)
			}
		}
	}

	return img
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

func WriteLog(logFile string, logData string) {
	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println(err)
	}
	defer f.Close()

	// write new log line to file with time at the start
	if _, err := f.WriteString(time.Now().Format("2006-01-02 15:04:05") + " - " + logData + "\n"); err != nil {
		fmt.Println(err)
	}
}

func KillProcessById(pid int) error {
	if pid <= 0 {
		return errors.New("pid must be greater than 0")
	}
	process, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	return process.Kill()
}

func CamelToSnake(s string) string {
	var result bytes.Buffer
	var prevChar rune

	for _, char := range s {
		if unicode.IsUpper(char) {
			// For the first character, we don't want to prepend an underscore
			if result.Len() > 0 && (unicode.IsLower(prevChar) || (prevChar != 0 && unicode.IsUpper(prevChar) && (len([]rune(s)) > result.Len() && unicode.IsLower([]rune(s)[result.Len()])))) {
				result.WriteRune('_')
			}
			char = unicode.ToLower(char)
		}
		result.WriteRune(char)
		prevChar = char
	}

	return result.String()
}

func FileHash(file io.Reader) (string, error) {
	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", fmt.Errorf("failed to compute hash: %w", err)
	}

	calculatedHash := hasher.Sum(nil)
	calculatedHashStr := hex.EncodeToString(calculatedHash)
	return calculatedHashStr, nil
}

func Capitalize(str string) string {
	runes := []rune(str)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}
