package main

// Small logging helpers and package-level logger variables.
// Helpers avoid repetitive nil checks for `sugar` — they no-op when `sugar` is nil.
import (
	"fmt"
	"os"

	"go.uber.org/zap"
)

// zap logger and file (moved here for centralized logging)
var (
    logger  *zap.Logger
    sugar   *zap.SugaredLogger
    logFile *os.File
)

// logDebug logs a debug message built from args.
func logDebug(args ...interface{}) {
    if sugar != nil {
        sugar.Debug(args...)
    }
}

// logDebugf logs a formatted debug message.
func logDebugf(format string, args ...interface{}) {
    if sugar != nil {
        sugar.Debugf(format, args...)
    }
}

// logInfo logs an info message.
func logInfo(args ...interface{}) {
    if sugar != nil {
        sugar.Info(args...)
    }
}

// logErrorf logs an error formatted message (fallback to fmt if sugar missing).
func logErrorf(format string, args ...interface{}) {
    if sugar != nil {
        sugar.Errorf(format, args...)
    } else {
        // ensure at least something is printed when no logger configured
        fmt.Printf(format+"\n", args...)
    }
}
