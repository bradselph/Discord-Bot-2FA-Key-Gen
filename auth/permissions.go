package auth

import (
	"Discord-Bot-2FA-Key-Gen/config"
	"Discord-Bot-2FA-Key-Gen/logger"

	"github.com/bwmarrin/discordgo"
)

type PermissionChecker struct {
	config *config.Config
}

func NewPermissionChecker(cfg *config.Config) *PermissionChecker {
	return &PermissionChecker{
		config: cfg,
	}
}

func (p *PermissionChecker) HasPermission(_ *discordgo.Session, i *discordgo.InteractionCreate) bool {
	if i.Member == nil || i.Member.User == nil {
		logger.Warn("No member or user information in interaction")
		return false
	}

	userID := i.Member.User.ID

	if userID == p.config.DevUserID {
		logger.Debug("Dev user access granted:", userID)
		return true
	}

	if len(p.config.AllowedRoles) == 0 {
		logger.Debug("No role restrictions configured, allowing access")
		return true
	}

	if i.Member.Roles == nil {
		logger.Debug("User has no roles:", userID)
		return false
	}

	userRoles := make(map[string]bool)
	for _, roleID := range i.Member.Roles {
		if roleID != "" {
			userRoles[roleID] = true
		}
	}

	for _, allowedRole := range p.config.AllowedRoles {
		if allowedRole != "" && userRoles[allowedRole] {
			logger.Debug("User has allowed role:", userID, allowedRole)
			return true
		}
	}

	logger.Warn("Access denied for user:", userID, "- missing required roles")
	return false
}

func (p *PermissionChecker) LogUnauthorizedAccess(userID, username, command string) {
	logger.Warn("Unauthorized access attempt:",
		"UserID:", userID,
		"Username:", username,
		"Command:", command,
	)
}
