package api

import (
	"sort"
	"strconv"
	"strings"
	"sync"
)

// ---   Cache Global Variables (Backend Library Cache) --- ---
var Global_Cache_LibraryStore *LibraryCache

type LibraryCache struct {
	sections map[string]*LibrarySection // Key: Library Title
	mu       sync.RWMutex
}

// NewLibraryCache creates a new LibraryCache instance
func Cache_NewLibraryCache() *LibraryCache {
	return &LibraryCache{
		sections: make(map[string]*LibrarySection),
	}
}

func init() {
	Global_Cache_LibraryStore = Cache_NewLibraryCache()
}

// UpdateSection updates or adds a LibrarySection in the cache.
// If the section already exists, its metadata and media items are updated.
// New media items are appended to the section.
// If the section does not exist, it is added to the cache.
func (c *LibraryCache) UpdateSection(section *LibrarySection) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if section already exists
	// If it does, we need to update it
	if existing, exists := c.sections[section.Title]; exists {
		// Update metadata
		existing.Type = section.Type
		existing.ID = section.ID
		existing.Path = section.Path

		// Create a map of existing items for O(1) lookup
		existingItems := make(map[string]*MediaItem)
		for i := range existing.MediaItems {
			existingItems[existing.MediaItems[i].TMDB_ID] = &existing.MediaItems[i]
		}

		// Update existing items and collect new ones
		var newItems []MediaItem
		for _, newItem := range section.MediaItems {
			if existingItem, found := existingItems[newItem.TMDB_ID]; found {
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
		// If section does not exist, add it to the cache
		c.sections[section.Title] = section
	}
}

// UpdateMediaItem updates a specific media item in a section
func (c *LibraryCache) UpdateMediaItem(sectionTitle string, item *MediaItem) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if section exists
	if section, exists := c.sections[sectionTitle]; exists {
		// Create a map of existing items for O(1) lookup
		existingItems := make(map[string]*MediaItem)
		for i := range section.MediaItems {
			existingItems[section.MediaItems[i].TMDB_ID] = &section.MediaItems[i]
		}
		if existingItem, found := existingItems[item.TMDB_ID]; found {
			// Update existing item
			*existingItem = *item
		} else {
			// Append new item
			section.MediaItems = append(section.MediaItems, *item)
			section.TotalSize = len(section.MediaItems)
		}
	}
}

// GetSectionByTitle retrieves a section by Title
func (c *LibraryCache) GetSectionByTitle(title string) (*LibrarySection, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	section, exists := c.sections[title]
	return section, exists
}

// GetAllSectionsSortedByTitle returns all sections sorted by Title
func (c *LibraryCache) GetAllSectionsSortedByTitle() []*LibrarySection {
	c.mu.RLock()
	defer c.mu.RUnlock()

	sections := make([]*LibrarySection, 0, len(c.sections))
	for _, section := range c.sections {
		sections = append(sections, section)
	}

	sort.Slice(sections, func(i, j int) bool {
		return sections[i].Title < sections[j].Title
	})

	return sections
}

// RemoveSectionByTitle removes a section from the cache by Title
func (c *LibraryCache) RemoveSectionByTitle(title string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.sections, title)
}

// ClearAllSections removes all sections from the cache
func (c *LibraryCache) ClearAllSections() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.sections = make(map[string]*LibrarySection)
}

// GetMediaItemFromSectionByTMDBID retrieves a media item by TMDB ID from a specific section
func (c *LibraryCache) GetMediaItemFromSectionByTMDBID(sectionTitle, tmdbID string) (*MediaItem, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	section, exists := c.sections[sectionTitle]
	if !exists {
		return &MediaItem{}, false
	}

	var newestItem *MediaItem
	for _, item := range section.MediaItems {
		if item.TMDB_ID == tmdbID {
			newestItem = &item
			break
		}
	}

	if newestItem != nil {
		return newestItem, true
	}
	return &MediaItem{}, false
}

// GetMediaItemFromSectionByTitleAndYear retrieves the TMDB ID from a media item by its title and year
func (c *LibraryCache) GetMediaItemFromSectionByTitleAndYear(sectionTitle, itemTitle string, year int) (*MediaItem, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	section, exists := c.sections[sectionTitle]
	if !exists {
		return &MediaItem{}, false
	}

	cleanedSearchTitle := CleanStringForComparison(stripYearFromTitle(itemTitle))
	for _, item := range section.MediaItems {
		cleanedTitle := CleanStringForComparison(stripYearFromTitle(item.Title))
		if cleanedTitle == cleanedSearchTitle && item.Year == year {
			return &item, true
		}
	}

	return &MediaItem{}, false
}

// IsEmpty checks if the cache is empty
func (c *LibraryCache) IsEmpty() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.sections) == 0
}

func CleanStringForComparison(input string) string {
	var b strings.Builder
	input = strings.ToLower(input)
	for _, r := range input {
		switch r {
		case '-', '_', '.', ',', ':', ';', '!', '?', '\'', '(', ')', '[', ']', '{', '}':
			// skip these characters
			continue
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

func stripYearFromTitle(title string) string {
	parts := strings.Fields(title)
	if len(parts) > 1 {
		last := parts[len(parts)-1]
		if _, err := strconv.Atoi(last); err == nil {
			return strings.Join(parts[:len(parts)-1], " ")
		}
	}
	return title
}
