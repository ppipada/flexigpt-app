package reqresp

import (
	"bytes"
	"encoding/json"
	"errors"
	"reflect"
)

func jsonEqual(a, b json.RawMessage) bool {
	var o1 any
	var o2 any

	if err := json.Unmarshal(a, &o1); err != nil {
		return false
	}
	if err := json.Unmarshal(b, &o2); err != nil {
		return false
	}
	// Direct reflect Deepequal would have issues when there are pointers, keyorders etc.
	// unmarshalling into a interface and then doing deepequal removes those issues.
	return reflect.DeepEqual(o1, o2)
}

func jsonStringsEqual(a, b string) bool {
	return jsonEqual([]byte(a), []byte(b))
}

func getJSONStrings(args ...any) ([]string, error) {
	ret := make([]string, 0, len(args))
	for _, a := range args {
		jsonBytes, err := json.Marshal(a)
		if err != nil {
			// Log.Printf("%s", err.Error()).
			return nil, err
		}
		ret = append(ret, string(jsonBytes))
	}
	return ret, nil
}

func jsonStructEqual(arg1, arg2 any) (bool, error) {
	vals, err := getJSONStrings(arg1, arg2)
	if err != nil {
		return false, errors.New("could not encode struct to json")
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
		if !a[i].ID.Equal(&b[i].ID) {
			return false
		}
		if !bytes.Equal(a[i].Params, b[i].Params) {
			return false
		}
	}
	return true
}

func compareUnionRequestSlices(a, b []UnionRequest) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].JSONRPC != b[i].JSONRPC || *a[i].Method != *b[i].Method {
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

func stringToPointer(s string) *string {
	return &s
}
