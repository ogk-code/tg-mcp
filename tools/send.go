package tools

import (
	"context"
	"fmt"

	"github.com/gotd/td/tg"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"tg-mcp/client"
)

type SendMessageInput struct {
	Chat string `json:"chat"`
	Text string `json:"text"`
}

type ReplyMessageInput struct {
	Chat      string `json:"chat"`
	Text      string `json:"text"`
	MessageID int    `json:"message_id"`
}

type ReplyMessageOutput struct {
	Success   bool   `json:"success"`
	MessageID int    `json:"message_id,omitempty"`
	Message   string `json:"message,omitempty"`
}

type SendMessageOutput struct {
	Success   bool   `json:"success"`
	MessageID int    `json:"message_id,omitempty"`
	Message   string `json:"message,omitempty"`
}

func SendMessage(c *client.Client) func(ctx context.Context, req *mcp.CallToolRequest, input SendMessageInput) (*mcp.CallToolResult, SendMessageOutput, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input SendMessageInput) (*mcp.CallToolResult, SendMessageOutput, error) {
		if !c.IsAuthorized() {
			return nil, SendMessageOutput{
				Success: false,
				Message: "Not authorized. Please use auth_send_code and auth_submit_code first.",
			}, nil
		}

		sender := c.Sender()
		if sender == nil {
			return nil, SendMessageOutput{
				Success: false,
				Message: "Client is not running",
			}, nil
		}

		api := c.API()
		inputPeer, err := getPeerFromDialogsOrResolve(ctx, api, input.Chat)
		if err != nil {
			return nil, SendMessageOutput{
				Success: false,
				Message: fmt.Sprintf("Failed to resolve chat: %v", err),
			}, nil
		}

		updates, err := sender.To(inputPeer).Text(ctx, input.Text)
		if err != nil {
			return nil, SendMessageOutput{
				Success: false,
				Message: fmt.Sprintf("Failed to send message: %v", err),
			}, nil
		}

		messageID := extractMessageID(updates)

		return nil, SendMessageOutput{
			Success:   true,
			MessageID: messageID,
			Message:   "Message sent successfully",
		}, nil
	}
}

func ReplyMessage(c *client.Client) func(ctx context.Context, req *mcp.CallToolRequest, input ReplyMessageInput) (*mcp.CallToolResult, ReplyMessageOutput, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input ReplyMessageInput) (*mcp.CallToolResult, ReplyMessageOutput, error) {
		if !c.IsAuthorized() {
			return nil, ReplyMessageOutput{
				Success: false,
				Message: "Not authorized",
			}, nil
		}

		sender := c.Sender()
		if sender == nil {
			return nil, ReplyMessageOutput{
				Success: false,
				Message: "Client is not running",
			}, nil
		}

		api := c.API()
		inputPeer, err := getPeerFromDialogsOrResolve(ctx, api, input.Chat)
		if err != nil {
			return nil, ReplyMessageOutput{
				Success: false,
				Message: fmt.Sprintf("Failed to resolve chat: %v", err),
			}, nil
		}

		updates, err := sender.To(inputPeer).Reply(input.MessageID).Text(ctx, input.Text)
		if err != nil {
			return nil, ReplyMessageOutput{
				Success: false,
				Message: fmt.Sprintf("Failed to send reply: %v", err),
			}, nil
		}

		messageID := extractMessageID(updates)

		return nil, ReplyMessageOutput{
			Success:   true,
			MessageID: messageID,
			Message:   "Reply sent successfully",
		}, nil
	}
}

func extractMessageID(updates tg.UpdatesClass) int {
	switch u := updates.(type) {
	case *tg.Updates:
		for _, update := range u.Updates {
			if msg, ok := update.(*tg.UpdateMessageID); ok {
				return msg.ID
			}
		}
	case *tg.UpdateShortSentMessage:
		return u.ID
	}
	return 0
}

type ForwardMessageInput struct {
	FromChat  string `json:"from_chat"`
	ToChat    string `json:"to_chat"`
	MessageID int    `json:"message_id"`
}

type ForwardMessageOutput struct {
	Success   bool   `json:"success"`
	MessageID int    `json:"message_id,omitempty"`
	Message   string `json:"message,omitempty"`
}

