// internal/client/client.go
package client

import (
	"fmt"

	"github.com/emersion/go-imap/v2/imapclient"
	"github.com/vearne/imap-cli/internal/config"
)

func Dial(cfg *config.Config) (*imapclient.Client, error) {
	var c *imapclient.Client
	var err error
	switch cfg.TLSMode {
	case config.TLSTLS:
		c, err = imapclient.DialTLS(cfg.Addr(), nil)
	case config.TLSStartTLS:
		c, err = imapclient.DialStartTLS(cfg.Addr(), nil)
	case config.TLSNone:
		c, err = imapclient.DialInsecure(cfg.Addr(), nil)
	default:
		return nil, fmt.Errorf("unknown TLS mode %q", cfg.TLSMode)
	}
	if err != nil {
		return nil, fmt.Errorf("connect %s: %w", cfg.Addr(), err)
	}
	if err := c.Login(cfg.Username, cfg.Password).Wait(); err != nil {
		c.Logout().Wait() //nolint:errcheck
		return nil, fmt.Errorf("login as %q failed: %w\nHints: QQ/163 mailboxes need an authorization code (授权码), not the web login password; Gmail needs an app password with IMAP enabled in settings", cfg.Username, err)
	}
	return c, nil
}

// Run dials, logs in, executes fn with the logged-in client, then logs out.
func Run(cfg *config.Config, fn func(c *imapclient.Client) error) error {
	c, err := Dial(cfg)
	if err != nil {
		return err
	}
	fnErr := fn(c)
	if err := c.Logout().Wait(); err != nil && fnErr == nil {
		return fmt.Errorf("logout: %w", err)
	}
	return fnErr
}
