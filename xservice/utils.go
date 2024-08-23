package xservice

import (
	"os"
	"reflect"

	"github.com/mitchellh/mapstructure"
)

func StringExpandEnv() mapstructure.DecodeHookFuncKind {
	return func(
		f reflect.Kind,
		t reflect.Kind,
		data interface{}) (interface{}, error) {
		if f != reflect.String || t != reflect.String {
			return data, nil
		}

		return os.ExpandEnv(data.(string)), nil
	}
}
