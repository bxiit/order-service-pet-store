package dto

type OrderDTO struct {
	ID     int32   `json:"id,omitempty"`
	UserId int32   `json:"user_id"`
	ItemId int32   `json:"item_id"`
	Item   ItemDTO `json:"item"`
}

type ItemDTO struct {
	ID          int32  `json:"id,omitempty"`
	Name        string `json:"name,omitempty"`
	Price       int32  `json:"price,omitempty"`
	Description string `json:"description,omitempty"`
	Quantity    int32  `json:"quantity,omitempty"`
	ImageURL    string `json:"image_url"`
}
