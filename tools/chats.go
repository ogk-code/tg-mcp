package tools

import (
	"context"
	"fmt"
	"time"

	"github.com/gotd/td/tg"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"tg-mcp/client"
)

func formatDate(timestamp int) string {
	return time.Unix(int64(timestamp), 0).Format(time.RFC3339)
}

type ListChatsInput struct {
	Limit int `json:"limit,omitempty"`
}

type Chat struct {
	ID          int64  `json:"id"`
	Title       string `json:"title"`
	Type        string `json:"type"`
	Username    string `json:"username,omitempty"`
	UnreadCount int    `json:"unread_count"`
}

type ListChatsOutput struct {
	Success bool   `json:"success"`
	Chats   []Chat `json:"chats,omitempty"`
	Message string `json:"message,omitempty"`
}

func ListChats(c *client.Client) func(ctx context.Context, req *mcp.CallToolRequest, input ListChatsInput) (*mcp.CallToolResult, ListChatsOutput, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input ListChatsInput) (*mcp.CallToolResult, ListChatsOutput, error) {
		if !c.IsAuthorized() {
			return nil, ListChatsOutput{
				Success: false,
				Message: "Not authorized. Please use auth_send_code and auth_submit_code first.",
			}, nil
		}

		api := c.API()
		if api == nil {
			return nil, ListChatsOutput{
				Success: false,
				Message: "Client is not running",
			}, nil
		}

		limit := input.Limit
		if limit <= 0 {
			limit = 20
		}
		if limit > 100 {
			limit = 100
		}

		dialogs, err := api.MessagesGetDialogs(ctx, &tg.MessagesGetDialogsRequest{
			OffsetPeer: &tg.InputPeerEmpty{},
			Limit:      limit,
		})
		if err != nil {
			return nil, ListChatsOutput{
				Success: false,
				Message: fmt.Sprintf("Failed to get dialogs: %v", err),
			}, nil
		}

		dialogList, chats, users := extractDialogsData(dialogs)

		userMap := make(map[int64]*tg.User)
		for _, u := range users {
			if user, ok := u.(*tg.User); ok {
				userMap[user.ID] = user
			}
		}

		chatMap := make(map[int64]tg.ChatClass)
		for _, ch := range chats {
			switch c := ch.(type) {
			case *tg.Chat:
				chatMap[c.ID] = c
			case *tg.Channel:
				chatMap[c.ID] = c
			}
		}

		result := make([]Chat, 0, len(dialogList))
		for _, d := range dialogList {
			dialog, ok := d.(*tg.Dialog)
			if !ok {
				continue
			}

			chat := Chat{
				UnreadCount: dialog.UnreadCount,
			}

			switch p := dialog.Peer.(type) {
			case *tg.PeerUser:
				chat.ID = p.UserID
				chat.Type = "user"
				if user, ok := userMap[p.UserID]; ok {
					chat.Title = user.FirstName
					if user.LastName != "" {
						chat.Title += " " + user.LastName
					}
					chat.Username = user.Username
				}
			case *tg.PeerChat:
				chat.ID = p.ChatID
				chat.Type = "chat"
				if ch, ok := chatMap[p.ChatID]; ok {
					if c, ok := ch.(*tg.Chat); ok {
						chat.Title = c.Title
					}
				}
			case *tg.PeerChannel:
				chat.ID = p.ChannelID
				chat.Type = "channel"
				if ch, ok := chatMap[p.ChannelID]; ok {
					if c, ok := ch.(*tg.Channel); ok {
						chat.Title = c.Title
						chat.Username = c.Username
					}
				}
			}

			result = append(result, chat)
		}

		return nil, ListChatsOutput{
			Success: true,
			Chats:   result,
		}, nil
	}
}

func extractDialogsData(md tg.MessagesDialogsClass) ([]tg.DialogClass, []tg.ChatClass, []tg.UserClass) {
	switch m := md.(type) {
	case *tg.MessagesDialogs:
		return m.Dialogs, m.Chats, m.Users
	case *tg.MessagesDialogsSlice:
		return m.Dialogs, m.Chats, m.Users
	default:
		return nil, nil, nil
	}
}

type ChatOverview struct {
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	Type        string    `json:"type"`
	Username    string    `json:"username,omitempty"`
	UnreadCount int       `json:"unread_count"`
	Messages    []Message `json:"messages,omitempty"`
}

type GetChatsOverviewInput struct {
	ChatsLimit    int `json:"chats_limit,omitempty"`
	MessagesLimit int `json:"messages_limit,omitempty"`
}

type GetChatsOverviewOutput struct {
	Success bool           `json:"success"`
	Chats   []ChatOverview `json:"chats,omitempty"`
	Message string         `json:"message,omitempty"`
}

