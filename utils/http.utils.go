package utils

import (
	"fmt"
	"path/filepath"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
)

// Response format that's going to be used globally
// IN => success status, message, return data
// OUT => gin.H http transferable object
func ResponseWrapper(success bool, message string, data any) gin.H {
	return gin.H{"success": success, "message": message, "data": data}
}

func AddTimeToDate(dateString string, endOfDay bool) string {
	timeStringToAdd := "00:00:00"
	if endOfDay {
		timeStringToAdd = "23:59:59"
	}
	return fmt.Sprintf("%s %s", dateString, timeStringToAdd)
}

func TimePtr(t time.Time) *time.Time {
	return &t
}

// GetProjectRoot returns the root directory of your project no matter where it's called from.
func GetProjectRoot() string {
	_, b, _, _ := runtime.Caller(0) // path to this file
	// go from current file's dir (e.g., util/) up to project root
	return filepath.Clean(filepath.Join(filepath.Dir(b), ".."))
}
