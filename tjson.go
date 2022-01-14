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
}

func GetFromType(key string) (interface{}, error) {
	if f, in := typeMap[key]; in {
		return f(), nil
	} else {
		return nil, fmt.Errorf("key %s not found", key)
	}
}

func NewFromTypedObject(data []byte) (interface{}, error) {
	type typedJsonObject struct {
		Type string `json:"type"`
	}

	typedObject := typedJsonObject{}
	if err := json.Unmarshal(data, &typedObject); err != nil {
		return nil, err
	}

	return GetFromType(typedObject.Type)
}

func ParseJsonTags(data []byte, x interface{}) error {
	t := reflect.TypeOf(x).Elem()
	V := reflect.ValueOf(x).Elem()
	n := t.NumField()

	raw := map[string]json.RawMessage{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	for i := 0; i < n; i++ {
		f := t.Field(i)
		jsonKey := strings.Split(f.Tag.Get(JSON), ",")[0]

		if r, in := raw[jsonKey]; in {

			v := reflect.New(f.Type).Interface()

			if err := json.Unmarshal(r, v); err != nil {
				return err
			}
			V.Field(i).Set(reflect.ValueOf(v).Elem())

		}

	}

	return nil
}

func unpack(b []byte, v reflect.Value) error {

	switch v.Type().Kind() {
	case reflect.Interface:
		return unpackInterface(b, v)
	case reflect.Slice:
		return unpackSlice(b, v)
	case reflect.Map:
		return unpackMap(b, v)
	}

	return nil
}

func unpackInterface(b []byte, v reflect.Value) error {
	x, err := NewFromTypedObject(b)

	if err != nil {
		return err
	}

	if err := json.Unmarshal(b, x); err != nil {
		return err
	}
	v.Set(reflect.ValueOf(x).Convert(v.Type()))
	return nil
}

func unpackSlice(b []byte, v reflect.Value) error {

	var c []json.RawMessage
	if err := json.Unmarshal(b, &c); err != nil {
		return err
	}

	v.Set(reflect.MakeSlice(v.Type(), len(c), len(c)))

	for i, raw := range c {
		if err := unpack(raw, v.Index(i)); err != nil {
			return err
		}
	}

	return nil
}

func unpackMap(b []byte, v reflect.Value) error {

	var c map[string]json.RawMessage
	if err := json.Unmarshal(b, &c); err != nil {
		return err
	}

	v.Set(reflect.MakeMapWithSize(v.Type(), 0))

	for rawKey, rawValue := range c {
		key := reflect.New(v.Type().Key()).Interface()
		if err := json.Unmarshal([]byte(rawKey), &key); err != nil {
			return err
		}

		u := reflect.New(v.Type().Elem()).Elem()
		unpack(rawValue, u)

		v.SetMapIndex(reflect.ValueOf(key).Elem(), u)
	}

	return nil
}

func ParseTJsonTags(data []byte, x interface{}) error {

	t := reflect.TypeOf(x).Elem()
	V := reflect.ValueOf(x).Elem()
	n := t.NumField()

	raw := map[string]json.RawMessage{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	for i := 0; i < n; i++ {
		f := t.Field(i)
		jsonKey := strings.Split(f.Tag.Get(TJSON), ",")[0]

		if r, in := raw[jsonKey]; in {
			if err := unpack(r, V.Field(i)); err != nil {
				return err
			}
		}

	}

	return nil
}
