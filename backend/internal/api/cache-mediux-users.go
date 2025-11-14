package api

import "sync"

var Global_Cache_MediuxUsers *MediuxUserCache

type MediuxUserCache struct {
	users map[string]*MediuxUserInfo // Key: Username
	mu    sync.RWMutex
}

func Cache_NewMediuxUserCache() *MediuxUserCache {
	return &MediuxUserCache{
		users: make(map[string]*MediuxUserInfo),
	}
}

func init() {
	Global_Cache_MediuxUsers = Cache_NewMediuxUserCache()
}

func (c *MediuxUserCache) StoreMediuxUsers(users []MediuxUserInfo) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for i := range users {
		user := &users[i]
		c.users[user.Username] = user
	}
}

func (c *MediuxUserCache) GetMediuxUserByUsername(username string) (*MediuxUserInfo, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	user, exists := c.users[username]
	return user, exists
}
