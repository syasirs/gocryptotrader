package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/thrasher-corp/gocryptotrader/backtester/common"
	"github.com/thrasher-corp/gocryptotrader/common/file"
)

func TestLoadBacktesterConfig(t *testing.T) {
	t.Parallel()
	cfg, err := GenerateDefaultConfig()
	if err != nil {
		t.Error(err)
	}
	testConfig, err := json.Marshal(cfg)
	if err != nil {
		t.Error(err)
	}
	dir := t.TempDir()
	f := filepath.Join(dir, "test.config")
	err = file.Write(f, testConfig)
	if err != nil {
		t.Error(err)
	}
	_, err = ReadBacktesterConfigFromPath(f)
	if err != nil {
		t.Error(err)
	}

	_, err = ReadBacktesterConfigFromPath("test")
	if !errors.Is(err, common.ErrFileNotFound) {
		t.Errorf("received '%v' expected '%v'", err, common.ErrFileNotFound)
	}
}

func TestGenerateDefaultConfig(t *testing.T) {
	t.Parallel()
	cfg, err := GenerateDefaultConfig()
	if err != nil {
		t.Error(err)
	}
	wd, err := os.Getwd()
	if err != nil {
		t.Error(err)
	}
	if filepath.Join(wd, "examples", "ftx-cash-carry.strat") != cfg.SingleRunStrategyConfig {
		t.Errorf("expected '%v' received '%v'", filepath.Join(wd, "examples", "ftx-cash-carry.strat"), cfg.SingleRunStrategyConfig)
	}
}
