package tools

import (
	"context"
	"fmt"

	"github.com/gotd/td/tg"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"tg-mcp/client"
)

type CreateChannelInput struct {
	Title     string `json:"title"`
	About     string `json:"about,omitempty"`
	Broadcast bool   `json:"broadcast"`
}

type CreateChannelOutput struct {
	Success   bool   `json:"success"`
	ChannelID int64  `json:"channel_id,omitempty"`
	Message   string `json:"message,omitempty"`
}

func CreateChannel(c *client.Client) func(ctx context.Context, req *mcp.CallToolRequest, input CreateChannelInput) (*mcp.CallToolResult, CreateChannelOutput, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input CreateChannelInput) (*mcp.CallToolResult, CreateChannelOutput, error) {
		if !c.IsAuthorized() {
			return nil, CreateChannelOutput{
				Success: false,
				Message: "Not authorized",
			}, nil
		}

		api := c.API()
		if api == nil {
			return nil, CreateChannelOutput{
				Success: false,
				Message: "Client is not running",
			}, nil
		}

		updates, err := api.ChannelsCreateChannel(ctx, &tg.ChannelsCreateChannelRequest{
			Title:     input.Title,
			About:     input.About,
			Broadcast: input.Broadcast,
			Megagroup: !input.Broadcast,
		})
		if err != nil {
			return nil, CreateChannelOutput{
				Success: false,
				Message: fmt.Sprintf("Failed to create channel: %v", err),
			}, nil
		}

		var channelID int64
		if u, ok := updates.(*tg.Updates); ok {
			for _, chat := range u.Chats {
				if ch, ok := chat.(*tg.Channel); ok {
					channelID = ch.ID
					break
				}
			}
		}

		channelType := "group"
		if input.Broadcast {
			channelType = "channel"
		}

		return nil, CreateChannelOutput{
			Success:   true,
			ChannelID: channelID,
			Message:   fmt.Sprintf("Created %s: %s", channelType, input.Title),
		}, nil
	}
}

type EditChannelInput struct {
	Channel string `json:"channel"`
	Title   string `json:"title,omitempty"`
	About   string `json:"about,omitempty"`
}

type EditChannelOutput struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

func EditChannel(c *client.Client) func(ctx context.Context, req *mcp.CallToolRequest, input EditChannelInput) (*mcp.CallToolResult, EditChannelOutput, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input EditChannelInput) (*mcp.CallToolResult, EditChannelOutput, error) {
		if !c.IsAuthorized() {
			return nil, EditChannelOutput{
				Success: false,
				Message: "Not authorized",
			}, nil
		}

		api := c.API()
		if api == nil {
			return nil, EditChannelOutput{
				Success: false,
				Message: "Client is not running",
			}, nil
		}

		inputChannel, err := getChannelFromDialogs(ctx, api, input.Channel)
		if err != nil {
			return nil, EditChannelOutput{
				Success: false,
				Message: fmt.Sprintf("Failed to find channel: %v", err),
			}, nil
		}

		if input.Title != "" {
			_, err = api.ChannelsEditTitle(ctx, &tg.ChannelsEditTitleRequest{
				Channel: inputChannel,
				Title:   input.Title,
			})
			if err != nil {
				return nil, EditChannelOutput{
					Success: false,
					Message: fmt.Sprintf("Failed to edit title: %v", err),
				}, nil
			}
		}

		if input.About != "" {
			_, err = api.MessagesEditChatAbout(ctx, &tg.MessagesEditChatAboutRequest{
				Peer:  &tg.InputPeerChannel{ChannelID: inputChannel.ChannelID, AccessHash: inputChannel.AccessHash},
				About: input.About,
			})
			if err != nil {
				return nil, EditChannelOutput{
					Success: false,
					Message: fmt.Sprintf("Failed to edit about: %v", err),
				}, nil
			}
		}

		return nil, EditChannelOutput{
			Success: true,
			Message: "Channel updated",
		}, nil
	}
}

type DeleteChannelInput struct {
	Channel string `json:"channel"`
}

type DeleteChannelOutput struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

func DeleteChannel(c *client.Client) func(ctx context.Context, req *mcp.CallToolRequest, input DeleteChannelInput) (*mcp.CallToolResult, DeleteChannelOutput, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input DeleteChannelInput) (*mcp.CallToolResult, DeleteChannelOutput, error) {
		if !c.IsAuthorized() {
			return nil, DeleteChannelOutput{
				Success: false,
				Message: "Not authorized",
			}, nil
		}

		api := c.API()
		if api == nil {
			return nil, DeleteChannelOutput{
				Success: false,
				Message: "Client is not running",
			}, nil
		}

		inputChannel, err := getChannelFromDialogs(ctx, api, input.Channel)
		if err != nil {
			return nil, DeleteChannelOutput{
				Success: false,
				Message: fmt.Sprintf("Failed to find channel: %v", err),
			}, nil
		}

		_, err = api.ChannelsDeleteChannel(ctx, inputChannel)
		if err != nil {
			return nil, DeleteChannelOutput{
				Success: false,
				Message: fmt.Sprintf("Failed to delete channel: %v", err),
			}, nil
		}

		return nil, DeleteChannelOutput{
			Success: true,
			Message: "Channel deleted",
		}, nil
	}
}

