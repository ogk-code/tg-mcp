package client

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/tg"

	"tg-mcp/storage"
)

type Client struct {
	client  *telegram.Client
	api     *tg.Client
	sender  *message.Sender
	storage *storage.FileStorage
	appID   int
	appHash string

	mu         sync.RWMutex
	running    bool
	authorized bool
	phone      string
	codeHash   string
}

type Config struct {
	AppID       int
	AppHash     string
	SessionFile string
}

func ConfigFromEnv() (*Config, error) {
	appIDStr := os.Getenv("TG_APP_ID")
	if appIDStr == "" {
		return nil, fmt.Errorf("TG_APP_ID environment variable is required")
	}
	appID, err := strconv.Atoi(appIDStr)
	if err != nil {
		return nil, fmt.Errorf("TG_APP_ID must be a number: %w", err)
	}

	appHash := os.Getenv("TG_APP_HASH")
	if appHash == "" {
		return nil, fmt.Errorf("TG_APP_HASH environment variable is required")
	}

	sessionFile := os.Getenv("TG_SESSION_FILE")

	return &Config{
		AppID:       appID,
		AppHash:     appHash,
		SessionFile: sessionFile,
	}, nil
}

func New(cfg *Config) *Client {
	sessionStorage := storage.NewFileStorage(cfg.SessionFile)

	tgClient := telegram.NewClient(cfg.AppID, cfg.AppHash, telegram.Options{
		SessionStorage: sessionStorage,
	})

	return &Client{
		client:  tgClient,
		storage: sessionStorage,
		appID:   cfg.AppID,
		appHash: cfg.AppHash,
	}
}

func (c *Client) Run(ctx context.Context, f func(ctx context.Context) error) error {
	return c.client.Run(ctx, func(ctx context.Context) error {
		c.mu.Lock()
		c.api = c.client.API()
		c.sender = message.NewSender(c.api)
		c.running = true
		c.mu.Unlock()

		defer func() {
			c.mu.Lock()
			c.running = false
			c.mu.Unlock()
		}()

		status, err := c.client.Auth().Status(ctx)
		if err != nil {
			return fmt.Errorf("failed to get auth status: %w", err)
		}

		c.mu.Lock()
		c.authorized = status.Authorized
		c.mu.Unlock()

		return f(ctx)
	})
}

func (c *Client) API() *tg.Client {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.api
}

func (c *Client) Sender() *message.Sender {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.sender
}

func (c *Client) IsRunning() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.running
}

func (c *Client) IsAuthorized() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.authorized
}

func (c *Client) GetPhone() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.phone
}

func (c *Client) SendCode(ctx context.Context, phone string) (string, error) {
	if !c.IsRunning() {
		return "", fmt.Errorf("client is not running")
	}

	sentCode, err := c.api.AuthSendCode(ctx, &tg.AuthSendCodeRequest{
		PhoneNumber: phone,
		APIID:       c.appID,
		APIHash:     c.appHash,
		Settings:    tg.CodeSettings{},
	})
	if err != nil {
		return "", fmt.Errorf("failed to send code: %w", err)
	}

	var codeHash string
	switch s := sentCode.(type) {
	case *tg.AuthSentCode:
		codeHash = s.PhoneCodeHash
	case *tg.AuthSentCodeSuccess:
		c.mu.Lock()
		c.authorized = true
		c.mu.Unlock()
		return "", nil
	default:
		return "", fmt.Errorf("unexpected response type")
	}

	c.mu.Lock()
	c.phone = phone
	c.codeHash = codeHash
	c.mu.Unlock()

	return codeHash, nil
}

func (c *Client) SignIn(ctx context.Context, code string, password string) error {
	if !c.IsRunning() {
		return fmt.Errorf("client is not running")
	}

	c.mu.RLock()
	phone := c.phone
	codeHash := c.codeHash
	c.mu.RUnlock()

	if phone == "" || codeHash == "" {
		return fmt.Errorf("SendCode must be called first")
	}

	_, err := c.api.AuthSignIn(ctx, &tg.AuthSignInRequest{
		PhoneNumber:   phone,
		PhoneCodeHash: codeHash,
		PhoneCode:     code,
	})
	if err != nil {
		if isSessionPasswordNeeded(err) {
			if password == "" {
				return fmt.Errorf("2FA password required - please provide password parameter")
			}
			_, err = c.client.Auth().Password(ctx, password)
			if err != nil {
				return fmt.Errorf("failed to authenticate with password: %w", err)
			}
		} else {
			return fmt.Errorf("failed to sign in: %w", err)
		}
	}

	c.mu.Lock()
	c.authorized = true
	c.mu.Unlock()

	return nil
}

func (c *Client) CheckAuthStatus(ctx context.Context) (bool, error) {
	if !c.IsRunning() {
		return false, fmt.Errorf("client is not running")
	}

	status, err := c.client.Auth().Status(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to get auth status: %w", err)
	}

	c.mu.Lock()
	c.authorized = status.Authorized
	c.mu.Unlock()

	return status.Authorized, nil
}

func isSessionPasswordNeeded(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "SESSION_PASSWORD_NEEDED")
}

func (c *Client) Logout(ctx context.Context) error {
	if !c.IsRunning() {
		return fmt.Errorf("client is not running")
	}

	if !c.IsAuthorized() {
		return fmt.Errorf("not authorized")
	}

	_, err := c.api.AuthLogOut(ctx)
	if err != nil {
		return fmt.Errorf("failed to logout: %w", err)
	}

	if err := c.storage.Clear(); err != nil {
		return fmt.Errorf("failed to clear session: %w", err)
	}

	c.mu.Lock()
	c.authorized = false
	c.phone = ""
	c.codeHash = ""
	c.mu.Unlock()

	return nil
}
