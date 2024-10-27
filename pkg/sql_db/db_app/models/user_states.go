package models

import "time"

type UserState struct {
	ID                 string
	UserAddress        string
	EncryptedUserState []byte
	AuthSignature      string
	CreatedAt          time.Time
	ModifiedAt         time.Time
}
