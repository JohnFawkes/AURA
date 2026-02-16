package utils

import (
	"aura/logging"
	"context"
	"os"
	"strconv"
	"syscall"
)

func SetUMask(ctx context.Context) {
	_, logAction := logging.AddSubActionToContext(ctx, "Setting UMASK from Environment Variable", logging.LevelDebug)
	defer logAction.Complete()

	if umaskStr := os.Getenv("UMASK"); umaskStr != "" {
		if umask, err := strconv.ParseInt(umaskStr, 8, 0); err == nil {
			syscall.Umask(int(umask))
			logAction.AppendResult("umask_set", true)
			logAction.AppendResult("umask_value", umaskStr)
		}
	} else {
		logAction.AppendResult("umask_set", false)
	}
}
