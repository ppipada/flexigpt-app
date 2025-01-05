package jsonrpc

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"reflect"
)

func jsonEqual(a, b json.RawMessage) bool {
	var o1 interface{}
	var o2 interface{}

	if err := json.Unmarshal(a, &o1); err != nil {
		return false
	}
	if err := json.Unmarshal(b, &o2); err != nil {
		return false
	}
	// Direct reflect Deepequal would have issues when there are pointers, keyorders etc.
	// unmarshalling into a interface and then doing deepequal removes those issues
	return reflect.DeepEqual(o1, o2)
}

func jsonStringsEqual(a, b string) bool {
	return jsonEqual([]byte(a), []byte(b))
}

func getJSONStrings(args ...interface{}) ([]string, error) {
	var ret []string
	for _, a := range args {
		jsonBytes, err := json.Marshal(a)
		if err != nil {
			return nil, err
		}
		ret = append(ret, string(jsonBytes))
	}
	return ret, nil
}

func jsonStructEqual(arg1 interface{}, arg2 interface{}) (bool, error) {
	vals, err := getJSONStrings(arg1, arg2)
	if err != nil {
		log.Fatalf("Could not encode struct to json")
	}
	return jsonStringsEqual(vals[0], vals[1]), nil
}

func compareRequestSlices(a, b []Request[json.RawMessage]) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].JSONRPC != b[i].JSONRPC || a[i].Method != b[i].Method {
			return false
		}
		if !a[i].ID.Equal(b[i].ID) {
			return false
		}
		if !bytes.Equal(a[i].Params, b[i].Params) {
			return false
		}
	}
	return true
}

func arraysAreSimilar(arr1, arr2 []int) bool {
	if len(arr1) != len(arr2) {
		return false
	}
	if len(arr1) != 0 {
		counts1 := make(map[int]int)
		counts2 := make(map[int]int)

		for _, num := range arr1 {
			counts1[num]++
		}

		for _, num := range arr2 {
			counts2[num]++
		}

		for key, count1 := range counts1 {
			if count2, exists := counts2[key]; !exists || count1 != count2 {
				return false
			}
		}
	}

	return true
}

// AddParams defines the parameters for the "add" method
type AddParams struct {
	A int `json:"a"`
	B int `json:"b"`
}

type AddResult struct {
	Sum int `json:"sum"`
}

type NotifyParams struct {
	Message string `json:"message"`
}

// ConcatParams defines the parameters for the "concat" method
type ConcatParams struct {
	S1 string `json:"s1"`
	S2 string `json:"s2"`
}

// PingParams defines the parameters for the "ping" notification
type PingParams struct {
	Message string `json:"message"`
}

// AddEndpoint is the handler for the "add" method
func AddEndpoint(ctx context.Context, params AddParams) (AddResult, error) {
	res := params.A + params.B
	return AddResult{Sum: res}, nil
}

// ConcatEndpoint is the handler for the "concat" method
func ConcatEndpoint(ctx context.Context, params ConcatParams) (string, error) {
	return params.S1 + params.S2, nil
}

// PingEndpoint is the handler for the "ping" notification
func PingEndpoint(ctx context.Context, params PingParams) error {
	return nil
}

func NotifyEndpoint(ctx context.Context, params NotifyParams) error {
	// Process notification
	return nil
}
