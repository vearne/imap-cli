package config

import (
	"fmt"
	"os"
	"strconv"
)

const (
	TLSTLS      = "tls"
	TLSStartTLS = "starttls"
	TLSNone     = "none"
)

type Config struct {
	Host     string
	Port     int
	Username string
	Password string
	TLSMode  string
}

func (c Config) Addr() string { return fmt.Sprintf("%s:%d", c.Host, c.Port) }

func Load() (Config, error) {
	cfg := Config{Port: 993, TLSMode: TLSTLS}
	var err error
	if cfg.Host, err = required("IMAP_HOST"); err != nil {
		return cfg, err
	}
	if cfg.Username, err = required("IMAP_USERNAME"); err != nil {
		return cfg, err
	}
	if cfg.Password, err = required("IMAP_PASSWORD"); err != nil {
		return cfg, err
	}
	if v := os.Getenv("IMAP_PORT"); v != "" {
		p, err := strconv.Atoi(v)
		if err != nil || p < 1 || p > 65535 {
			return cfg, fmt.Errorf("IMAP_PORT %q is not a valid port number", v)
		}
		cfg.Port = p
	}
	if v := os.Getenv("IMAP_TLS"); v != "" {
		switch v {
		case TLSTLS, TLSStartTLS, TLSNone:
			cfg.TLSMode = v
		default:
			return cfg, fmt.Errorf("IMAP_TLS must be one of tls, starttls, none; got %q", v)
		}
	}
	return cfg, nil
}

func required(name string) (string, error) {
	v := os.Getenv(name)
	if v == "" {
		return "", fmt.Errorf("required environment variable %s is not set", name)
	}
	return v, nil
}
