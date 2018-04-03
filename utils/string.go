package utils

import (
	"reflect"
	"strconv"
)

func StrToInt(str string) int {
	var i int
	var err error
	i, err = strconv.Atoi(str)
	if err != nil {
		i = 0
	}
	return i
}

func InArray(val interface{}, array interface{}) (exists bool, index int) {
	exists = false
	index = -1

	switch reflect.TypeOf(array).Kind() {
	case reflect.Slice:
		s := reflect.ValueOf(array)

		for i := 0; i < s.Len(); i++ {
			if reflect.DeepEqual(val, s.Index(i).Interface()) == true {
				index = i
				exists = true
				return
			}
		}
	}

	return
}
