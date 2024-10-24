package models

import (
	"time"

	"github.com/holiman/uint256"
)

type RelationshipL2BatchIndexBlockContents struct {
	L2BatchIndex    uint256.Int
	BlockContentsID string
	CreatedAt       time.Time
}
