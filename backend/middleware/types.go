package main

type Item struct {
  Upc       string `json:"upc"`
  Name      string `json:"name"`
  Image     string `json:"image"`
  ExpDate   string `json:"exp_date"`
  Storage   int    `json:"storage_id"`
}

type Count struct {
  Count     int `json:"count"`
}
