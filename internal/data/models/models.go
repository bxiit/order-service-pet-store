package models

type Order struct {
	ID     int32 `json:"id,omitempty"`
	UserId int32 `json:"user_id"`
	ItemId int32 `json:"item_id"`
}
