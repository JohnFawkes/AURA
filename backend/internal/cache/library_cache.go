package cache

import (
	"aura/internal/modals"
	"sort"
	"sync"
)

var LibraryCacheStore *LibraryCache

func init() {
	LibraryCacheStore = NewLibraryCache()
}

type LibraryCache struct {
	sections map[string]*modals.LibrarySection // Key: Library Title
	mu       sync.RWMutex
}

// NewLibraryCache creates a new LibraryCache instance
func NewLibraryCache() *LibraryCache {
	return &LibraryCache{
		sections: make(map[string]*modals.LibrarySection),
	}
}

// Updates a section in the cache, updating existing media items and appending new ones
func (c *LibraryCache) Update(section *modals.LibrarySection) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if existing, exists := c.sections[section.Title]; exists {
		// Update metadata
		existing.Type = section.Type
		existing.ID = section.ID

		// Create a map of existing items for O(1) lookup
		existingItems := make(map[string]*modals.MediaItem)
		for i := range existing.MediaItems {
			existingItems[existing.MediaItems[i].RatingKey] = &existing.MediaItems[i]
		}

		// Update existing items and collect new ones
		var newItems []modals.MediaItem
		for _, newItem := range section.MediaItems {
			if existingItem, found := existingItems[newItem.RatingKey]; found {
				// Update existing item
				*existingItem = newItem
			} else {
				// Collect new item for appending
				newItems = append(newItems, newItem)
			}
		}

		// Append new items
		existing.MediaItems = append(existing.MediaItems, newItems...)
		existing.TotalSize = len(existing.MediaItems)
	} else {
		// Create new section
		c.sections[section.Title] = section
	}
}

// Update specific media item in a section
func (c *LibraryCache) UpdateMediaItem(sectionTitle string, item *modals.MediaItem) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if section, exists := c.sections[sectionTitle]; exists {
		// Create a map of existing items for O(1) lookup
		existingItems := make(map[string]*modals.MediaItem)
		for i := range section.MediaItems {
			existingItems[section.MediaItems[i].RatingKey] = &section.MediaItems[i]
		}
		if existingItem, found := existingItems[item.RatingKey]; found {
			// Update existing item
			*existingItem = *item
		} else {
			// Append new item
			section.MediaItems = append(section.MediaItems, *item)
			section.TotalSize = len(section.MediaItems)
		}
	}
}

// Get retrieves a section by Title
func (c *LibraryCache) Get(title string) (*modals.LibrarySection, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	section, exists := c.sections[title]
	return section, exists
}

// GetAllSorted returns all sections sorted by Title
func (c *LibraryCache) GetAllSorted() []*modals.LibrarySection {
	c.mu.RLock()
	defer c.mu.RUnlock()

	sections := make([]*modals.LibrarySection, 0, len(c.sections))
	for _, section := range c.sections {
		sections = append(sections, section)
	}

	sort.Slice(sections, func(i, j int) bool {
		return sections[i].Title < sections[j].Title
	})

	return sections
}

// Remove removes a section from the cache by Title
func (c *LibraryCache) Remove(title string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.sections, title)
}

// Clear removes all sections from the cache
func (c *LibraryCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.sections = make(map[string]*modals.LibrarySection)
}

// GetMediaItemFromSection retrieves a media item by RatingKey from a specific section
func (c *LibraryCache) GetMediaItemFromSection(sectionTitle, ratingKey string) (*modals.MediaItem, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	section, exists := c.sections[sectionTitle]
	if !exists {
		return &modals.MediaItem{}, false
	}

	for _, item := range section.MediaItems {
		if item.RatingKey == ratingKey {
			return &item, true
		}
	}

	return &modals.MediaItem{}, false
}

// GetMediaItemFromSectionByTMDBID retrieves a media item by TMDB ID from a specific section
func (c *LibraryCache) GetMediaItemFromSectionByTMDBID(sectionTitle, tmdbID string) (*modals.MediaItem, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	section, exists := c.sections[sectionTitle]
	if !exists {
		return &modals.MediaItem{}, false
	}

	var newestItem *modals.MediaItem
	for _, item := range section.MediaItems {
		for _, guid := range item.Guids {
			if guid.Provider == "tmdb" && guid.ID == tmdbID {
				if newestItem == nil || item.UpdatedAt > newestItem.UpdatedAt {
					newestItem = &item
				}
			}
		}
	}

	if newestItem != nil {
		return newestItem, true
	}
	return &modals.MediaItem{}, false
}

// GetTMDBIDFromMediaItemRatingKey retrieves the TMDB ID from a media item by its RatingKey
func (c *LibraryCache) GetTMDBIDFromMediaItemRatingKey(sectionTitle, ratingKey string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	section, exists := c.sections[sectionTitle]
	if !exists {
		return "", false
	}

	for _, item := range section.MediaItems {
		if item.RatingKey == ratingKey {
			for _, guid := range item.Guids {
				if guid.Provider == "tmdb" {
					return guid.ID, true
				}
			}
			break // No need to continue if we found the item
		}
	}

	return "", false
}

// Print all media item titles in the cache for a specific section
func (c *LibraryCache) PrintMediaItemsInSection(sectionTitle string) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	section, exists := c.sections[sectionTitle]
	if !exists {
		return
	}

	for _, item := range section.MediaItems {
		println(item.Title)
	}
}

// IsEmpty checks if the cache is empty
func (c *LibraryCache) IsEmpty() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.sections) == 0
}
