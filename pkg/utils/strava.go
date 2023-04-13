package utils

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"os"
	"strconv"
	"strings"
	"syscall"
	"unsafe"
)

func StringToUint64(string2 string) uint64 {
	u, _ := strconv.ParseUint(string2, 10, 64)
	return u
}

func Strava(value interface{}) string {
	var key string
	if value == nil {
		return key
	}
	switch value.(type) {
	case float64:
		ft := value.(float64)
		key = strconv.FormatFloat(ft, 'f', -1, 64)
	case float32:
		ft := value.(float32)
		key = strconv.FormatFloat(float64(ft), 'f', -1, 64)
	case int:
		it := value.(int)
		key = strconv.Itoa(it)
	case uint:
		it := value.(uint)
		key = strconv.Itoa(int(it))
	case int8:
		it := value.(int8)
		key = strconv.Itoa(int(it))
	case uint8:
		it := value.(uint8)
		key = strconv.Itoa(int(it))
	case int16:
		it := value.(int16)
		key = strconv.Itoa(int(it))
	case uint16:
		it := value.(uint16)
		key = strconv.Itoa(int(it))
	case int32:
		it := value.(int32)
		key = strconv.Itoa(int(it))
	case uint32:
		it := value.(uint32)
		key = strconv.Itoa(int(it))
	case int64:
		it := value.(int64)
		key = strconv.FormatInt(it, 10)
	case uint64:
		it := value.(uint64)
		key = strconv.FormatUint(it, 10)
	case string:
		key = value.(string)
	case []byte:
		key = string(value.([]byte))
	case [16]uint8:
		// uuid
		arr := value.([16]uint8)
		uuid, _ := uuid.FromBytes(arr[:])
		key = uuid.String()
	default:
		newValue, _ := json.Marshal(value)
		key = string(newValue)
	}

	return key
}

func RemoveRepeatedElement(arr []string) (newArr []string) {
	newArr = make([]string, 0)
	for i := 0; i < len(arr); i++ {
		repeat := false
		for j := i + 1; j < len(arr); j++ {
			if arr[i] == arr[j] {
				repeat = true
				break
			}
		}
		if !repeat {
			newArr = append(newArr, arr[i])
		}
	}
	return
}

func GenerateString(leftVar, middleVar, rightVar string) string {
	totalLength, err := getTerminalWidth()
	if err != nil {
		return ""
	}
	middleLength := totalLength - 10 - len(leftVar) - len(rightVar)
	if middleLength <= 0 {
		middleLength = 5
	}

	middlePart := strings.Repeat(middleVar, middleLength)

	result := fmt.Sprintf("%s%s%s", leftVar, middlePart, rightVar)

	return result
}

func getTerminalWidth() (int, error) {
	var size [4]uint16
	_, _, err := syscall.Syscall(syscall.SYS_IOCTL, os.Stdout.Fd(), uintptr(syscall.TIOCGWINSZ), uintptr(unsafe.Pointer(&size)))
	if err != 0 {
		return 0, fmt.Errorf("failed to get terminal width: %v", err)
	}
	return int(size[1]), nil
}
