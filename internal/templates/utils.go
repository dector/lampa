package templates

import (
	"fmt"
	"strconv"
	"time"
)

func FormatGenerationTime(s string) string {
	t, _ := time.Parse(time.RFC3339, s)
	return t.Format("January 2, 2006 at 15:04:05")
}

func FormatFileSize(sizeBytes string) string {
	size, err := strconv.Atoi(sizeBytes)
	if err != nil {
		return "??"
	}

	return fmt.Sprintf("%.2f MB", float64(size)/(1024*1024))
}
