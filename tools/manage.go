package tools

import (
	"context"
	"fmt"
	"strconv"

	"github.com/gotd/td/tg"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"tg-mcp/client"
)

type DeleteChatInput struct {
	Chat string `json:"chat"`
}

type DeleteChatOutput struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

func DeleteChat(c *client.Client) func(ctx context.Context, req *mcp.CallToolRequest, input DeleteChatInput) (*mcp.CallToolResult, DeleteChatOutput, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input DeleteChatInput) (*mcp.CallToolResult, DeleteChatOutput, error) {
		if !c.IsAuthorized() {
			return nil, DeleteChatOutput{
				Success: false,
				Message: "Not authorized",
			}, nil
		}

		api := c.API()
		if api == nil {
			return nil, DeleteChatOutput{
				Success: false,
				Message: "Client is not running",
			}, nil
		}

		inputPeer, err := getPeerFromDialogs(ctx, api, input.Chat)
		if err != nil {
			return nil, DeleteChatOutput{
				Success: false,
				Message: fmt.Sprintf("Failed to find chat: %v", err),
			}, nil
		}

		_, err = api.MessagesDeleteHistory(ctx, &tg.MessagesDeleteHistoryRequest{
			Peer:   inputPeer,
			MaxID:  0,
			Revoke: true,
		})
		if err != nil {
			return nil, DeleteChatOutput{
				Success: false,
				Message: fmt.Sprintf("Failed to delete chat: %v", err),
			}, nil
		}

		return nil, DeleteChatOutput{
			Success: true,
			Message: "Chat deleted",
		}, nil
	}
}

type LeaveChannelInput struct {
	Channel string `json:"channel"`
}

type LeaveChannelOutput struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

func LeaveChannel(c *client.Client) func(ctx context.Context, req *mcp.CallToolRequest, input LeaveChannelInput) (*mcp.CallToolResult, LeaveChannelOutput, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input LeaveChannelInput) (*mcp.CallToolResult, LeaveChannelOutput, error) {
		if !c.IsAuthorized() {
			return nil, LeaveChannelOutput{
				Success: false,
				Message: "Not authorized",
			}, nil
		}

		api := c.API()
		if api == nil {
			return nil, LeaveChannelOutput{
				Success: false,
				Message: "Client is not running",
			}, nil
		}

		inputChannel, err := getChannelFromDialogs(ctx, api, input.Channel)
		if err == nil {
			_, err = api.ChannelsLeaveChannel(ctx, inputChannel)
			if err != nil {
				return nil, LeaveChannelOutput{
					Success: false,
					Message: fmt.Sprintf("Failed to leave channel: %v", err),
				}, nil
			}
			return nil, LeaveChannelOutput{
				Success: true,
				Message: "Left channel/supergroup",
			}, nil
		}

		chatID, err := getChatIDFromDialogs(ctx, api, input.Channel)
		if err != nil {
			return nil, LeaveChannelOutput{
				Success: false,
				Message: fmt.Sprintf("Failed to find channel/group: %v", err),
			}, nil
		}

		_, err = api.MessagesDeleteChatUser(ctx, &tg.MessagesDeleteChatUserRequest{
			ChatID:        chatID,
			UserID:        &tg.InputUserSelf{},
			RevokeHistory: true,
		})
		if err != nil {
			return nil, LeaveChannelOutput{
				Success: false,
				Message: fmt.Sprintf("Failed to leave group: %v", err),
			}, nil
		}

		return nil, LeaveChannelOutput{
			Success: true,
			Message: "Left group",
		}, nil
	}
}

func getPeerFromDialogs(ctx context.Context, api *tg.Client, chat string) (tg.InputPeerClass, error) {
	dialogs, err := api.MessagesGetDialogs(ctx, &tg.MessagesGetDialogsRequest{
		OffsetPeer: &tg.InputPeerEmpty{},
		Limit:      200,
	})
	if err != nil {
		return nil, err
	}

	chatID, isNumeric := parseID(chat)

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

	return nil, fmt.Errorf("chat not found: %s", chat)
}

func getChannelFromDialogs(ctx context.Context, api *tg.Client, channel string) (*tg.InputChannel, error) {
	dialogs, err := api.MessagesGetDialogs(ctx, &tg.MessagesGetDialogsRequest{
		OffsetPeer: &tg.InputPeerEmpty{},
		Limit:      200,
	})
	if err != nil {
		return nil, err
	}

	channelID, isNumeric := parseID(channel)

	var chats []tg.ChatClass

	switch d := dialogs.(type) {
	case *tg.MessagesDialogs:
		chats = d.Chats
	case *tg.MessagesDialogsSlice:
		chats = d.Chats
	}

	for _, ch := range chats {
		if c, ok := ch.(*tg.Channel); ok {
			if isNumeric && c.ID == channelID {
				return &tg.InputChannel{ChannelID: c.ID, AccessHash: c.AccessHash}, nil
			}
			if !isNumeric && c.Username == channel {
				return &tg.InputChannel{ChannelID: c.ID, AccessHash: c.AccessHash}, nil
			}
		}
	}

	return nil, fmt.Errorf("channel not found: %s", channel)
}

func parseID(s string) (int64, bool) {
	id, err := strconv.ParseInt(s, 10, 64)
	return id, err == nil
}

func getChatIDFromDialogs(ctx context.Context, api *tg.Client, chat string) (int64, error) {
	dialogs, err := api.MessagesGetDialogs(ctx, &tg.MessagesGetDialogsRequest{
		OffsetPeer: &tg.InputPeerEmpty{},
		Limit:      200,
	})
	if err != nil {
		return 0, err
	}

	chatID, isNumeric := parseID(chat)

	var chats []tg.ChatClass

	switch d := dialogs.(type) {
	case *tg.MessagesDialogs:
		chats = d.Chats
	case *tg.MessagesDialogsSlice:
		chats = d.Chats
	}

	for _, ch := range chats {
		if c, ok := ch.(*tg.Chat); ok {
			if isNumeric && c.ID == chatID {
				return c.ID, nil
			}
			if !isNumeric && c.Title == chat {
				return c.ID, nil
			}
		}
	}

	return 0, fmt.Errorf("group not found: %s", chat)
}

func RegisterManageTools(server *mcp.Server, c *client.Client) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "delete_chat",
		Description: "Delete a chat/dialog by username or ID (removes chat history)",
	}, DeleteChat(c))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "leave_channel",
		Description: "Leave a channel or group by username or ID",
	}, LeaveChannel(c))
}
