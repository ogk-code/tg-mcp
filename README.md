# tg-mcp

MCP server that connects AI assistants to Telegram via MTProto API.


Built with [gotd/td](https://github.com/gotd/td) and [go-sdk](https://github.com/modelcontextprotocol/go-sdk).

## Features

- **Authentication**: Login with phone number and 2FA support
- **Messages**: Send and read messages
- **Chats**: List dialogs, get chat overview with recent messages
- **Management**: Leave channels/groups, delete chats

## Available Tools

| Tool | Description |
|------|-------------|
| `auth_status` | Check authorization status |
| `auth_send_code` | Send login code to phone |
| `auth_submit_code` | Submit code (with optional 2FA password) |
| `auth_logout` | Logout and clear session |
| `list_chats` | Get list of dialogs with unread counts |
| `get_chats_overview` | Get all chats with recent messages in one request |
| `get_messages` | Get messages from a specific chat |
| `send_message` | Send a message to a chat |
| `leave_channel` | Leave a channel or group |
| `delete_chat` | Delete a chat/dialog |

## Installation

### Prerequisites

1. Go 1.21+
2. Telegram API credentials from https://my.telegram.org/apps

### Build

```bash
git clone https://github.com/yourusername/tg-mcp.git
cd tg-mcp
go build -o tg-mcp .
```

Cross-compile for Windows:
```bash
GOOS=windows GOARCH=amd64 go build -o tg-mcp.exe .
```

## Configuration

### Claude Desktop

Add to `claude_desktop_config.json`:

**macOS**: `~/Library/Application Support/Claude/claude_desktop_config.json`

**Windows**: `%APPDATA%\Claude\claude_desktop_config.json`

```json
{
  "mcpServers": {
    "telegram": {
      "command": "/path/to/tg-mcp",
      "args": [],
      "env": {
        "TG_APP_ID": "your_app_id",
        "TG_APP_HASH": "your_app_hash"
      }
    }
  }
}
```

Windows example:
```json
{
  "mcpServers": {
    "telegram": {
      "command": "C:\\Users\\YourName\\tg-mcp.exe",
      "args": [],
      "env": {
        "TG_APP_ID": "your_app_id",
        "TG_APP_HASH": "your_app_hash"
      }
    }
  }
}
```

### Claude Code

Add to `~/.claude.json`:

```json
{
  "mcpServers": {
    "telegram": {
      "type": "stdio",
      "command": "/path/to/tg-mcp",
      "env": {
        "TG_APP_ID": "your_app_id",
        "TG_APP_HASH": "your_app_hash"
      }
    }
  }
}
```

### Open WebUI (via mcpo)

Use [mcpo](https://github.com/open-webui/mcpo) to expose as OpenAPI:

```bash
pip install mcpo
TG_APP_ID=your_id TG_APP_HASH=your_hash mcpo --port 8000 -- /path/to/tg-mcp
```

Then add `http://localhost:8000` as OpenAPI server in Open WebUI.

## Environment Variables

| Variable | Required | Description |
|----------|----------|-------------|
| `TG_APP_ID` | Yes | Telegram API ID |
| `TG_APP_HASH` | Yes | Telegram API Hash |
| `TG_SESSION_FILE` | No | Custom session file path (default: `~/.tg-mcp-session.json`) |

## Usage

After configuring, restart your MCP client (Claude Desktop, Claude Code, etc.).

### First-time Authorization

1. Use `auth_send_code` with your phone number
2. You'll receive a code in Telegram
3. Use `auth_submit_code` with the code (and password if 2FA is enabled)
4. Done! Session is saved for future use

### Example Prompts

- "Show me my recent Telegram chats"
- "Send a message to @username saying hello"
- "What are my unread messages?"
- "Leave the channel @spamchannel"

## Security

- **Session file** (`~/.tg-mcp-session.json`) contains your auth key — keep it private!
- **APP_ID/APP_HASH** are not sensitive — they identify the app, not your account
- Uses MTProto (user API), not Bot API — full account access

## License

MIT
