package tools

import (
	"context"
	"fmt"
	"strconv"

	"github.com/gotd/td/tg"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"tg-mcp/client"
)

type SendMessageInput struct {
	Chat string `json:"chat"`
	Text string `json:"text"`
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

		var updates tg.UpdatesClass
		var err error

		if userID, parseErr := strconv.ParseInt(input.Chat, 10, 64); parseErr == nil {
			api := c.API()
			dialogs, err := api.MessagesGetDialogs(ctx, &tg.MessagesGetDialogsRequest{
				OffsetPeer: &tg.InputPeerEmpty{},
				Limit:      100,
			})
			if err != nil {
				return nil, SendMessageOutput{
					Success: false,
					Message: fmt.Sprintf("Failed to get dialogs: %v", err),
				}, nil
			}

			var inputPeer tg.InputPeerClass
			if md, ok := dialogs.(*tg.MessagesDialogs); ok {
				for _, u := range md.Users {
					if user, ok := u.(*tg.User); ok && user.ID == userID {
						inputPeer = &tg.InputPeerUser{UserID: user.ID, AccessHash: user.AccessHash}
						break
					}
				}
			} else if md, ok := dialogs.(*tg.MessagesDialogsSlice); ok {
				for _, u := range md.Users {
					if user, ok := u.(*tg.User); ok && user.ID == userID {
						inputPeer = &tg.InputPeerUser{UserID: user.ID, AccessHash: user.AccessHash}
						break
					}
				}
			}

			if inputPeer == nil {
				return nil, SendMessageOutput{
					Success: false,
					Message: fmt.Sprintf("User %d not found in dialogs", userID),
				}, nil
			}

			updates, err = sender.To(inputPeer).Text(ctx, input.Text)
		} else {
			updates, err = sender.Resolve(input.Chat).Text(ctx, input.Text)
		}

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

func RegisterSendTools(server *mcp.Server, c *client.Client) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "send_message",
		Description: "Send a text message to a Telegram chat by username or ID",
	}, SendMessage(c))
}
