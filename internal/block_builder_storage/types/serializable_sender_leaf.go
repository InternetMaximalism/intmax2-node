package types

type SerializableSenderLeaf struct {
	Sender  string `json:"sender"`
	IsValid bool   `json:"isValid"`
}
