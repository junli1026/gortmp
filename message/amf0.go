package message

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/junli1026/gortmp/logging"
	utils "github.com/junli1026/gortmp/utils"
	"math"
)

func deserializeAMF0(data []byte) ([]interface{}, error) {
	if len(data) == 0 {
		return nil, nil
	}
	arr := make([]interface{}, 0)
	i := 0
	for {
		l, o, err := readAMF0Value(data[i:])
		if err != nil {
			logging.Logger.Error(err)
			return arr, err
		}
		arr = append(arr, o)
		i += l
		if i >= len(data) {
			break
		}
	}
	return arr, nil
}

func readAMF0Value(data []byte) (int, interface{}, error) {
	if data == nil {
		return 0, nil, errors.New("nil input")
	}
	notSupported := fmt.Sprintf("AMF0 type %d is not supported", data[0])

	switch data[0] {
	case 0x00: //number
		return readAMF0Number(data)
	case 0x01:
		return readAMF0Boolean(data)
	case 0x02: //string
		return readAMF0String(data)
	case 0x03: //object-start
		return readAMF0Object(data)
	case 0x05:
		return readAMF0Null(data)
	case 0x08: //ecma array
		return readECMAArray(data)
	default:
		logging.Logger.Error(notSupported)
		return 0, nil, errors.New(notSupported)
	}
}

func readAMF0Null(data []byte) (int, interface{}, error) {
	return 1, nil, nil
}

func readAMF0Boolean(data []byte) (int, bool, error) {
	if len(data) < 2 {
		return 0, false, errors.New("invalid bool data")
	}
	return 2, data[1] != 0, nil
}

func readECMAArray(data []byte) (int, map[string]interface{}, error) {
	if data == nil {
		return 0, nil, errors.New("nil input")
	}
	if data[0] != 0x08 {
		return 0, nil, errors.New("ecma marker mismatch")
	}
	if len(data) < 5 {
		return 0, nil, errors.New("data length not enough for ecma array")
	}
	l := utils.ReadUint32(data[1:5])
	m := make(map[string]interface{})
	index := 5
	for i := 0; i < int(l); i++ {

		if index >= len(data) {
			return 0, nil, errors.New("broken ecma data")
		}
		if index+2 < len(data) &&
			data[index] == 0x00 &&
			data[index+1] == 0x00 &&
			data[index+2] == 0x09 {
			index += 3
			break
		}
		sz, key, value, err := readAMF0ObjectWithoutMarker(data[index:])
		if err != nil {
			return 0, nil, err
		}
		index += sz
		m[key] = value
	}

	if index+2 < len(data) &&
		data[index] == 0x00 &&
		data[index+1] == 0x00 &&
		data[index+2] == 0x09 {
		index += 3
	}
	return index, m, nil
}

func readAMF0Number(data []byte) (int, float64, error) {
	if data == nil {
		return 0, 0, errors.New("nil input")
	}
	if data[0] != 0x00 {
		return 0, 0, errors.New("number marker mismatch")
	}
	if len(data) < 9 {
		return 0, 0, errors.New("data length not enough for number")
	}
	n := utils.ReadFloat64(data[1:9])
	return 9, n, nil
}

func readAMF0StringWithoutMarker(data []byte) (int, string, error) {
	if data == nil {
		return 0, "", errors.New("nil input")
	}
	if len(data) < 2 {
		return 0, "", errors.New("data length not enough for string")
	}
	l := int(utils.ReadUint32(data[0:2]))
	if len(data) < 2+l {
		return 0, "", errors.New("data length not enough for string")
	}
	str := string(data[2 : 2+l])
	return 2 + l, str, nil
}

func readAMF0String(data []byte) (int, string, error) {
	if data == nil {
		return 0, "", errors.New("nil input")
	}
	if data[0] != 0x02 {
		return 0, "", errors.New("string marker mismatch")
	}
	l, str, err := readAMF0StringWithoutMarker(data[1:])
	if err != nil {
		return 0, "", err
	}
	return l + 1, str, nil
}

