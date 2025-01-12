package main

type Upc struct {
	Upc string `json:"upc"`
}

type Item struct {
	Upc     string `json:"upc"`
	Name    string `json:"name"`
	Image   string `json:"image"`
	Count int64 `json:"count"`
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