type SetChannelUsernameInput struct {
	Channel  string `json:"channel"`
	Username string `json:"username"`
}

type SetChannelUsernameOutput struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

func SetChannelUsername(c *client.Client) func(ctx context.Context, req *mcp.CallToolRequest, input SetChannelUsernameInput) (*mcp.CallToolResult, SetChannelUsernameOutput, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input SetChannelUsernameInput) (*mcp.CallToolResult, SetChannelUsernameOutput, error) {
		if !c.IsAuthorized() {
			return nil, SetChannelUsernameOutput{
				Success: false,
				Message: "Not authorized",
			}, nil
		}

		api := c.API()
		if api == nil {
			return nil, SetChannelUsernameOutput{
				Success: false,
				Message: "Client is not running",
			}, nil
		}

		inputChannel, err := getChannelFromDialogs(ctx, api, input.Channel)
		if err != nil {
			return nil, SetChannelUsernameOutput{
				Success: false,
				Message: fmt.Sprintf("Failed to find channel: %v", err),
			}, nil
		}

		_, err = api.ChannelsUpdateUsername(ctx, &tg.ChannelsUpdateUsernameRequest{
			Channel:  inputChannel,
			Username: input.Username,
		})
		if err != nil {
			return nil, SetChannelUsernameOutput{
				Success: false,
				Message: fmt.Sprintf("Failed to set username: %v", err),
			}, nil
		}

		return nil, SetChannelUsernameOutput{
			Success: true,
			Message: fmt.Sprintf("Username set to @%s", input.Username),
		}, nil
	}
}

type InviteToChannelInput struct {
	Channel string   `json:"channel"`
	Users   []string `json:"users"`
}

type InviteToChannelOutput struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

func InviteToChannel(c *client.Client) func(ctx context.Context, req *mcp.CallToolRequest, input InviteToChannelInput) (*mcp.CallToolResult, InviteToChannelOutput, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input InviteToChannelInput) (*mcp.CallToolResult, InviteToChannelOutput, error) {
		if !c.IsAuthorized() {
			return nil, InviteToChannelOutput{
				Success: false,
				Message: "Not authorized",
			}, nil
		}

		api := c.API()
		if api == nil {
			return nil, InviteToChannelOutput{
				Success: false,
				Message: "Client is not running",
			}, nil
		}

		inputChannel, err := getChannelFromDialogs(ctx, api, input.Channel)
		if err != nil {
			return nil, InviteToChannelOutput{
				Success: false,
				Message: fmt.Sprintf("Failed to find channel: %v", err),
			}, nil
		}

		var inputUsers []tg.InputUserClass
		for _, user := range input.Users {
			resolved, err := api.ContactsResolveUsername(ctx, &tg.ContactsResolveUsernameRequest{
				Username: user,
			})
			if err != nil {
				continue
			}
			for _, u := range resolved.Users {
				if usr, ok := u.(*tg.User); ok {
					inputUsers = append(inputUsers, &tg.InputUser{
						UserID:     usr.ID,
						AccessHash: usr.AccessHash,
					})
				}
			}
		}

		if len(inputUsers) == 0 {
			return nil, InviteToChannelOutput{
				Success: false,
				Message: "No valid users found",
			}, nil
		}

		_, err = api.ChannelsInviteToChannel(ctx, &tg.ChannelsInviteToChannelRequest{
			Channel: inputChannel,
			Users:   inputUsers,
		})
		if err != nil {
			return nil, InviteToChannelOutput{
				Success: false,
				Message: fmt.Sprintf("Failed to invite users: %v", err),
			}, nil
		}

		return nil, InviteToChannelOutput{
			Success: true,
			Message: fmt.Sprintf("Invited %d users", len(inputUsers)),
		}, nil
	}
}

type GetChannelInfoInput struct {
	Channel string `json:"channel"`
}

type ChannelInfo struct {
	ID         int64  `json:"id"`
	Title      string `json:"title"`
	Username   string `json:"username,omitempty"`
	About      string `json:"about,omitempty"`
	Members    int    `json:"members"`
	Admins     int    `json:"admins,omitempty"`
	Broadcast  bool   `json:"broadcast"`
	Verified   bool   `json:"verified"`
	Restricted bool   `json:"restricted"`
	InviteLink string `json:"invite_link,omitempty"`
}

