package cache

import (
	"aura/models"
	"sync"
	"time"
)

var MediuxUsers *MediuxUserCache

type MediuxUserCache struct {
	users          map[string]*models.MediuxUserInfo // Key: Username
	mu             sync.RWMutex
	LastFullUpdate int64
}

func Cache_NewMediuxUserCache() *MediuxUserCache {
	return &MediuxUserCache{
		users: make(map[string]*models.MediuxUserInfo),
	}
}

func init() {
	MediuxUsers = Cache_NewMediuxUserCache()
}

func (c *MediuxUserCache) StoreMediuxUsers(users []models.MediuxUserInfo) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for i := range users {
		user := &users[i]
		c.users[user.Username] = user
	}
	c.LastFullUpdate = time.Now().Unix()
}

func (c *MediuxUserCache) GetMediuxUsers() []models.MediuxUserInfo {
	c.mu.RLock()
	defer c.mu.RUnlock()

	userList := make([]models.MediuxUserInfo, 0, len(c.users))
	for _, user := range c.users {
		userList = append(userList, *user)
	}
	return userList
}

func (c *MediuxUserCache) GetMediuxUserByUsername(username string) (*models.MediuxUserInfo, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	user, exists := c.users[username]
	return user, exists
}
