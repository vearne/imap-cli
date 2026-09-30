package config

import "testing"

func setEnv(t *testing.T, kv map[string]string) {
	t.Helper()
	for _, k := range []string{"IMAP_HOST", "IMAP_PORT", "IMAP_USERNAME", "IMAP_PASSWORD", "IMAP_TLS"} {
		t.Setenv(k, "")
	}
	for k, v := range kv {
		t.Setenv(k, v)
	}
}

func TestLoadDefaults(t *testing.T) {
	setEnv(t, map[string]string{"IMAP_HOST": "imap.example.com", "IMAP_USERNAME": "bob", "IMAP_PASSWORD": "secret"})
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() = %v", err)
	}
	if cfg.Port != 993 || cfg.TLSMode != TLSTLS || cfg.Addr() != "imap.example.com:993" {
		t.Fatalf("defaults wrong: %+v", cfg)
	}
}

func TestLoadMissingRequired(t *testing.T) {
	setEnv(t, map[string]string{"IMAP_HOST": "imap.example.com", "IMAP_USERNAME": "bob"}) // no password
	if _, err := Load(); err == nil || err.Error() != "required environment variable IMAP_PASSWORD is not set" {
		t.Fatalf("Load() err = %v, want missing IMAP_PASSWORD", err)
	}
}

func TestLoadInvalidTLSMode(t *testing.T) {
	setEnv(t, map[string]string{"IMAP_HOST": "h", "IMAP_USERNAME": "u", "IMAP_PASSWORD": "p", "IMAP_TLS": "ssl3"})
	if _, err := Load(); err == nil {
		t.Fatal("Load() should reject unknown IMAP_TLS value")
	}
}
