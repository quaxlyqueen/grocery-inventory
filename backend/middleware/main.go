package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"reflect"
	"sync"
	"time"
)

var db sql.DB

// Get the product name and image from openfoodfacts. Pass in the UPC.
func getProduct(upc string) (OpenFoodFactsResponse, error) {
	endpoint := "https://world.openfoodfacts.net/api/v2/product/" + upc + "?product_type=food&fields=product%2Cproduct_name%2Cimage_small_url"
	resp, err := http.Get(endpoint)
	if err != nil {
		log.Println("error accessing openfoodfact.net api")
		return OpenFoodFactsResponse{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return OpenFoodFactsResponse{}, fmt.Errorf("API call returned status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return OpenFoodFactsResponse{}, fmt.Errorf("failed to read response body: %w", err)
	}

	var response OpenFoodFactsResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return OpenFoodFactsResponse{}, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return response, nil
}

// Initializes the connection to the sqlite3 database.
func dbInit() (sql.DB, error) {
	const file string = "../database/inventory.db"
	database, err := sql.Open("sqlite3", file)
	if err != nil {
		log.Println(err)
		return sql.DB{}, err
	}

	return *database, nil
}

// Executes a command on the sqlite3 database
func dbExec(e string) {
	_, err := db.Exec(e)
	if err != nil {
		log.Println("error executing db command")
		log.Println(err)
		return
	}
}

// Queries the sqlite3 database
func dbQuery(q string) ([]interface{}, error) {
	rows, err := db.Query(q)
	if err != nil {
		log.Println("error querying the database!")
		log.Println(err)
		return nil, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		log.Println("error getting column names:", err)
		return nil, err
	}

	values := make([]interface{}, len(columns))
	scanArgs := make([]interface{}, len(columns))
	for i := range values {
		scanArgs[i] = &values[i]
	}

	var results []interface{}
	for rows.Next() {
		err = rows.Scan(scanArgs...)
		if err != nil {
			log.Fatal(err)
			return nil, err
		}

		var rowData = make(map[string]interface{})
		for i, col := range columns {
			val := values[i]
			rowData[col] = val
		}
		results = append(results, rowData)
	}

	return results, nil
}

func jsonToItem(w http.ResponseWriter, r *http.Request) (Upc, error) {
	input, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, "")
		log.Println("error reading json body")
		return Upc{}, err
	}
	defer r.Body.Close()

	item := Upc{}
	err = json.Unmarshal(input, &item)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, "")
		log.Println("error converting JSON to Item")
		return Upc{}, err
	}
	return item, nil
}

func addItem(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	item, err := jsonToItem(w, r)
	if err != nil {
		return
	}

	offResponse, err := getProduct(item.Upc)
	insertIntoItems := "INSERT OR REPLACE INTO items (upc, name, image, count) VALUES ('" + item.Upc + "', '" + offResponse.Product.ProductName + "', '" + offResponse.Product.ImageSmallURL + "', COALESCE((SELECT count FROM items WHERE upc = '" + item.Upc + "'), 0) + 1);"
	dbExec(insertIntoItems)

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "")
}

