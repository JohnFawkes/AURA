package mediux

import (
	"aura/logging"
	"context"
)

func PreloadMediuxUsers(ctx context.Context) {
	ctx, logAction := logging.AddSubActionToContext(ctx, "Preloading MediUX Users", logging.LevelTrace)
	defer logAction.Complete()

	users, Err := GetAllUsers(ctx)
	if Err.Message != "" {
		return
	}

	logAction.AppendResult("users_count", len(users))
}
