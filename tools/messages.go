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

type GetHistoryInput struct {
	Chat     string `json:"chat"`
	Limit    int    `json:"limit,omitempty"`
	OffsetID int    `json:"offset_id,omitempty"`
}

type GetHistoryOutput struct {
	Success    bool      `json:"success"`
	Messages   []Message `json:"messages,omitempty"`
	NextOffset int       `json:"next_offset,omitempty"`
	HasMore    bool      `json:"has_more"`
	Total      int       `json:"total,omitempty"`
	Message    string    `json:"message,omitempty"`
}

func GetHistory(c *client.Client) func(ctx context.Context, req *mcp.CallToolRequest, input GetHistoryInput) (*mcp.CallToolResult, GetHistoryOutput, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input GetHistoryInput) (*mcp.CallToolResult, GetHistoryOutput, error) {
		if !c.IsAuthorized() {
			return nil, GetHistoryOutput{
				Success: false,
				Message: "Not authorized",
			}, nil
		}

		api := c.API()
		if api == nil {
			return nil, GetHistoryOutput{
				Success: false,
				Message: "Client is not running",
			}, nil
		}

		limit := input.Limit
		if limit <= 0 {
			limit = 100
		}
		if limit > 1000 {
			limit = 1000
		}

		inputPeer, err := resolvePeerOrDialogs(ctx, api, input.Chat)
		if err != nil {
			return nil, GetHistoryOutput{
				Success: false,
				Message: fmt.Sprintf("Failed to resolve chat: %v", err),
			}, nil
		}

		history, err := api.MessagesGetHistory(ctx, &tg.MessagesGetHistoryRequest{
			Peer:     inputPeer,
			Limit:    limit,
			OffsetID: input.OffsetID,
		})
		if err != nil {
			return nil, GetHistoryOutput{
				Success: false,
				Message: fmt.Sprintf("Failed to get history: %v", err),
			}, nil
		}

		messages, users := extractMessagesAndUsers(history)
		total := extractTotalCount(history)

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
		var lastID int
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
				lastID = m.ID
			}
		}

		hasMore := len(messages) == limit
		nextOffset := 0
		if hasMore && lastID > 0 {
			nextOffset = lastID
		}

		return nil, GetHistoryOutput{
			Success:    true,
			Messages:   result,
			NextOffset: nextOffset,
			HasMore:    hasMore,
			Total:      total,
		}, nil
	}
}

func resolvePeerOrDialogs(ctx context.Context, api *tg.Client, chat string) (tg.InputPeerClass, error) {
	if chatID, err := parseIntID(chat); err == nil {
		dialogs, err := api.MessagesGetDialogs(ctx, &tg.MessagesGetDialogsRequest{
			OffsetPeer: &tg.InputPeerEmpty{},
			Limit:      200,
		})
		if err != nil {
			return nil, err
		}

		var users []tg.UserClass
		var chats []tg.ChatClass
		switch d := dialogs.(type) {
		case *tg.MessagesDialogs:
			users = d.Users
			chats = d.Chats
		case *tg.MessagesDialogsSlice:
			users = d.Users
			chats = d.Chats
		}

		for _, u := range users {
			if user, ok := u.(*tg.User); ok && user.ID == chatID {
				return &tg.InputPeerUser{UserID: user.ID, AccessHash: user.AccessHash}, nil
			}
		}

		for _, ch := range chats {
			switch c := ch.(type) {
			case *tg.Chat:
				if c.ID == chatID {
					return &tg.InputPeerChat{ChatID: c.ID}, nil
				}
			case *tg.Channel:
				if c.ID == chatID {
					return &tg.InputPeerChannel{ChannelID: c.ID, AccessHash: c.AccessHash}, nil
				}
			}
		}

		return nil, fmt.Errorf("chat %d not found in dialogs", chatID)
	}

	return resolvePeer(ctx, api, chat)
}

func parseIntID(s string) (int64, error) {
	return parseInt64(s)
}

func parseInt64(s string) (int64, error) {
	var id int64
	_, err := fmt.Sscanf(s, "%d", &id)
	return id, err
}

func extractTotalCount(mm tg.MessagesMessagesClass) int {
	switch m := mm.(type) {
	case *tg.MessagesMessages:
		return len(m.Messages)
	case *tg.MessagesMessagesSlice:
		return m.Count
	case *tg.MessagesChannelMessages:
		return m.Count
	default:
		return 0
	}
}

func RegisterMessagesTools(server *mcp.Server, c *client.Client) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_messages",
		Description: "Get recent messages from a Telegram chat (up to 100)",
	}, GetMessages(c))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_history",
		Description: "Get chat history with pagination. Use limit (up to 1000) and offset_id for chunked loading. Returns next_offset for next chunk.",
	}, GetHistory(c))
}
