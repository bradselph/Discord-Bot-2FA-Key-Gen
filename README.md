# Discord 2FA Bot

A secure Discord bot for generating Time-based One-Time Passwords (TOTP) and 2FA secret keys with QR code support.

## Description

This Discord bot provides a convenient way to generate 2FA verification codes and secret keys directly within Discord.Features role-based permissions, rate limiting. The bot supports generating new TOTP secrets with QR codes and creating verification codes from existing secrets.

## Installation
### Prerequisites

- Go 1.24
- Discord Bot Token
- Discord Server with appropriate permissions

### Setup

1. Clone the repository:
```bash
git clone https://github.com/bradselph/Discord-Bot-2FA-Key-Gen.git
cd Discord-Bot-2FA-Key-Gen
```

2. Install dependencies:
```bash
go mod tidy
```

3. Create a `.env` file:
```bash
cp .env.example .env
```

4. Configure your environment variables in `.env`:
```env
DISCORD_BOT_TOKEN=your_bot_token_here
DEV_USER_ID=your_discord_user_id
GUILD_ID=your_server_id_here
ALLOWED_ROLES=role_id_1,role_id_2,role_id_3
LOG_LEVEL=INFO
COMMAND_COOLDOWN=5
```

5. Build and run:
```bash
go build -o Discord-Bot-2FA-Key-Gen
./Discord-Bot-2FA-Key-Gen
```

## Configuration

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `DISCORD_BOT_TOKEN` | Your Discord bot token | - | Yes |
| `DEV_USER_ID` | Discord user ID with admin access | - | No |
| `GUILD_ID` | Discord server ID (leave empty for global commands) | - | No |
| `ALLOWED_ROLES` | Comma-separated role IDs that can use the bot | - | No |
| `LOG_LEVEL` | Logging level (DEBUG, INFO, WARN, ERROR, FATAL) | INFO | No |
| `COMMAND_COOLDOWN` | Cooldown between commands in seconds | 5 | No |

## Discord Bot Setup

1. Go to the [Discord Developer Portal](https://discord.com/developers/applications)
2. Create a new application
3. Go to the "Bot" section and create a bot
4. Copy the bot token to your `.env` file
5. In "OAuth2" > "URL Generator":
    - Select scopes: `bot` and `applications.commands`
    - No additional bot permissions needed
6. Use the generated URL to invite the bot to your server

## Usage

The bot provides two slash commands:

### `/2fa-code`
Generate a verification code from an existing secret key.

**Parameters:**
- `secret` (required) - Your 2FA secret key in Base32 format

**Example:**
```
/2fa-code secret:JBSWY3DPEHPK3PXP
```

### `/2fa-generate`
Generate a new 2FA secret key with QR code.

**Parameters:**
- `issuer` (optional) - Service name (defaults to "Discord 2FA Bot")
- `account` (optional) - Account name (defaults to your Discord username)

**Example:**
```
/2fa-generate issuer:MyService account:john.doe
```
### Structure

```
Discord-2FA-Bot/
├── auth/           # Permission checking and authorization
├── bot/            # Discord bot handlers and cooldown management
├── config/         # Configuration loading and validation
├── logger/         # Structured logging system
├── totp/           # TOTP generation and QR code creation
├── main.go         # Application entry point
├── go.mod          # Go module dependencies
└── .env            # Environment configuration
```

## License

This project is licensed under the GNU Affero General Public License v3.0 (AGPL-3.0) - see the [LICENSE](LICENSE) file for details.

### AGPL-3.0 Summary

This license requires that if you run this software on a server and provide it as a service to others, you must provide the source code to users of that service. This ensures that any improvements or modifications to the bot remain open source.

## Disclaimer

This bot is provided for educational and convenience purposes. Users are responsible for the security of their 2FA secrets. Always verify generated codes with official authenticator applications before relying on them for important accounts.

## Acknowledgments

- [pquerna/otp](https://github.com/pquerna/otp) - TOTP implementation
- [skip2/go-qrcode](https://github.com/skip2/go-qrcode) - QR code generation
- [bwmarrin/discordgo](https://github.com/bwmarrin/discordgo) - Discord API wrapper
- [joho/godotenv](https://github.com/joho/godotenv) - Environment configuration