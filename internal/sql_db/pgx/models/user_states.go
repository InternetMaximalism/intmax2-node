package models

import "time"

type UserState struct {
	ID                 string
	UserAddress        string
	EncryptedUserState string
	AuthSignature      string
	BlockNumber        int64
	CreatedAt          time.Time
	UpdatedAt          time.Time
}
