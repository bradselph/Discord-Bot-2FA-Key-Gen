package totp

import (
	"crypto/rand"
	"encoding/base32"
	"fmt"
	"strings"
	"time"

	"Discord-Bot-2FA-Key-Gen/logger"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"github.com/skip2/go-qrcode"
)

type Generator struct{}

type Result struct {
	Code          string
	RemainingTime int
	ValidUntil    time.Time
	Secret        string
	QRCode        []byte
	URI           string
}

type SecretResult struct {
	Secret string
	QRCode []byte
	URI    string
}

func New() *Generator {
	return &Generator{}
}

func (t *Generator) GenerateSecret(issuer, accountName string) (*SecretResult, error) {
	if issuer == "" {
		issuer = "Discord 2FA Bot"
	}
	if accountName == "" {
		accountName = "User"
	}

	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      issuer,
		AccountName: accountName,
		Algorithm:   otp.AlgorithmSHA1,
		Digits:      otp.DigitsSix,
		Period:      30,
	})
	if err != nil {
		logger.Error("Failed to generate TOTP key:", err)
		return nil, fmt.Errorf("failed to generate secret key")
	}

	qrCode, err := qrcode.Encode(key.URL(), qrcode.Medium, 256)
	if err != nil {
		logger.Error("Failed to generate QR code:", err)
		return nil, fmt.Errorf("failed to generate QR code")
	}

	logger.Info("Generated new TOTP secret")

	return &SecretResult{
		Secret: key.Secret(),
		QRCode: qrCode,
		URI:    key.URL(),
	}, nil
}

func (t *Generator) ValidateSecret(secret string) error {
	if secret == "" {
		return fmt.Errorf("secret key cannot be empty")
	}

	secret = t.normalizeSecret(secret)

	if len(secret) < 16 {
		return fmt.Errorf("secret key too short (minimum 16 characters)")
	}
	if len(secret) > 128 {
		return fmt.Errorf("secret key too long (maximum 128 characters)")
	}

	if !t.isValidBase32(secret) {
		return fmt.Errorf("invalid Base32 format (only A-Z and 2-7 allowed)")
	}

	if strings.Count(secret, "=") > 6 {
		return fmt.Errorf("invalid Base32 padding")
	}

	_, err := base32.StdEncoding.DecodeString(secret)
	if err != nil {
		return fmt.Errorf("failed to decode Base32 secret")
	}

	return nil
}

func (t *Generator) GenerateCode(secret string) (*Result, error) {
	if err := t.ValidateSecret(secret); err != nil {
		logger.Warn("Invalid secret validation:", err)
		return nil, err
	}

	secret = t.normalizeSecret(secret)

	now := time.Now()
	code, err := totp.GenerateCode(secret, now)
	if err != nil {
		logger.Error("Failed to generate TOTP code:", err)
		return nil, fmt.Errorf("failed to generate verification code")
	}

	remainingSeconds := 30 - int(now.Unix()%30)
	if remainingSeconds <= 0 {
		remainingSeconds = 30
	}

	validUntil := now.Add(time.Duration(remainingSeconds) * time.Second)

	uri := fmt.Sprintf("otpauth://totp/Discord-2FA-Bot:User?secret=%s&issuer=Discord-2FA-Bot", secret)
	qrCode, err := qrcode.Encode(uri, qrcode.Medium, 256)
	if err != nil {
		logger.Warn("Failed to generate QR code:", err)
		qrCode = nil
	}

	logger.Debug("Generated TOTP code:", code, "Valid for:", remainingSeconds, "seconds")

	return &Result{
		Code:          code,
		RemainingTime: remainingSeconds,
		ValidUntil:    validUntil,
		Secret:        secret,
		QRCode:        qrCode,
		URI:           uri,
	}, nil
}

func (t *Generator) normalizeSecret(secret string) string {
	secret = strings.ToUpper(strings.ReplaceAll(secret, " ", ""))
	secret = strings.ReplaceAll(secret, "-", "")
	secret = strings.ReplaceAll(secret, "_", "")
	return secret
}

func (t *Generator) isValidBase32(s string) bool {
	for _, c := range s {
		if !((c >= 'A' && c <= 'Z') || (c >= '2' && c <= '7') || c == '=') {
			return false
		}
	}
	return true
}

func (t *Generator) GenerateRandomSecret() (string, error) {
	bytes := make([]byte, 20)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return base32.StdEncoding.EncodeToString(bytes), nil
}
