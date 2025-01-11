package main

type Item struct {
	Upc     string `json:"upc"`
	Name    string `json:"name"`
	Image   string `json:"image"`
	ExpDate string `json:"exp_date"`
	Storage int    `json:"storage_id"`
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

type Count struct {
	Count int `json:"count"`
}
