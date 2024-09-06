package models

import (
	"time"
)

type CtrlProcessingJobs struct {
	ProcessingJobName string
	CreatedAt         time.Time
}