func GetChatsOverview(c *client.Client) func(ctx context.Context, req *mcp.CallToolRequest, input GetChatsOverviewInput) (*mcp.CallToolResult, GetChatsOverviewOutput, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input GetChatsOverviewInput) (*mcp.CallToolResult, GetChatsOverviewOutput, error) {
		if !c.IsAuthorized() {
			return nil, GetChatsOverviewOutput{
				Success: false,
				Message: "Not authorized. Please use auth_send_code and auth_submit_code first.",
			}, nil
		}

		api := c.API()
		if api == nil {
			return nil, GetChatsOverviewOutput{
				Success: false,
				Message: "Client is not running",
			}, nil
		}

		chatsLimit := input.ChatsLimit
		if chatsLimit <= 0 {
			chatsLimit = 20
		}
		if chatsLimit > 50 {
			chatsLimit = 50
		}

		messagesLimit := input.MessagesLimit
		if messagesLimit <= 0 {
			messagesLimit = 3
		}
		if messagesLimit > 10 {
			messagesLimit = 10
		}

		dialogs, err := api.MessagesGetDialogs(ctx, &tg.MessagesGetDialogsRequest{
			OffsetPeer: &tg.InputPeerEmpty{},
			Limit:      chatsLimit,
		})
		if err != nil {
			return nil, GetChatsOverviewOutput{
				Success: false,
				Message: fmt.Sprintf("Failed to get dialogs: %v", err),
			}, nil
		}

		dialogList, chats, users := extractDialogsData(dialogs)

		userMap := make(map[int64]*tg.User)
		for _, u := range users {
			if user, ok := u.(*tg.User); ok {
				userMap[user.ID] = user
			}
		}

		chatMap := make(map[int64]tg.ChatClass)
		for _, ch := range chats {
			switch c := ch.(type) {
			case *tg.Chat:
				chatMap[c.ID] = c
			case *tg.Channel:
				chatMap[c.ID] = c
			}
		}

		result := make([]ChatOverview, 0, len(dialogList))
		for _, d := range dialogList {
			dialog, ok := d.(*tg.Dialog)
			if !ok {
				continue
			}

			chat := ChatOverview{
				UnreadCount: dialog.UnreadCount,
			}

			var inputPeer tg.InputPeerClass

			switch p := dialog.Peer.(type) {
			case *tg.PeerUser:
				chat.ID = p.UserID
				chat.Type = "user"
				if user, ok := userMap[p.UserID]; ok {
					chat.Title = user.FirstName
					if user.LastName != "" {
						chat.Title += " " + user.LastName
					}
					chat.Username = user.Username
					inputPeer = &tg.InputPeerUser{
						UserID:     user.ID,
						AccessHash: user.AccessHash,
					}
				}
			case *tg.PeerChat:
				chat.ID = p.ChatID
				chat.Type = "chat"
				if ch, ok := chatMap[p.ChatID]; ok {
					if c, ok := ch.(*tg.Chat); ok {
						chat.Title = c.Title
					}
				}
				inputPeer = &tg.InputPeerChat{ChatID: p.ChatID}
			case *tg.PeerChannel:
				chat.ID = p.ChannelID
				chat.Type = "channel"
				if ch, ok := chatMap[p.ChannelID]; ok {
					if channel, ok := ch.(*tg.Channel); ok {
						chat.Title = channel.Title
						chat.Username = channel.Username
						inputPeer = &tg.InputPeerChannel{
							ChannelID:  channel.ID,
							AccessHash: channel.AccessHash,
						}
					}
				}
			}

			if inputPeer != nil {
				history, err := api.MessagesGetHistory(ctx, &tg.MessagesGetHistoryRequest{
					Peer:  inputPeer,
					Limit: messagesLimit,
				})
				if err == nil {
					messages, msgUsers := extractMessagesAndUsers(history)

					msgUserMap := make(map[int64]string)
					for _, u := range msgUsers {
						if user, ok := u.(*tg.User); ok {
							name := user.FirstName
							if user.LastName != "" {
								name += " " + user.LastName
							}
							msgUserMap[user.ID] = name
						}
					}

					for _, msg := range messages {
						if m, ok := msg.(*tg.Message); ok {
							fromID := int64(0)
							fromName := ""
							if m.FromID != nil {
								if userPeer, ok := m.FromID.(*tg.PeerUser); ok {
									fromID = userPeer.UserID
									fromName = msgUserMap[fromID]
								}
							}

							chat.Messages = append(chat.Messages, Message{
								ID:       m.ID,
								Text:     m.Message,
								FromID:   fromID,
								FromName: fromName,
								Date:     formatDate(m.Date),
								IsOut:    m.Out,
							})
						}
					}
				}
			}

			result = append(result, chat)
		}

		return nil, GetChatsOverviewOutput{
			Success: true,
			Chats:   result,
		}, nil
	}
}

func RegisterChatsTools(server *mcp.Server, c *client.Client) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_chats",
		Description: "Get list of Telegram dialogs/chats with unread counts",
	}, ListChats(c))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_chats_overview",
		Description: "Get all chats with their recent messages in one request. Use chats_limit (default 20, max 50) and messages_limit (default 3, max 10) to control output size.",
	}, GetChatsOverview(c))
}
