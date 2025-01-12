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

func jsonToItem(w http.ResponseWriter, r *http.Request) (Item, error) {
	input, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, "")
		log.Println("error reading json body")
		return Item{}, err
	}

	defer r.Body.Close()
	item := Item{}
	err = json.Unmarshal(input, &item)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, "")
		log.Println("error converting JSON to Item")
		return Item{}, err
	}
	log.Println("success")
	return item, nil
}

func addItem(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	item, err := jsonToItem(w, r)
	if err != nil {
		return
	}

	offResponse, err := getProduct(item.Upc)
	currentTime := time.Now()
	log.Println(offResponse.Product.ProductName)
	insertIntoItems := "INSERT OR REPLACE INTO items (upc, name, image, count) VALUES ('" + item.Upc + "', '" + offResponse.Product.ProductName + "', '" + offResponse.Product.ImageSmallURL + "', COALESCE((SELECT count FROM items WHERE upc = '" + item.Upc + "'), 0) + 1);"
	insertIntoGroceries := "INSERT INTO groceries (item, date_added, exp_date) VALUES ('" + item.Upc + "', '" + currentTime.String() + "', '" + item.ExpDate + "');"
	dbExec(insertIntoItems)
	dbExec(insertIntoGroceries)

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "")
}

func deleteItem(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
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
func listGroceries(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	items, err := dbQuery("SELECT * FROM groceries;")
	if err != nil {
		log.Println("error from dbQuery")
		log.Println(err)
		return
	}

	var groceryJson []GroceryJSON
	for _, itemInterface := range items {
		// Type assertion to convert itemInterface to Grocery
		itemMap, ok := itemInterface.(map[string]interface{})
		if !ok {
			log.Printf("Error: Failed to convert item to Grocery type: %v", reflect.TypeOf(itemInterface))
			continue
		}

		// Extract values from the map
		var grocery Grocery
		if id, ok := itemMap["id"].(float64); ok {
			grocery.ID = int64(id)
		}
		if item, ok := itemMap["item"].(string); ok {
			grocery.Item = item
		}
		if dateAdded, ok := itemMap["date_added"].(string); ok {
			grocery.DateAdded = dateAdded
		}
		if expDate, ok := itemMap["exp_date"].(string); ok {
			grocery.ExpDate = expDate
		}
		if storageID, ok := itemMap["storage_id"].(float64); ok {
			grocery.StorageID = int(storageID)
		}

		groceryJson = append(groceryJson, GroceryJSON{
			ID:        grocery.ID,
			Item:      grocery.Item,
			DateAdded: grocery.DateAdded,
			ExpDate:   grocery.ExpDate,
			StorageID: grocery.StorageID,
		})
	}

	jsonData, err := json.Marshal(groceryJson)
	if err != nil {
		log.Println("Error marshalling to JSON:", err)
		http.Error(w, "Error marshalling groceries to JSON", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(jsonData)
	if err != nil {
		log.Println("Error writing JSON response:", err)
	}
}

func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
	(*w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
	(*w).Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
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
		"/listGroceries",
	}

	function := []func(http.ResponseWriter, *http.Request){
		addItem,
		deleteItem,
		listGroceries,
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
