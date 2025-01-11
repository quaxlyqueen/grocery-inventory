package main
import (
	"context"
	"encoding/json"
	"io"
	"log"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"
	"database/sql"
	"strconv"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
)
const file string = "/home/violet/documents/development/grocery-inventory/backend/database/inventory.db"
func addItem(w http.ResponseWriter, r *http.Request) {
        enableCors(&w)
	input, err := io.ReadAll(r.Body)
	if err != nil {
		return
	}
	defer r.Body.Close()
	item := Item{}
	json.Unmarshal(input, &item)

        db, err := sql.Open("sqlite3", file)
        if err != nil {
                log.Println(err)
	        return
        }
        currentTime := time.Now()
        insertIntoItems := "INSERT OR REPLACE INTO items (upc, name, image, count) VALUES ('" + item.Upc + "', '" + item.Name + "', '" + item.Image + "', COALESCE((SELECT count FROM items WHERE upc = '" + item.Upc + "'), 0) + 1);"
        insertIntoGroceries := "INSERT INTO groceries (item, date_added, exp_date, storage_id) VALUES ('" + item.Upc + "', '" + currentTime.String() + "', '" + item.ExpDate + "', '" + strconv.Itoa(item.Storage) + "');"
        db.Exec(insertIntoItems)
        db.Exec(insertIntoGroceries)

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "")
}

func deleteItem(w http.ResponseWriter, r *http.Request) {
        enableCors(&w)
	input, err := io.ReadAll(r.Body)
	if err != nil {
		return
	}
	defer r.Body.Close()
	item := Item{}
	json.Unmarshal(input, &item)

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
}

func listGroceries(w http.ResponseWriter, r *http.Request) {
        enableCors(&w)
        db, err := sql.Open("sqlite3", file)
        if err != nil {
                log.Println(err)
	        return
        }

        rows, rerr := db.Query("SELECT * FROM groceries;")
        if rerr != nil {
                log.Println(rerr)
	        return
        }
        defer rows.Close()
        for rows.Next() {
	        var (
		        id   int64
		        item string
		        date_added string
		        exp_date string
		        storage_id int
	        )
	        if err := rows.Scan(&id, &item, &date_added, &exp_date, &storage_id); err != nil {
		        log.Fatal(err)
	        }
	        log.Printf("%d: %s\n", id, item)
	}
}

func enableCors(w *http.ResponseWriter) {
        (*w).Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
        (*w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS") 
        (*w).Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
}

func main() {
        var wg sync.WaitGroup

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
	        Addr: "0.0.0.0:5786",
	        // Good practice to set timeouts to avoid Slowloris attacks.
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
