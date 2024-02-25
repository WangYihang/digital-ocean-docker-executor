package http_grab

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/jszwec/csvutil"
)

type HTTPGrabProgress struct {
	Timestamp     int64 `csv:"timestamp" json:"timestamp"`
	FinishedTasks int64 `csv:"finished_tasks" json:"finished_tasks"`
	TotalTasks    int64 `csv:"total_tasks" json:"total_tasks"`
	Line          string
}

func NewHTTPGrabProgress(message string) *HTTPGrabProgress {
	content := strings.Join([]string{
		"timestamp,finished_tasks,total_tasks",
		strings.Replace(message, " ", "", -1),
	}, "\n")
	var progresses []HTTPGrabProgress
	if err := csvutil.Unmarshal([]byte(content), &progresses); err != nil {
		slog.Error("error occured while parsing progress", slog.String("error", err.Error()))
		return &HTTPGrabProgress{}
	}
	if len(progresses) == 0 {
		return &HTTPGrabProgress{}
	} else {
		progress := progresses[0]
		progress.Line = message
		return &progress
	}
}

func (z HTTPGrabProgress) String() string {
	return fmt.Sprintf("%s (%f%%)", z.Line, float64(z.FinishedTasks)/float64(z.TotalTasks)*100)
}

func (z HTTPGrabProgress) Done() bool {
	return z.FinishedTasks >= 13327232
}