type GetChannelInfoOutput struct {
	Success bool        `json:"success"`
	Channel ChannelInfo `json:"channel,omitempty"`
	Message string      `json:"message,omitempty"`
}

func GetChannelInfo(c *client.Client) func(ctx context.Context, req *mcp.CallToolRequest, input GetChannelInfoInput) (*mcp.CallToolResult, GetChannelInfoOutput, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input GetChannelInfoInput) (*mcp.CallToolResult, GetChannelInfoOutput, error) {
		if !c.IsAuthorized() {
			return nil, GetChannelInfoOutput{
				Success: false,
				Message: "Not authorized",
			}, nil
		}

		api := c.API()
		if api == nil {
			return nil, GetChannelInfoOutput{
				Success: false,
				Message: "Client is not running",
			}, nil
		}

		inputChannel, err := getChannelFromDialogs(ctx, api, input.Channel)
		if err != nil {
			return nil, GetChannelInfoOutput{
				Success: false,
				Message: fmt.Sprintf("Failed to find channel: %v", err),
			}, nil
		}

		fullChannel, err := api.ChannelsGetFullChannel(ctx, inputChannel)
		if err != nil {
			return nil, GetChannelInfoOutput{
				Success: false,
				Message: fmt.Sprintf("Failed to get channel info: %v", err),
			}, nil
		}

		info := ChannelInfo{}

		if fc, ok := fullChannel.FullChat.(*tg.ChannelFull); ok {
			info.About = fc.About
			info.Members = fc.ParticipantsCount
			info.Admins = fc.AdminsCount
			if fc.ExportedInvite != nil {
				if invite, ok := fc.ExportedInvite.(*tg.ChatInviteExported); ok {
					info.InviteLink = invite.Link
				}
			}
		}

		for _, chat := range fullChannel.Chats {
			if ch, ok := chat.(*tg.Channel); ok {
				info.ID = ch.ID
				info.Title = ch.Title
				info.Username = ch.Username
				info.Broadcast = ch.Broadcast
				info.Verified = ch.Verified
				info.Restricted = ch.Restricted
				break
			}
		}

		return nil, GetChannelInfoOutput{
			Success: true,
			Channel: info,
		}, nil
	}
}

type ExportInviteLinkInput struct {
	Channel string `json:"channel"`
}

type ExportInviteLinkOutput struct {
	Success bool   `json:"success"`
	Link    string `json:"link,omitempty"`
	Message string `json:"message,omitempty"`
}

func ExportInviteLink(c *client.Client) func(ctx context.Context, req *mcp.CallToolRequest, input ExportInviteLinkInput) (*mcp.CallToolResult, ExportInviteLinkOutput, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input ExportInviteLinkInput) (*mcp.CallToolResult, ExportInviteLinkOutput, error) {
		if !c.IsAuthorized() {
			return nil, ExportInviteLinkOutput{
				Success: false,
				Message: "Not authorized",
			}, nil
		}

		api := c.API()
		if api == nil {
			return nil, ExportInviteLinkOutput{
				Success: false,
				Message: "Client is not running",
			}, nil
		}

		peer, err := getPeerFromDialogs(ctx, api, input.Channel)
		if err != nil {
			return nil, ExportInviteLinkOutput{
				Success: false,
				Message: fmt.Sprintf("Failed to find channel: %v", err),
			}, nil
		}

		exported, err := api.MessagesExportChatInvite(ctx, &tg.MessagesExportChatInviteRequest{
			Peer: peer,
		})
		if err != nil {
			return nil, ExportInviteLinkOutput{
				Success: false,
				Message: fmt.Sprintf("Failed to export invite link: %v", err),
			}, nil
		}

		var link string
		if invite, ok := exported.(*tg.ChatInviteExported); ok {
			link = invite.Link
		}

		return nil, ExportInviteLinkOutput{
			Success: true,
			Link:    link,
		}, nil
	}
}

type GetChannelMembersInput struct {
	Channel string `json:"channel"`
	Limit   int    `json:"limit,omitempty"`
	Offset  int    `json:"offset,omitempty"`
	Filter  string `json:"filter,omitempty"`
}

type ChannelMember struct {
	ID        int64  `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name,omitempty"`
	Username  string `json:"username,omitempty"`
	Bot       bool   `json:"bot"`
	Admin     bool   `json:"admin"`
	Creator   bool   `json:"creator"`
}

type GetChannelMembersOutput struct {
	Success bool            `json:"success"`
	Members []ChannelMember `json:"members,omitempty"`
	Total   int             `json:"total,omitempty"`
	Message string          `json:"message,omitempty"`
}

