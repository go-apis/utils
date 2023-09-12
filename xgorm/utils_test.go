package xgorm

import (
	"context"
	"testing"
)

func TestRecreate(t *testing.T) {
	ctx := context.Background()
	cfg := &DbConfig{
		AwsRegion: "us-east-1",
		Username:  "noops",
		Password:  "mysecret",
		Host:      "localhost",
		Port:      5432,
		Database:  "config",
		SSLMode:   "disable",
	}

	if err := recreate(ctx, cfg); err != nil {
		t.Fatal(err)
	}
}
