package xshortuuid

import "github.com/google/uuid"

// DefaultEncoder is the default encoder uses when generating new UUIDs, and is
// based on Base57.
var DefaultEncoder = &encoder{newAlphabet(DefaultAlphabet)}

func Encode(u uuid.UUID) string {
	return DefaultEncoder.Encode(u)
}

func Decode(s string) (uuid.UUID, error) {
	return DefaultEncoder.Decode(s)
}

func NewEncoder(alphabet string) Encoder {
	return &encoder{newAlphabet(alphabet)}
}