func readAMF0ObjectWithoutMarker(data []byte) (int, string, interface{}, error) {
	i := 0
	l, key, err := readAMF0StringWithoutMarker(data[i:])
	if err != nil {
		return 0, "", nil, err
	}
	i += l

	// check object-end marker
	if i < len(data) && data[i] == 0x09 {
		i++
		return i, "", nil, nil
	}

	if i >= len(data) {
		return i, "", nil, nil
	}

	l, val, err := readAMF0Value(data[i:])
	if err != nil {
		return i + l, "", nil, err
	}
	i += l

	return i, key, val, nil
}

func checkObjectEndMarker(data []byte, index int) (int, bool) {
	// check object-end marker
	if index < len(data) && data[index] == 0x09 {
		return index + 1, true
	}

	// check object-end marker
	if index+2 < len(data) && data[index] == 0x00 && data[index+1] == 0x00 && data[index+2] == 0x09 {
		return index + 3, true
	}
	return index, false
}

func readAMF0Object(data []byte) (int, map[string]interface{}, error) {
	if data == nil {
		return 0, nil, errors.New("nil input")
	}
	if data[0] != 0x03 {
		return 0, nil, errors.New("object marker mismatch")
	}
	m := make(map[string]interface{})
	i := 1
	for {
		if next, end := checkObjectEndMarker(data, i); end {
			i = next
			break
		}

		l, key, value, err := readAMF0ObjectWithoutMarker(data[i:])
		if err != nil {
			return 0, nil, err
		}
		i += l
		if key != "" && value != nil {
			m[key] = value
		}

		if next, end := checkObjectEndMarker(data, i); end {
			i = next
			break
		}

		if i >= len(data) {
			break
		}
	}
	return i, m, nil
}

func serializeAMF0(arr []interface{}) ([]byte, error) {
	data := make([]byte, 0)
	var d []byte
	var err error
	for _, o := range arr {
		d, err = buildAMF0Value(o)
		if err != nil {
			return nil, err
		}
		data = append(data, d[:]...)
	}
	return data, nil
}

func buildAMF0Value(o interface{}) (d []byte, err error) {
	switch v := o.(type) {
	case string:
		d, err = buildAMF0String(v)
	case float64:
		d, err = buildAMF0Number(float64(v))
	case int:
		d, err = buildAMF0Number(float64(v))
	case map[string]interface{}:
		d, err = buildAMF0Object(v)
	case nil:
		d, err = buildAMF0Null()
	default:
		logging.Logger.Warning("not supported type, ignore", v)
	}
	if err != nil {
		return nil, err
	}
	return
}

func buildAMF0String(str string) ([]byte, error) {
	if len(str) > 0xFFFF {
		return nil, errors.New("string too long")
	}
	l := uint16(len(str))
	data := make([]byte, 1+2)
	data[0] = 0x02
	binary.BigEndian.PutUint16(data[1:3], l)
	data = append(data, []byte(str)[:]...)
	return data, nil
}

func buildAMF0Null() ([]byte, error) {
	return []byte{0x05}, nil
}

func buildAMF0Number(n float64) ([]byte, error) {
	data := make([]byte, 9)
	data[0] = 0x00
	binary.BigEndian.PutUint64(data[1:9], math.Float64bits(n))
	return data, nil
}

func buildAMF0Object(m map[string]interface{}) ([]byte, error) {
	data := make([]byte, 1)
	data[0] = 0x03
	var tmp []byte
	var err error
	for k, v := range m {
		if tmp, err = buildAMF0String(k); err != nil {
			return nil, err
		}
		data = append(data, tmp[1:]...)

		if tmp, err = buildAMF0Value(v); err != nil {
			return nil, err
		}
		data = append(data, tmp[:]...)
	}
	data = append(data, 0x00, 0x00, 0x09) // append object-end marker
	return data, nil
}