func ForwardMessage(c *client.Client) func(ctx context.Context, req *mcp.CallToolRequest, input ForwardMessageInput) (*mcp.CallToolResult, ForwardMessageOutput, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input ForwardMessageInput) (*mcp.CallToolResult, ForwardMessageOutput, error) {
		if !c.IsAuthorized() {
			return nil, ForwardMessageOutput{
				Success: false,
				Message: "Not authorized",
			}, nil
		}

		api := c.API()
		if api == nil {
			return nil, ForwardMessageOutput{
				Success: false,
				Message: "Client is not running",
			}, nil
		}

		fromPeer, err := getPeerFromDialogsOrResolve(ctx, api, input.FromChat)
		if err != nil {
			return nil, ForwardMessageOutput{
				Success: false,
				Message: fmt.Sprintf("Failed to resolve source chat: %v", err),
			}, nil
		}

		toPeer, err := getPeerFromDialogsOrResolve(ctx, api, input.ToChat)
		if err != nil {
			return nil, ForwardMessageOutput{
				Success: false,
				Message: fmt.Sprintf("Failed to resolve destination chat: %v", err),
			}, nil
		}

		updates, err := api.MessagesForwardMessages(ctx, &tg.MessagesForwardMessagesRequest{
			FromPeer: fromPeer,
			ToPeer:   toPeer,
			ID:       []int{input.MessageID},
			RandomID: []int64{int64(input.MessageID) + 1000000},
		})
		if err != nil {
			return nil, ForwardMessageOutput{
				Success: false,
				Message: fmt.Sprintf("Failed to forward message: %v", err),
			}, nil
		}

		messageID := extractMessageID(updates)

		return nil, ForwardMessageOutput{
			Success:   true,
			MessageID: messageID,
			Message:   "Message forwarded successfully",
		}, nil
	}
}

func getPeerFromDialogsOrResolve(ctx context.Context, api *tg.Client, chat string) (tg.InputPeerClass, error) {
	chatID, isNumeric := parseID(chat)

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
		if user, ok := u.(*tg.User); ok {
			if isNumeric && user.ID == chatID {
				return &tg.InputPeerUser{UserID: user.ID, AccessHash: user.AccessHash}, nil
			}
			if !isNumeric && user.Username == chat {
				return &tg.InputPeerUser{UserID: user.ID, AccessHash: user.AccessHash}, nil
			}
		}
	}

	for _, ch := range chats {
		switch c := ch.(type) {
		case *tg.Chat:
			if isNumeric && c.ID == chatID {
				return &tg.InputPeerChat{ChatID: c.ID}, nil
			}
		case *tg.Channel:
			if isNumeric && c.ID == chatID {
				return &tg.InputPeerChannel{ChannelID: c.ID, AccessHash: c.AccessHash}, nil
			}
			if !isNumeric && c.Username == chat {
				return &tg.InputPeerChannel{ChannelID: c.ID, AccessHash: c.AccessHash}, nil
			}
		}
	}

	if !isNumeric {
		resolved, err := api.ContactsResolveUsername(ctx, &tg.ContactsResolveUsernameRequest{
			Username: chat,
		})
		if err == nil {
			for _, u := range resolved.Users {
				if user, ok := u.(*tg.User); ok {
					return &tg.InputPeerUser{UserID: user.ID, AccessHash: user.AccessHash}, nil
				}
			}
			for _, ch := range resolved.Chats {
				switch c := ch.(type) {
				case *tg.Chat:
					return &tg.InputPeerChat{ChatID: c.ID}, nil
				case *tg.Channel:
					return &tg.InputPeerChannel{ChannelID: c.ID, AccessHash: c.AccessHash}, nil
				}
			}
		}
	}

	return nil, fmt.Errorf("chat not found: %s", chat)
}

func RegisterSendTools(server *mcp.Server, c *client.Client) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "send_message",
		Description: "Send a text message to a Telegram chat by username or ID",
	}, SendMessage(c))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "reply_message",
		Description: "Reply to a specific message in a chat",
	}, ReplyMessage(c))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "forward_message",
		Description: "Forward a message from one chat to another",
	}, ForwardMessage(c))
}
