package dto

type ProductRequest struct {
	Name        string `json:"name" binding:"required"`
	Price       int    `json:"price" binding:"required"`
	Stock       int    `json:"stock" binding:"required"`
	Img         string `json:"img" binding:"required"`
	Description string `json:"description" binding:"required"`
	CategoryID  int    `json:"categoryId" binding:"required"`
}

type ProductResponse struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Price       int    `json:"price"`
	Stock       int    `json:"stock"`
	Img         string `json:"img"`
	Description string `json:"description"`
	CategoryID  int    `json:"categoryId"`
}
