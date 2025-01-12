package main

type Item struct {
	Upc     string `json:"upc"`
	Name    string `json:"name"`
	Image   string `json:"image"`
	ExpDate string `json:"exp_date"`
	Storage int    `json:"storage_id"`
}

type OpenFoodFactsResponse struct {
  Code        string 			`json:"code"`
  Product     Product     `json:"product"`
}

type Product struct {
  ImageSmallURL string `json:"image_small_url"`
  ProductName   string `json:"product_name"`
}

type Grocery struct {
	ID        int64
	Item      string
	DateAdded string
	ExpDate   string
	StorageID int
}

type GroceryJSON struct {
	ID        int64  `json:"id"`
	Item      string `json:"item"`
	DateAdded string `json:"date_added"`
	ExpDate   string `json:"exp_date"`
	StorageID int    `json:"storage_id"`
}
