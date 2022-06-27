package es

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"reflect"
)

type Object map[string]interface{}

var (
	errUnknownType = errors.New("unknown body type")
)

func GetObject(body interface{}) (Object, error) {
	val := reflect.ValueOf(body)
	if val.Kind() == reflect.Map {
		switch doc := body.(type) {
		case map[string]interface{}:
			return doc, nil
		case Object:
			return doc, nil
		default:
			return nil, errUnknownType
		}
	}

	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() == reflect.Struct {
		doc := map[string]interface{}{}
		typ := val.Type()
		for i := 0; i < typ.NumField(); i++ {
			fieldType := typ.Field(i)
			tag := fieldType.Tag.Get("es")
			if tag == "" {
				continue
			}

			field := val.Field(i)
			if field.Kind() == reflect.Ptr {
				field = field.Elem()
			}
			if !field.IsValid() {
				continue
			}

			doc[tag] = field.Interface()
		}
		return doc, nil
	}

	return nil, errUnknownType
}

func GetObjectReader(body interface{}) (io.Reader, error) {
	doc, err := GetObject(body)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	err = json.NewEncoder(&buf).Encode(doc)
	if err != nil {
		return nil, err
	}
	return &buf, nil
}
