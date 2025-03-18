package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"go.etcd.io/bbolt"
)

type Product struct {
	ID    string  `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

var db *bbolt.DB

func initDB() {
	var err error
	db, err = bbolt.Open("products.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}

	err = db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("Products"))
		return err
	})

	if err != nil {
		log.Fatal(err)
	}
}

func seedProducts() {
	exampleProducts := []Product{
		{"1", "Laptop", 999.99},
		{"2", "Mouse", 19.99},
		{"3", "Keyboard", 49.99},
		{"4", "Monitor", 199.99},
		{"5", "Headphones", 89.99},
	}
	db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("Products"))
		for _, p := range exampleProducts {
			data, _ := json.Marshal(p)
			b.Put([]byte(p.ID), data)
		}
		for i := 6; i <= 10000; i++ {
			id := strconv.Itoa(i)
			p := Product{ID: id, Name: fmt.Sprintf("Product-%d", i), Price: rand.Float64() * 100}
			data, _ := json.Marshal(p)
			b.Put([]byte(id), data)
		}
		return nil
	})
}

func createProduct(w http.ResponseWriter, r *http.Request) {
	var p Product
	json.NewDecoder(r.Body).Decode(&p)
	p.ID = strconv.Itoa(int(time.Now().UnixNano()))
	db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("Products"))
		data, _ := json.Marshal(p)
		return b.Put([]byte(p.ID), data)
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(p)
}

func getAllProducts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var products []Product
	params := r.URL.Query()
	limit, _ := strconv.Atoi(params.Get("limit"))
	offset, _ := strconv.Atoi(params.Get("offset"))
	minPrice, _ := strconv.ParseFloat(params.Get("minPrice"), 64)
	maxPrice, _ := strconv.ParseFloat(params.Get("maxPrice"), 64)

	db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("Products"))
		counter := 0
		b.ForEach(func(k, v []byte) error {
			var p Product
			json.Unmarshal(v, &p)
			if (minPrice == 0 || p.Price >= minPrice) && (maxPrice == 0 || p.Price <= maxPrice) {
				if counter >= offset && (limit == 0 || len(products) < limit) {
					products = append(products, p)
				}
				counter++
			}
			return nil
		})
		return nil
	})
	json.NewEncoder(w).Encode(products)
}

func getProductByID(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	var product Product
	var found bool
	db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("Products"))
		v := b.Get([]byte(params["id"]))
		if v != nil {
			json.Unmarshal(v, &product)
			found = true
		}
		return nil
	})

	if !found {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Product not found"})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(product)
}

func main() {
	initDB()
	seedProducts()
	defer db.Close()

	r := mux.NewRouter()
	r.HandleFunc("/products", getAllProducts).Methods(http.MethodGet)
	r.HandleFunc("/products/{id}", getProductByID).Methods(http.MethodGet)
	r.HandleFunc("/products", createProduct).Methods(http.MethodPost)

	port := 8080
	fmt.Printf("Server running on port %d\n", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), r))
}
