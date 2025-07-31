package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"Discord-Bot-2FA-Key-Gen/auth"
	"Discord-Bot-2FA-Key-Gen/bot"
	"Discord-Bot-2FA-Key-Gen/config"
	"Discord-Bot-2FA-Key-Gen/logger"
	"Discord-Bot-2FA-Key-Gen/totp"

	"github.com/bwmarrin/discordgo"
)

func main() {
	cfg := config.Load()

	logger.Init(cfg.LogLevel)
	logger.Info("Starting 2FA Discord Bot...")

	if cfg.DiscordToken == "" {
		logger.Fatal("DISCORD_BOT_TOKEN is required")
	}

	totpGen := totp.New()
	permChecker := auth.NewPermissionChecker(cfg)
	cooldownDuration := time.Duration(cfg.CommandCooldown) * time.Second
	commandHandler := bot.NewCommandHandler(totpGen, permChecker, cooldownDuration)

	dg, err := discordgo.New("Bot " + cfg.DiscordToken)
	if err != nil {
		logger.Fatal("Error creating Discord session:", err)
	}

	dg.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		logger.Info("Bot is ready! Logged in as:", r.User.Username)
	})

	dg.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		handleInteraction(s, i, commandHandler)
	})

	dg.Identify.Intents = discordgo.IntentsGuilds

	err = dg.Open()
	if err != nil {
		logger.Fatal("Error opening Discord connection:", err)
	}
	defer func() {
		if err := dg.Close(); err != nil {
			logger.Error("Error closing Discord connection:", err)
		}
	}()

	if err := registerCommands(dg, cfg.GuildID); err != nil {
		logger.Fatal("Failed to register commands:", err)
	}

	commandHandler.StartCleanupRoutine()

	logger.Info("Bot is running. Press Ctrl+C to exit.")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-stop

	logger.Info("Shutting down bot...")
}

func handleInteraction(s *discordgo.Session, i *discordgo.InteractionCreate, handler *bot.CommandHandler) {
	defer func() {
		if r := recover(); r != nil {
			logger.Error("Panic in interaction handler:", r)
		}
	}()

	switch i.ApplicationCommandData().Name {
	case "2fa-code":
		handler.Handle2FACode(s, i)
	case "2fa-generate":
		handler.Handle2FAGenerate(s, i)
	}
}

func registerCommands(s *discordgo.Session, guildID string) error {
	commands := []*discordgo.ApplicationCommand{
		{
			Name:        "2fa-code",
			Description: "Generate a 2FA verification code from your secret key",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "secret",
					Description: "Your 2FA secret key (Base32 format)",
					Required:    true,
				},
			},
		},
		{
			Name:        "2fa-generate",
			Description: "Generate a new 2FA secret key with QR code",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "issuer",
					Description: "Service name (optional, defaults to 'Discord 2FA Bot')",
					Required:    false,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "account",
					Description: "Account name (optional, defaults to your username)",
					Required:    false,
				},
			},
		},
	}

	for _, command := range commands {
		_, err := s.ApplicationCommandCreate(s.State.User.ID, guildID, command)
		if err != nil {
			logger.Error("Cannot create command", command.Name+":", err)
			return err
		}
		logger.Info("Registered command:", command.Name)
	}

	return nil
}
