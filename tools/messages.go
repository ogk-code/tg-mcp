package tools

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gotd/td/tg"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"tg-mcp/client"
)

type GetMessagesInput struct {
	Chat  string `json:"chat"`
	Limit int    `json:"limit,omitempty"`
}

type Message struct {
	ID       int    `json:"id"`
	Text     string `json:"text"`
	FromID   int64  `json:"from_id,omitempty"`
	FromName string `json:"from_name,omitempty"`
	Date     string `json:"date"`
	IsOut    bool   `json:"is_out"`
}

type GetMessagesOutput struct {
	Success  bool      `json:"success"`
	Messages []Message `json:"messages,omitempty"`
	Message  string    `json:"message,omitempty"`
}

func GetMessages(c *client.Client) func(ctx context.Context, req *mcp.CallToolRequest, input GetMessagesInput) (*mcp.CallToolResult, GetMessagesOutput, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input GetMessagesInput) (*mcp.CallToolResult, GetMessagesOutput, error) {
		if !c.IsAuthorized() {
			return nil, GetMessagesOutput{
				Success: false,
				Message: "Not authorized. Please use auth_send_code and auth_submit_code first.",
			}, nil
		}

		api := c.API()
		if api == nil {
			return nil, GetMessagesOutput{
				Success: false,
				Message: "Client is not running",
			}, nil
		}

		limit := input.Limit
		if limit <= 0 {
			limit = 10
		}
		if limit > 100 {
			limit = 100
		}

		inputPeer, err := resolvePeer(ctx, api, input.Chat)
		if err != nil {
			return nil, GetMessagesOutput{
				Success: false,
				Message: fmt.Sprintf("Failed to resolve chat: %v", err),
			}, nil
		}

		history, err := api.MessagesGetHistory(ctx, &tg.MessagesGetHistoryRequest{
			Peer:  inputPeer,
			Limit: limit,
		})
		if err != nil {
			return nil, GetMessagesOutput{
				Success: false,
				Message: fmt.Sprintf("Failed to get messages: %v", err),
			}, nil
		}

		messages, users := extractMessagesAndUsers(history)

		userMap := make(map[int64]string)
		for _, u := range users {
			if user, ok := u.(*tg.User); ok {
				name := user.FirstName
				if user.LastName != "" {
					name += " " + user.LastName
				}
				userMap[user.ID] = name
			}
		}

		result := make([]Message, 0, len(messages))
		for _, msg := range messages {
			if m, ok := msg.(*tg.Message); ok {
				fromID := int64(0)
				fromName := ""
				if m.FromID != nil {
					if userPeer, ok := m.FromID.(*tg.PeerUser); ok {
						fromID = userPeer.UserID
						fromName = userMap[fromID]
					}
				}

				result = append(result, Message{
					ID:       m.ID,
					Text:     m.Message,
					FromID:   fromID,
					FromName: fromName,
					Date:     time.Unix(int64(m.Date), 0).Format(time.RFC3339),
					IsOut:    m.Out,
				})
			}
		}

		return nil, GetMessagesOutput{
			Success:  true,
			Messages: result,
		}, nil
	}
}

func resolvePeer(ctx context.Context, api *tg.Client, chat string) (tg.InputPeerClass, error) {
	username := strings.TrimPrefix(chat, "@")

	resolved, err := api.ContactsResolveUsername(ctx, &tg.ContactsResolveUsernameRequest{
		Username: username,
	})
	if err != nil {
		return nil, err
	}

	for _, u := range resolved.Users {
		if user, ok := u.(*tg.User); ok {
			return &tg.InputPeerUser{
				UserID:     user.ID,
				AccessHash: user.AccessHash,
			}, nil
		}
	}

	for _, ch := range resolved.Chats {
		switch c := ch.(type) {
		case *tg.Chat:
			return &tg.InputPeerChat{ChatID: c.ID}, nil
		case *tg.Channel:
			return &tg.InputPeerChannel{
				ChannelID:  c.ID,
				AccessHash: c.AccessHash,
			}, nil
		}
	}

	return nil, fmt.Errorf("could not resolve peer: %s", chat)
}

func extractMessagesAndUsers(mm tg.MessagesMessagesClass) ([]tg.MessageClass, []tg.UserClass) {
	switch m := mm.(type) {
	case *tg.MessagesMessages:
		return m.Messages, m.Users
	case *tg.MessagesMessagesSlice:
		return m.Messages, m.Users
	case *tg.MessagesChannelMessages:
		return m.Messages, m.Users
	default:
		return nil, nil
	}
}

func RegisterMessagesTools(server *mcp.Server, c *client.Client) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_messages",
		Description: "Get recent messages from a Telegram chat by username or ID",
	}, GetMessages(c))
}
