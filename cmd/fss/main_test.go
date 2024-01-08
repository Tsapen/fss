package main

import (
	"testing"

	fsstest "github.com/Tsapen/fss/cmd/fss/fss-test"
	"github.com/Tsapen/fss/internal/config"
)

func TestFSS(t *testing.T) {
	clientCfg, err := config.GetForClient()
	if err != nil {
		t.Fatalf("read client configs: %v\n", err)
	}

	fssCfg, err := config.GetForFSS()
	if err != nil {
		t.Fatalf("read client configs: %v\n", err)
	}

	fsstest.TestFSS(t, clientCfg.Address, fssCfg)
}
