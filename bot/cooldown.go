package bot

import (
	"sync"
	"time"
)

type CooldownManager struct {
	cooldowns map[string]time.Time
	duration  time.Duration
	mutex     sync.RWMutex
}

func NewCooldownManager(duration time.Duration) *CooldownManager {
	return &CooldownManager{
		cooldowns: make(map[string]time.Time),
		duration:  duration,
	}
}

func (c *CooldownManager) IsOnCooldown(userID string) bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if lastUsed, exists := c.cooldowns[userID]; exists {
		return time.Since(lastUsed) < c.duration
	}
	return false
}

func (c *CooldownManager) SetCooldown(userID string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.cooldowns[userID] = time.Now()
}

func (c *CooldownManager) GetRemainingCooldown(userID string) time.Duration {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if lastUsed, exists := c.cooldowns[userID]; exists {
		elapsed := time.Since(lastUsed)
		if elapsed < c.duration {
			return c.duration - elapsed
		}
	}
	return 0
}

func (c *CooldownManager) CleanupExpired() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	now := time.Now()
	for userID, lastUsed := range c.cooldowns {
		if now.Sub(lastUsed) >= c.duration {
			delete(c.cooldowns, userID)
		}
	}
}
