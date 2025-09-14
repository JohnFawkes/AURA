package config

import "aura/internal/modals"

// Global is a pointer to the global configuration instance.
// It is used throughout the application to access configuration settings.
var Global *modals.Config = &modals.Config{}

// By default, the configuration is invalid until it is validated.
var ConfigLoaded bool = false
var ConfigValid bool = false

var ConfigMediuxValid bool = true
var ConfigMediaServerValid bool = true
