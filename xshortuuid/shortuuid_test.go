package xshortuuid

import (
	"testing"

	"github.com/google/uuid"
)

func TestAll(t *testing.T) {
	data := []struct {
		uuid  uuid.UUID
		short string
	}{
		{uuid.MustParse("00000000-0000-0000-0000-000000000000"), "2222222222222222222222"},
		{uuid.MustParse("5f951342-039b-4a95-9724-4e1759c457b7"), "K2MsPjU6bVQEufwwSUoytg"},
		{uuid.MustParse("db9f430e-3528-4f89-82b9-7dc1ed8e7a0b"), "h6K36Q8gqEVHEdZUbb8rsU"},
		{uuid.MustParse("6e96b709-b2be-44e5-b1b1-2257b5a69fe1"), "MgYZe5gZkGeC4A8GHGGhZc"},
	}

	for _, d := range data {
		t.Run(d.short, func(t *testing.T) {
			short := Encode(d.uuid)
			if short != d.short {
				t.Errorf("Encode(%s) = %s, want %s", d.uuid, short, d.short)
			}
			u, err := Decode(short)
			if err != nil {
				t.Errorf("Decode(%s) = %v, want nil", short, err)
			}
			if u != d.uuid {
				t.Errorf("Decode(%s) = %s, want %s", short, u, d.uuid)
			}
		})
	}
}
