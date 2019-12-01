package message

import (
	"testing"
)

func Test_amf0(t *testing.T) {
	data := []interface{}{
		"connect",
		float64(1.0),
		map[string]interface{}{
			"name": "helloworld",
			"age":  float64(12),
		},
	}

	payload, err := serializeAMF0(data)
	if err != nil {
		t.Fail()
	}

	arr, err := deserializeAMF0(payload)
	if len(arr) != len(data) || err != nil {
		t.Fail()
	}

	for _, o := range arr {
		switch v := o.(type) {
		case string:
			if v != "connect" {
				t.Fail()
			}
		case float64:
			if v != 1.0 {
				t.Fail()
			}
		case map[string]interface{}:
			if v["name"] != "helloworld" {
				t.Fail()
			}
			if v["age"] != float64(12) {
				t.Fail()
			}
		}
	}

	/*
		m := arr[2].(map[string]interface{})
		if m == nil {
			t.Fail()
		}
		if m["name"] != "helloworld" {
			t.Fail()
		}
		if m["age"] != float64(12) {
			t.Fail()
		}
	*/
}
