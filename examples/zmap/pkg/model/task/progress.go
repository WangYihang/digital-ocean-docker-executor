package zmap_task

import (
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/WangYihang/digital-ocean-docker-executor/pkg/model/task"
	"github.com/jszwec/csvutil"
)

type ZMapProgress struct {
	RealTime              string  `csv:"real-time" json:"real-time"`
	TimeElapsed           int64   `csv:"time-elapsed" json:"time-elapsed"`
	TimeRemaining         int64   `csv:"time-remaining" json:"time-remaining"`
	PercentComplete       float64 `csv:"percent-complete" json:"percent-complete"`
	HitRate               float64 `csv:"hit-rate" json:"hit-rate"`
	ActiveSendThreads     int64   `csv:"active-send-threads" json:"active-send-threads"`
	SentTotal             int64   `csv:"sent-total" json:"sent-total"`
	SentLastOneSec        int64   `csv:"sent-last-one-sec" json:"sent-last-one-sec"`
	SentAvgPerSec         int64   `csv:"sent-avg-per-sec" json:"sent-avg-per-sec"`
	RecvSuccessTotal      int64   `csv:"recv-success-total" json:"recv-success-total"`
	RecvSuccessLastOneSec int64   `csv:"recv-success-last-one-sec" json:"recv-success-last-one-sec"`
	RecvSuccessAvgPerSec  int64   `csv:"recv-success-avg-per-sec" json:"recv-success-avg-per-sec"`
	RecvTotal             int64   `csv:"recv-total" json:"recv-total"`
	RecvTotalLastOneSec   int64   `csv:"recv-total-last-one-sec" json:"recv-total-last-one-sec"`
	RecvTotalAvgPerSec    int64   `csv:"recv-total-avg-per-sec" json:"recv-total-avg-per-sec"`
	PcapDropTotal         int64   `csv:"pcap-drop-total" json:"pcap-drop-total"`
	DropLastOneSec        int64   `csv:"drop-last-one-sec" json:"drop-last-one-sec"`
	DropAvgPerSec         int64   `csv:"drop-avg-per-sec" json:"drop-avg-per-sec"`
	SendtoFailTotal       int64   `csv:"sendto-fail-total" json:"sendto-fail-total"`
	SendtoFailLastOneSec  int64   `csv:"sendto-fail-last-one-sec" json:"sendto-fail-last-one-sec"`
	SendtoFailAvgPerSec   int64   `csv:"sendto-fail-avg-per-sec" json:"sendto-fail-avg-per-sec"`
	Line                  string
}

var PendingProgress *ZMapProgress
var DoneProgress *ZMapProgress

func init() {
	PendingProgress, _ = NewZMapProgress("0000-00-00 00:00:00,0,0,0,0.000000,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0")
	DoneProgress, _ = NewZMapProgress("0000-00-00 00:00:00,0,0,100,0.000000,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0")
}

func NewZMapProgress(message string) (*ZMapProgress, error) {
	// real-time,time-elapsed,time-remaining,percent-complete,hit-rate,active-send-threads,sent-total,sent-last-one-sec,sent-avg-per-sec,recv-success-total,recv-success-last-one-sec,recv-success-avg-per-sec,recv-total,recv-total-last-one-sec,recv-total-avg-per-sec,pcap-drop-total,drop-last-one-sec,drop-avg-per-sec,sendto-fail-total,sendto-fail-last-one-sec,sendto-fail-avg-per-sec
	// 2024-03-14 15:57:46,0,0,0.000000,1,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0
	// 2024-03-14 15:57:47,1,2522890,0.000040,1,1483,1483,1467,19,19,19,39,39,39,0,0,0,0,0,0
	content := strings.Join([]string{
		"real-time,time-elapsed,time-remaining,percent-complete,hit-rate,active-send-threads,sent-total,sent-last-one-sec,sent-avg-per-sec,recv-success-total,recv-success-last-one-sec,recv-success-avg-per-sec,recv-total,recv-total-last-one-sec,recv-total-avg-per-sec,pcap-drop-total,drop-last-one-sec,drop-avg-per-sec,sendto-fail-total,sendto-fail-last-one-sec,sendto-fail-avg-per-sec",
		message,
	}, "\n")
	var progresses []ZMapProgress
	if err := csvutil.Unmarshal([]byte(content), &progresses); err != nil {
		slog.Error("error occured while parsing progress", slog.String("error", err.Error()))
		return &ZMapProgress{}, err
	}
	if len(progresses) == 0 {
		return &ZMapProgress{}, fmt.Errorf("no progress found")
	} else {
		progress := progresses[0]
		progress.Line = message
		return &progress, nil
	}
}

func (z ZMapProgress) GetStatus() task.TaskStatus {
	if z.PercentComplete >= 100 {
		return task.FINISHED
	}
	if z.PercentComplete > 0 {
		return task.RUNNING
	}
	return task.PENDING
}

func (z ZMapProgress) NumTotal() int64 {
	return z.RecvTotal
}

func (z ZMapProgress) NumDoneWithSuccess() int64 {
	return z.RecvSuccessTotal
}

func (z ZMapProgress) NumDoneWithError() int64 {
	return z.RecvTotal - z.RecvSuccessTotal
}

func (z ZMapProgress) String() string {
	return fmt.Sprintf("%s (%f%%)", time.Duration(z.TimeRemaining)*time.Second, z.PercentComplete)
}
