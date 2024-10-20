package models

import (
	"encoding/json"
	"time"
)

type CtrlProcessingJobs struct {
	ProcessingJobName string
	Options           json.RawMessage
	CreatedAt         time.Time
	UpdatedAt         time.Time
}
