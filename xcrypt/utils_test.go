package xcrypt

import (
	"testing"
	"time"
)

type testStruct struct {
	A         string    `json:"a"`
	B         int       `json:"b"`
	Timestamp time.Time `json:"timestamp"`
}

func Test_All(t *testing.T) {
	c, err := NewCrypt([]byte("1234567890123456"))
	if err != nil {
		t.Fatal(err)
		return
	}

	t.Run("should encrypt a struct", func(t *testing.T) {
		data := testStruct{
			A:         "Hello",
			B:         123,
			Timestamp: time.Now(),
		}

		encrypted, err := EncryptJson(c, data)
		if err != nil {
			t.Fatal(err)
		}

		if len(encrypted) == 0 {
			t.Fatal("Expected encrypted data to have length > 0")
		}
	})

	t.Run("should decrypt a struct", func(t *testing.T) {
		encrypted := "/FWy/+HBNfP0mfSPHuTgMUSEaapVlgyAbRnM3FUgHq9EG6p3OTIdeTG2LwDr37jihd58kl+u46/vhKDe5UalujaXknwNHT6ATW9dpjdmjCQQLgBSnOndtVuyRwsh2YtFF/ri"

		var data testStruct
		err := DecryptJson(c, encrypted, &data)
		if err != nil {
			t.Fatal(err)
		}
	})

}
