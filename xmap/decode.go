package xmap

import (
	"reflect"

	"github.com/google/uuid"
	"github.com/mitchellh/mapstructure"
)

type DecodeOption func(*mapstructure.DecoderConfig)

func WithTagName(tagName string) DecodeOption {
	return func(config *mapstructure.DecoderConfig) {
		config.TagName = tagName
	}
}

func StringToUUIDHookFunc() mapstructure.DecodeHookFunc {
	return func(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
		if f.Kind() != reflect.String {
			return data, nil
		}
		if t != reflect.TypeOf(uuid.UUID{}) {
			return data, nil
		}

		return uuid.Parse(data.(string))
	}
}

func Decode(input interface{}, output interface{}, options ...DecodeOption) error {
	config := &mapstructure.DecoderConfig{
		WeaklyTypedInput: true,
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			StringToUUIDHookFunc(),
		),
		Result: output,
	}

	for _, option := range options {
		option(config)
	}

	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return err
	}

	return decoder.Decode(input)
}
