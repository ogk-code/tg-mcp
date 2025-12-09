package tools

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/gotd/td/tg"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"tg-mcp/client"
)

type GetUserInput struct {
	User string `json:"user"`
}

type UserProfile struct {
	ID        int64  `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name,omitempty"`
	Username  string `json:"username,omitempty"`
	Phone     string `json:"phone,omitempty"`
	Bio       string `json:"bio,omitempty"`
	Bot       bool   `json:"bot"`
	Verified  bool   `json:"verified"`
	Premium   bool   `json:"premium"`
	Online    string `json:"online,omitempty"`
}

type GetUserOutput struct {
	Success bool        `json:"success"`
	User    UserProfile `json:"user,omitempty"`
	Message string      `json:"message,omitempty"`
}

func GetUser(c *client.Client) func(ctx context.Context, req *mcp.CallToolRequest, input GetUserInput) (*mcp.CallToolResult, GetUserOutput, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input GetUserInput) (*mcp.CallToolResult, GetUserOutput, error) {
		if !c.IsAuthorized() {
			return nil, GetUserOutput{
				Success: false,
				Message: "Not authorized",
			}, nil
		}

		api := c.API()
		if api == nil {
			return nil, GetUserOutput{
				Success: false,
				Message: "Client is not running",
			}, nil
		}

		var user *tg.User
		var inputUser tg.InputUserClass

		if userID, err := strconv.ParseInt(input.User, 10, 64); err == nil {
			dialogs, err := api.MessagesGetDialogs(ctx, &tg.MessagesGetDialogsRequest{
				OffsetPeer: &tg.InputPeerEmpty{},
				Limit:      200,
			})
			if err != nil {
				return nil, GetUserOutput{
					Success: false,
					Message: fmt.Sprintf("Failed to get dialogs: %v", err),
				}, nil
			}

			var users []tg.UserClass
			switch d := dialogs.(type) {
			case *tg.MessagesDialogs:
				users = d.Users
			case *tg.MessagesDialogsSlice:
				users = d.Users
			}

			for _, u := range users {
				if usr, ok := u.(*tg.User); ok && usr.ID == userID {
					user = usr
					inputUser = &tg.InputUser{UserID: usr.ID, AccessHash: usr.AccessHash}
					break
				}
			}

			if user == nil {
				return nil, GetUserOutput{
					Success: false,
					Message: fmt.Sprintf("User %d not found in dialogs", userID),
				}, nil
			}
		} else {
			username := strings.TrimPrefix(input.User, "@")
			resolved, err := api.ContactsResolveUsername(ctx, &tg.ContactsResolveUsernameRequest{
				Username: username,
			})
			if err != nil {
				return nil, GetUserOutput{
					Success: false,
					Message: fmt.Sprintf("Failed to resolve username: %v", err),
				}, nil
			}

			for _, u := range resolved.Users {
				if usr, ok := u.(*tg.User); ok {
					user = usr
					inputUser = &tg.InputUser{UserID: usr.ID, AccessHash: usr.AccessHash}
					break
				}
			}

			if user == nil {
				return nil, GetUserOutput{
					Success: false,
					Message: "User not found",
				}, nil
			}
		}

		profile := UserProfile{
			ID:        user.ID,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Username:  user.Username,
			Phone:     user.Phone,
			Bot:       user.Bot,
			Verified:  user.Verified,
			Premium:   user.Premium,
		}

		if user.Status != nil {
			switch s := user.Status.(type) {
			case *tg.UserStatusOnline:
				profile.Online = "online"
			case *tg.UserStatusOffline:
				profile.Online = formatDate(s.WasOnline)
			case *tg.UserStatusRecently:
				profile.Online = "recently"
			case *tg.UserStatusLastWeek:
				profile.Online = "last week"
			case *tg.UserStatusLastMonth:
				profile.Online = "last month"
			}
		}

		if inputUser != nil {
			fullUser, err := api.UsersGetFullUser(ctx, inputUser)
			if err == nil && fullUser != nil {
				profile.Bio = fullUser.FullUser.About
			}
		}

		return nil, GetUserOutput{
			Success: true,
			User:    profile,
		}, nil
	}
}

func RegisterUsersTools(server *mcp.Server, c *client.Client) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_user",
		Description: "Get user profile information by username or ID",
	}, GetUser(c))
}
