package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"tg-mcp/client"
)

type AuthStatusInput struct{}

type AuthStatusOutput struct {
	Authorized bool   `json:"authorized"`
	Phone      string `json:"phone,omitempty"`
}

func AuthStatus(c *client.Client) func(ctx context.Context, req *mcp.CallToolRequest, input AuthStatusInput) (*mcp.CallToolResult, AuthStatusOutput, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input AuthStatusInput) (*mcp.CallToolResult, AuthStatusOutput, error) {
		authorized, err := c.CheckAuthStatus(ctx)
		if err != nil {
			return nil, AuthStatusOutput{}, err
		}

		return nil, AuthStatusOutput{
			Authorized: authorized,
			Phone:      c.GetPhone(),
		}, nil
	}
}

type SendCodeInput struct {
	Phone string `json:"phone"`
}

type SendCodeOutput struct {
	Success  bool   `json:"success"`
	CodeHash string `json:"code_hash"`
	Message  string `json:"message,omitempty"`
}

func AuthSendCode(c *client.Client) func(ctx context.Context, req *mcp.CallToolRequest, input SendCodeInput) (*mcp.CallToolResult, SendCodeOutput, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input SendCodeInput) (*mcp.CallToolResult, SendCodeOutput, error) {
		codeHash, err := c.SendCode(ctx, input.Phone)
		if err != nil {
			return nil, SendCodeOutput{
				Success: false,
				Message: err.Error(),
			}, nil
		}

		return nil, SendCodeOutput{
			Success:  true,
			CodeHash: codeHash,
			Message:  "Code sent to " + input.Phone,
		}, nil
	}
}

type SubmitCodeInput struct {
	Code     string `json:"code"`
	CodeHash string `json:"code_hash"`
	Password string `json:"password,omitempty"`
}

type SubmitCodeOutput struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

func AuthSubmitCode(c *client.Client) func(ctx context.Context, req *mcp.CallToolRequest, input SubmitCodeInput) (*mcp.CallToolResult, SubmitCodeOutput, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input SubmitCodeInput) (*mcp.CallToolResult, SubmitCodeOutput, error) {
		err := c.SignIn(ctx, input.Code, input.Password)
		if err != nil {
			return nil, SubmitCodeOutput{
				Success: false,
				Message: err.Error(),
			}, nil
		}

		return nil, SubmitCodeOutput{
			Success: true,
			Message: "Successfully authorized",
		}, nil
	}
}

type LogoutInput struct{}

type LogoutOutput struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

func AuthLogout(c *client.Client) func(ctx context.Context, req *mcp.CallToolRequest, input LogoutInput) (*mcp.CallToolResult, LogoutOutput, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input LogoutInput) (*mcp.CallToolResult, LogoutOutput, error) {
		err := c.Logout(ctx)
		if err != nil {
			return nil, LogoutOutput{
				Success: false,
				Message: err.Error(),
			}, nil
		}

		return nil, LogoutOutput{
			Success: true,
			Message: "Successfully logged out",
		}, nil
	}
}

func RegisterAuthTools(server *mcp.Server, c *client.Client) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "auth_status",
		Description: "Check Telegram authorization status",
	}, AuthStatus(c))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "auth_send_code",
		Description: "Send authorization code to phone number. Returns code_hash needed for auth_submit_code.",
	}, AuthSendCode(c))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "auth_submit_code",
		Description: "Complete authorization by submitting the code received via Telegram. If 2FA is enabled, include the password.",
	}, AuthSubmitCode(c))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "auth_logout",
		Description: "Logout from current Telegram session and clear stored credentials",
	}, AuthLogout(c))
}
