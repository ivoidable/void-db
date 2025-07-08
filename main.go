package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"
)

type Item struct {
	Value      interface{}
	Expiration int64
	Priority   int
}

type Transaction struct {
	Action string
	Key    string
	Item   Item
}

var (
	db            = make(map[string]Item)
	mu            sync.RWMutex
	once          sync.Once
	transactionLog *os.File
)

func main() {
	var err error
	transactionLog, err = os.OpenFile("transaction.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer transactionLog.Close()

	loadFromLog()
	fmt.Println("Loading from previous state...\n")

	once.Do(func() {
		go janitor()
	})

	fmt.Println("Server starting on port 6969")
	http.HandleFunc("/get", getHandler)
	http.HandleFunc("/set", setHandler)
	http.HandleFunc("/delete", deleteHandler)

	http.ListenAndServe(":6969", nil)
}

func loadFromLog() {
	file, err := os.Open("transaction.log")
	if err != nil {
		if os.IsNotExist(err) {
			return
		}
		panic(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var tx Transaction
		if err := json.Unmarshal(scanner.Bytes(), &tx); err != nil {
			fmt.Println("error unmarshalling transaction:", err)
			continue
		}

		switch tx.Action {
		case "set":
			db[tx.Key] = tx.Item
		case "delete":
			delete(db, tx.Key)
		}
	}
}

func writeTransaction(tx Transaction) {
	mu.Lock()
	defer mu.Unlock()

	data, err := json.Marshal(tx)
	if err != nil {
		fmt.Println("error marshalling transaction:", err)
		return
	}

	if _, err := transactionLog.Write(append(data, '\n')); err != nil {
		fmt.Println("error writing transaction to log:", err)
	}
}

func getHandler(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	if key == "" {
		http.Error(w, "missing key", http.StatusBadRequest)
		return
	}

	mu.RLock()
	item, found := db[key]
	mu.RUnlock()

	if !found {
		http.Error(w, "key not found", http.StatusNotFound)
		return
	}

	fmt.Println(key)
	fmt.Println("Requested Key Sent")
	json.NewEncoder(w).Encode(item.Value)
}

func setHandler(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Key      string      `json:"key"`
		Value    interface{} `json:"value"`
		TTL      int64       `json:"ttl"`
		Priority int         `json:"priority"`
	}

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var expiration int64
	if data.TTL > 0 {
		expiration = time.Now().Add(time.Duration(data.TTL) * time.Second).UnixNano()
	}

	item := Item{
		Value:      data.Value,
		Expiration: expiration,
		Priority:   data.Priority,
	}

	mu.Lock()
	db[data.Key] = item
	mu.Unlock()

	w.WriteHeader(http.StatusOK)
	writeTransaction(Transaction{Action: "set", Key: data.Key, Item: item})
	
}

func deleteHandler(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	if key == "" {
		http.Error(w, "missing key", http.StatusBadRequest)
		return
	}

	mu.Lock()
	delete(db, key)
	mu.Unlock()

	writeTransaction(Transaction{Action: "delete", Key: key})

	w.WriteHeader(http.StatusNoContent)
}

func janitor() {
	for range time.Tick(1 * time.Second) {
		mu.Lock()
		for key, item := range db {
			if item.Expiration > 0 && time.Now().UnixNano() > item.Expiration {
				delete(db, key)
				writeTransaction(Transaction{Action: "delete", Key: key})
			}
		}
		mu.Unlock()
	}
}
