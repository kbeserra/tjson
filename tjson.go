package tjson

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

const (
	JSON  = "json"
	TJSON = "tjson"
)

var typeMap map[string]func() interface{}

func RegisterType(key string, value func() interface{}) {
	if typeMap == nil {
		typeMap = make(map[string]func() interface{})
	}

	typeMap[key] = value

	fmt.Println(typeMap)
}

func GetFromType(key string) (interface{}, error) {
	if f, in := typeMap[key]; in {
		return f(), nil
	} else {
		return nil, fmt.Errorf("key %s not found", key)
	}
}

func newFromTypedObject(data []byte) (interface{}, error) {
	type typedJsonObject struct {
		Type string `json:"type"`
	}

	typedObject := typedJsonObject{}
	if err := json.Unmarshal(data, &typedObject); err != nil {
		return nil, err
	}

	return GetFromType(typedObject.Type)
}

func parseHelper(data []byte, x interface{}, tag string,
	new func([]byte, reflect.Type) (interface{}, error),
	convert func(interface{}, reflect.Type) reflect.Value,
) error {

	t := reflect.TypeOf(x).Elem()
	V := reflect.ValueOf(x).Elem()
	n := t.NumField()

	raw := map[string]json.RawMessage{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	for i := 0; i < n; i++ {
		f := t.Field(i)
		jsonKey := strings.Split(f.Tag.Get(tag), ",")[0]

		if r, in := raw[jsonKey]; in {

			v, err := new(r, f.Type)
			if err != nil {
				return err
			}

			if err := json.Unmarshal(r, v); err != nil {
				return err
			}
			V.Field(i).Set(convert(v, f.Type))
		}

	}

	return nil
}

func ParseJsonTags(data []byte, x interface{}) error {
	return parseHelper(data, x, JSON,
		func(b []byte, t reflect.Type) (interface{}, error) { return reflect.New(t).Interface(), nil },
		func(v interface{}, t reflect.Type) reflect.Value { return reflect.ValueOf(v).Elem() },
	)
}

func ParseTJsonTags(data []byte, x interface{}) error {
	return parseHelper(data, x, TJSON,
		func(b []byte, t reflect.Type) (interface{}, error) { return newFromTypedObject(b) },
		func(v interface{}, t reflect.Type) reflect.Value { return reflect.ValueOf(v).Convert(t) },
	)
}
