package utils

import (
	"fmt"
	"reflect"
)

func FormatListOfStructs(vL ...interface{}) string {
	retStr := "[\n"
	for _, v := range vL {
		retStr += FormatStruct(v) + ",\n"
	}
	retStr += "]"

	return retStr
}

func FormatStruct(v interface{}) string {
	val := reflect.ValueOf(v)
	typ := val.Type()

	if typ.Kind() != reflect.Struct {
		return "Provided value is not a struct"
	}

	retStr := "{\n"
	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		value := val.Field(i)

		retStr += fmt.Sprintf("  %s: %v, \n", field.Name, value)
	}
	retStr += "}"

	return retStr
}