func deleteItem(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	jsonToItem(w, r)
	/*
	   db, err := sql.Open("sqlite3", file)

	   	if err != nil {
	   	        log.Println(err)
	   		return
	   	}

	   q := "SELECT count FROM items WHERE upc == '" + item.Upc + "' AND count > 1;"
	   rows, rerr := db.Query(q)

	   	if rerr != nil {
	   	        log.Println(rerr)
	   		return
	   	}

	   defer rows.Close()

	   var (

	   	count int64

	   )

	   	for rows.Next() {
	   		if err := rows.Scan(&count); err != nil {
	   			log.Println(err)
	   		}
	   		log.Printf("Count: %d\n", count)
	   	}

	   	if count == 1 {
	   		q := "DELETE FROM items WHERE upc == '" + item.Upc + "'"
	   		qa := "DELETE FROM groceries WHERE ROWID IN (SELECT MIN(ROWID) as row_id FROM groceries WHERE item = '" + item.Upc + "');"
	   		_, err := db.Exec(q)
	   		if err != nil {
	   		        log.Println("error executing" + q)
	   		        log.Println(err)
	   		        return
	   		}

	   		_, qaerr := db.Exec(qa)
	   		if qaerr != nil {
	   		        log.Println("error executing" + qa)
	   		        log.Println(qaerr)
	   		        return
	   		}
	   	} else {

	   		q := "UPDATE items SET count = count -1 WHERE upc = '" + item.Upc + "';"
	   		qa := "DELETE FROM groceries WHERE ROWID IN (SELECT MIN(ROWID) as row_id FROM groceries WHERE item = '" + item.Upc + "');"
	   		_, err := db.Exec(q)
	   		if err != nil {
	   		        log.Println("error executing" + q)
	   		        log.Println(err)
	   		        return
	   		}

	   		_, qaerr := db.Exec(qa)
	   		if qaerr != nil {
	   		        log.Println("error executing" + qa)
	   		        log.Println(qaerr)
	   		        return
	   		}
	   	}

	   w.WriteHeader(http.StatusOK)
	   w.Header().Set("Content-Type", "application/json")
	   fmt.Fprintf(w, "")
	*/
}

// Retrieve all groceries
func listItems(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	items, err := dbQuery("SELECT * FROM items;")
	if err != nil {
		log.Println("error from dbQuery")
		log.Println(err)
		return
	}

	var itemsJson []Item
	for _, itemInterface := range items {
		// Type assertion to convert itemInterface to Grocery
		itemMap, ok := itemInterface.(map[string]interface{})
		if !ok {
			log.Printf("Error: Failed to convert item to Grocery type: %v", reflect.TypeOf(itemInterface))
			continue
		}

		// Extract values from the map
		var i Item
		if upc, ok := itemMap["upc"].(string); ok {
			i.Upc = upc
		}
		if name, ok := itemMap["name"].(string); ok {
			i.Name = name
		}
		if image, ok := itemMap["image"].(string); ok {
			i.Image = image
		}
		if count, ok := itemMap["count"].(int64); ok {
			i.Count = count
		}

		itemsJson = append(itemsJson, Item{
			Upc:   i.Upc,
			Name:  i.Name,
			Image: i.Image,
			Count: i.Count,
		})
	}

	jsonData, err := json.Marshal(itemsJson)
	if err != nil {
		log.Println("Error marshalling to JSON:", err)
		http.Error(w, "Error marshalling items to JSON", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, string(jsonData))
	return
}

func main() {
	var wg sync.WaitGroup
	database, err := dbInit()
	if err != nil {
		log.Println("error initializing the database! aborting")
		return
	}

	db = database

	endpoint := []string{
		"/addItem",
		"/deleteItem",
		"/listItems",
	}

	function := []func(http.ResponseWriter, *http.Request){
		addItem,
		deleteItem,
		listItems,
	}

	r := mux.NewRouter()
	ServeApi(r, "localhost", endpoint, function, "5787")

	srv := &http.Server{
		Addr:         "0.0.0.0:5786",
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      r, // Pass our instance of gorilla/mux in.
	}

	// Run our server in a goroutine so that it doesn't block.
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Println(err)
		}
	}()

	log.Println("Router is running on port 5786")
	c := make(chan os.Signal, 1)
	// We'll accept graceful shutdowns when quit via SIGINT (Ctrl+Shift+C)
	signal.Notify(c, os.Interrupt)
	// Block until we receive our signal.
	<-c
	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(30))
	defer cancel()
	defer db.Close()
	// Doesn't block if no connections, but will otherwise wait
	// until the timeout deadline.
	srv.Shutdown(ctx)
	// Optionally, you could run srv.Shutdown in a goroutine and block on
	// <-ctx.Done() if your application should wait for other services
	// to finalize based on context cancellation.
	log.Println("shutting down")
	wg.Wait()
	os.Exit(0)
}
