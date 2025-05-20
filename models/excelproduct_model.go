package models

type ProductImport struct {
	Name         string  `excel:"name" binding:"required"`
	Description  string  `excel:"description"`
	SalePrice    float64 `excel:"sale_price" binding:"required"`
	RegularPrice float64 `excel:"regular_price"`
	BaseCost     float64 `excel:"base_cost"`
	Image        string  `excel:"image" binding:"required"`
	Weight       float64 `excel:"weight" binding:"required"`
	CategoryIDs  string  `excel:"category_ids"` // e.g., "1,3,5"
}
