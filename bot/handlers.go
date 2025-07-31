package bot

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"Discord-Bot-2FA-Key-Gen/auth"
	"Discord-Bot-2FA-Key-Gen/logger"
	"Discord-Bot-2FA-Key-Gen/totp"

	"github.com/bwmarrin/discordgo"
)

type CommandHandler struct {
	totpGen         *totp.Generator
	permChecker     *auth.PermissionChecker
	cooldownManager *CooldownManager
}

func NewCommandHandler(totpGen *totp.Generator, permChecker *auth.PermissionChecker, cooldownDuration time.Duration) *CommandHandler {
	return &CommandHandler{
		totpGen:         totpGen,
		permChecker:     permChecker,
		cooldownManager: NewCooldownManager(cooldownDuration),
	}
}

func (h *CommandHandler) Handle2FACode(s *discordgo.Session, i *discordgo.InteractionCreate) {
	defer func() {
		if r := recover(); r != nil {
			logger.Error("Panic in 2FA code handler:", r)
			h.respondWithError(s, i, "An unexpected error occurred. Please try again later.")
		}
	}()

	if !h.validateInteraction(s, i) {
		return
	}

	userID := i.Member.User.ID
	username := i.Member.User.Username

	if !h.permChecker.HasPermission(s, i) {
		h.permChecker.LogUnauthorizedAccess(userID, username, "2fa-code")
		h.respondWithError(s, i, "You don't have permission to use this command.")
		return
	}

	if h.cooldownManager.IsOnCooldown(userID) {
		remaining := h.cooldownManager.GetRemainingCooldown(userID)
		h.respondWithError(s, i, fmt.Sprintf("Please wait %d seconds before using this command again.", int(remaining.Seconds())))
		return
	}

	options := i.ApplicationCommandData().Options
	if len(options) == 0 {
		h.respondWithError(s, i, "Please provide a 2FA secret key.")
		return
	}

	secret := strings.TrimSpace(options[0].StringValue())
	if secret == "" {
		h.respondWithError(s, i, "Secret key cannot be empty.")
		return
	}

	if len(secret) > 256 {
		h.respondWithError(s, i, "Secret key is too long.")
		return
	}

	result, err := h.totpGen.GenerateCode(secret)
	if err != nil {
		logger.Warn("TOTP code generation failed for user:", userID, "Error:", err)
		h.respondWithError(s, i, err.Error())
		return
	}

	h.cooldownManager.SetCooldown(userID)

	embed := &discordgo.MessageEmbed{
		Title: "2FA Verification Code",
		Color: 0x32AE4D,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Secret Key",
				Value:  fmt.Sprintf("||%s||", result.Secret),
				Inline: false,
			},
			{
				Name:   "Current Code",
				Value:  fmt.Sprintf("**`%s`**", result.Code),
				Inline: true,
			},
			{
				Name:   "Remaining Time",
				Value:  fmt.Sprintf("%d seconds", result.RemainingTime),
				Inline: true,
			},
			{
				Name:   "Security Notice",
				Value:  "This code is valid for 30 seconds. Do not share it with anyone.",
				Inline: false,
			},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Code refreshes every 30 seconds",
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}

	response := &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
			Flags:  discordgo.MessageFlagsEphemeral,
		},
	}

	if result.QRCode != nil {
		response.Data.Files = []*discordgo.File{
			{
				Name:        "qrcode.png",
				ContentType: "image/png",
				Reader:      bytes.NewReader(result.QRCode),
			},
		}
		embed.Image = &discordgo.MessageEmbedImage{
			URL: "attachment://qrcode.png",
		}
	}

	err = s.InteractionRespond(i.Interaction, response)
	if err != nil {
		logger.Error("Failed to respond to interaction:", err)
		h.respondWithError(s, i, "Failed to send response.")
		return
	}

	logger.Info("2FA code generated for user:", username, "(", userID, ")")
}

func (h *CommandHandler) Handle2FAGenerate(s *discordgo.Session, i *discordgo.InteractionCreate) {
	defer func() {
		if r := recover(); r != nil {
			logger.Error("Panic in 2FA generate handler:", r)
			h.respondWithError(s, i, "An unexpected error occurred. Please try again later.")
		}
	}()

	if !h.validateInteraction(s, i) {
		return
	}

	userID := i.Member.User.ID
	username := i.Member.User.Username

	if !h.permChecker.HasPermission(s, i) {
		h.permChecker.LogUnauthorizedAccess(userID, username, "2fa-generate")
		h.respondWithError(s, i, "You don't have permission to use this command.")
		return
	}

	if h.cooldownManager.IsOnCooldown(userID) {
		remaining := h.cooldownManager.GetRemainingCooldown(userID)
		h.respondWithError(s, i, fmt.Sprintf("Please wait %d seconds before using this command again.", int(remaining.Seconds())))
		return
	}

	options := i.ApplicationCommandData().Options
	issuer := "Discord 2FA Bot"
	accountName := username

	for _, option := range options {
		switch option.Name {
		case "issuer":
			if option.StringValue() != "" {
				issuer = strings.TrimSpace(option.StringValue())
			}
		case "account":
			if option.StringValue() != "" {
				accountName = strings.TrimSpace(option.StringValue())
			}
		}
	}

	result, err := h.totpGen.GenerateSecret(issuer, accountName)
	if err != nil {
		logger.Error("Secret generation failed for user:", userID, "Error:", err)
		h.respondWithError(s, i, "Failed to generate secret key.")
		return
	}

	h.cooldownManager.SetCooldown(userID)

	embed := &discordgo.MessageEmbed{
		Title: "New 2FA Secret Generated",
		Color: 0x4CAF50,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Secret Key",
				Value:  fmt.Sprintf("||%s||", result.Secret),
				Inline: false,
			},
			{
				Name:   "Issuer",
				Value:  issuer,
				Inline: true,
			},
			{
				Name:   "Account",
				Value:  accountName,
				Inline: true,
			},
			{
				Name:   "Setup Instructions",
				Value:  "1. Scan the QR code with your authenticator app\n2. Or manually enter the secret key\n3. Use `/2fa-code` to generate verification codes",
				Inline: false,
			},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Keep your secret key safe and private",
		},
		Image: &discordgo.MessageEmbedImage{
			URL: "attachment://qrcode.png",
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}

	response := &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
			Flags:  discordgo.MessageFlagsEphemeral,
			Files: []*discordgo.File{
				{
					Name:        "qrcode.png",
					ContentType: "image/png",
					Reader:      bytes.NewReader(result.QRCode),
				},
			},
		},
	}

	err = s.InteractionRespond(i.Interaction, response)
	if err != nil {
		logger.Error("Failed to respond to interaction:", err)
		h.respondWithError(s, i, "Failed to send response.")
		return
	}

	logger.Info("2FA secret generated for user:", username, "(", userID, ")")
}

func (h *CommandHandler) validateInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) bool {
	if i.Member == nil || i.Member.User == nil {
		h.respondWithError(s, i, "Unable to verify user information.")
		return false
	}
	return true
}

func (h *CommandHandler) respondWithError(s *discordgo.Session, i *discordgo.InteractionCreate, message string) {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: message,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})

	if err != nil {
		logger.Error("Failed to send error response:", err)
	}
}

func (h *CommandHandler) StartCleanupRoutine() {
	ticker := time.NewTicker(5 * time.Minute)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("Panic in cleanup routine:", r)
			}
		}()

		for range ticker.C {
			h.cooldownManager.CleanupExpired()
			logger.Debug("Cleaned up expired cooldowns")
		}
	}()
}
