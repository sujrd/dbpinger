package main

import (
	"testing"
  "gopkg.in/gcfg.v1"
)


func TestConfigLoad(t *testing.T) {

	var cfg Config
  confFile := "dbpinger.conf"

  err := gcfg.ReadFileInto(&cfg, confFile)

	if err != nil {
		t.Errorf("Failed to load config file %s", confFile)
	}

	if cfg.Main.Listen != "4146" {
		t.Errorf("Listen port %s is not 4146", cfg.Main.Listen)
	}

	if cfg.Main.DBHost != "localhost" {
		t.Errorf("DB Host %s is not localhost", cfg.Main.DBHost)
	}

	if cfg.Main.DBPort != "3306" {
		t.Errorf("DB Port %s is not localhost", cfg.Main.DBPort)
	}

	if cfg.Main.DBUser != "debian-sys-maint" {
		t.Errorf("DB Host %s is not debian-sys-maint", cfg.Main.DBUser)
	}

	if cfg.Main.DBPass != "0&542%" {
		t.Errorf("DB Pass %s does not match ", cfg.Main.DBPass)
	}

}