func GetChannelMembers(c *client.Client) func(ctx context.Context, req *mcp.CallToolRequest, input GetChannelMembersInput) (*mcp.CallToolResult, GetChannelMembersOutput, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input GetChannelMembersInput) (*mcp.CallToolResult, GetChannelMembersOutput, error) {
		if !c.IsAuthorized() {
			return nil, GetChannelMembersOutput{
				Success: false,
				Message: "Not authorized",
			}, nil
		}

		api := c.API()
		if api == nil {
			return nil, GetChannelMembersOutput{
				Success: false,
				Message: "Client is not running",
			}, nil
		}

		inputChannel, err := getChannelFromDialogs(ctx, api, input.Channel)
		if err != nil {
			return nil, GetChannelMembersOutput{
				Success: false,
				Message: fmt.Sprintf("Failed to find channel: %v", err),
			}, nil
		}

		limit := input.Limit
		if limit <= 0 {
			limit = 100
		}
		if limit > 200 {
			limit = 200
		}

		var filter tg.ChannelParticipantsFilterClass
		switch input.Filter {
		case "admins":
			filter = &tg.ChannelParticipantsAdmins{}
		case "bots":
			filter = &tg.ChannelParticipantsBots{}
		case "banned":
			filter = &tg.ChannelParticipantsKicked{Q: ""}
		case "restricted":
			filter = &tg.ChannelParticipantsBanned{Q: ""}
		default:
			filter = &tg.ChannelParticipantsRecent{}
		}

		participants, err := api.ChannelsGetParticipants(ctx, &tg.ChannelsGetParticipantsRequest{
			Channel: inputChannel,
			Filter:  filter,
			Offset:  input.Offset,
			Limit:   limit,
		})
		if err != nil {
			return nil, GetChannelMembersOutput{
				Success: false,
				Message: fmt.Sprintf("Failed to get members: %v", err),
			}, nil
		}

		cp, ok := participants.(*tg.ChannelsChannelParticipants)
		if !ok {
			return nil, GetChannelMembersOutput{
				Success: false,
				Message: "Unexpected response type",
			}, nil
		}

		userMap := make(map[int64]*tg.User)
		for _, u := range cp.Users {
			if user, ok := u.(*tg.User); ok {
				userMap[user.ID] = user
			}
		}

		adminSet := make(map[int64]bool)
		creatorSet := make(map[int64]bool)
		for _, p := range cp.Participants {
			switch pt := p.(type) {
			case *tg.ChannelParticipantCreator:
				creatorSet[pt.UserID] = true
				adminSet[pt.UserID] = true
			case *tg.ChannelParticipantAdmin:
				adminSet[pt.UserID] = true
			}
		}

		members := make([]ChannelMember, 0, len(cp.Participants))
		for _, p := range cp.Participants {
			var userID int64
			switch pt := p.(type) {
			case *tg.ChannelParticipant:
				userID = pt.UserID
			case *tg.ChannelParticipantSelf:
				userID = pt.UserID
			case *tg.ChannelParticipantCreator:
				userID = pt.UserID
			case *tg.ChannelParticipantAdmin:
				userID = pt.UserID
			case *tg.ChannelParticipantBanned:
				if peer, ok := pt.Peer.(*tg.PeerUser); ok {
					userID = peer.UserID
				}
			case *tg.ChannelParticipantLeft:
				if peer, ok := pt.Peer.(*tg.PeerUser); ok {
					userID = peer.UserID
				}
			}

			if userID == 0 {
				continue
			}

			user := userMap[userID]
			if user == nil {
				continue
			}

			members = append(members, ChannelMember{
				ID:        user.ID,
				FirstName: user.FirstName,
				LastName:  user.LastName,
				Username:  user.Username,
				Bot:       user.Bot,
				Admin:     adminSet[userID],
				Creator:   creatorSet[userID],
			})
		}

		return nil, GetChannelMembersOutput{
			Success: true,
			Members: members,
			Total:   cp.Count,
		}, nil
	}
}

func RegisterChannelsTools(server *mcp.Server, c *client.Client) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "create_channel",
		Description: "Create a new channel or supergroup. Set broadcast=true for channel, false for group.",
	}, CreateChannel(c))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "edit_channel",
		Description: "Edit channel/group title or description",
	}, EditChannel(c))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "delete_channel",
		Description: "Delete a channel or supergroup (irreversible)",
	}, DeleteChannel(c))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "set_channel_username",
		Description: "Set or change channel public username",
	}, SetChannelUsername(c))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "invite_to_channel",
		Description: "Invite users to a channel or group by their usernames",
	}, InviteToChannel(c))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_channel_info",
		Description: "Get detailed information about a channel or group",
	}, GetChannelInfo(c))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "export_invite_link",
		Description: "Export/create invite link for a channel or group",
	}, ExportInviteLink(c))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_channel_members",
		Description: "Get channel/group members. Filter: admins, bots, banned, restricted (default: recent). Supports pagination with offset/limit.",
	}, GetChannelMembers(c))
}
