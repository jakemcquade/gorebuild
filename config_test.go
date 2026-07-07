package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	content := "" +
		"server:\n" +
		"  port: 9000\n" +
		"projects:\n" +
		"  foo:\n" +
		"    path: /srv/foo\n" +
		"    branch: main\n" +
		"  bar:\n" +
		"    path: ~/bar\n"

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	if err := loadConfig(path); err != nil {
		t.Fatalf("loadConfig: %v", err)
	}

	if config.Server.Port != 9000 {
		t.Errorf("Server.Port = %d, want 9000", config.Server.Port)
	}
	if got := len(config.Projects); got != 2 {
		t.Fatalf("len(Projects) = %d, want 2", got)
	}
	if p := config.Projects["foo"]; p.Path != "/srv/foo" || p.Branch != "main" {
		t.Errorf(`Projects["foo"] = %+v, want {Path:/srv/foo Branch:main}`, p)
	}
	if p := config.Projects["bar"]; p.Path != "~/bar" || p.Branch != "" {
		t.Errorf(`Projects["bar"] = %+v, want {Path:~/bar Branch:}`, p)
	}
}

func TestLoadConfigMissingFile(t *testing.T) {
	if err := loadConfig(filepath.Join(t.TempDir(), "nope.yaml")); err == nil {
		t.Error("loadConfig(missing) = nil, want error")
	}
}
