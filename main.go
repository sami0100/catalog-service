package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Book struct {
	ID     primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Title  string  `json:"title" bson:"title"`
	Author string  `json:"author" bson:"author"`
	Price  float64 `json:"price" bson:"price"`
	Stock  int     `json:"stock" bson:"stock"`
}

var client *mongo.Client
var coll *mongo.Collection

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v)
}

func health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status":"ok","service":"catalog-service"})
}

func listBooks(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cur, err := coll.Find(ctx, bson.M{})
	if err != nil { http.Error(w, err.Error(), 500); return }
	var out []Book
	if err := cur.All(ctx, &out); err != nil { http.Error(w, err.Error(), 500); return }
	writeJSON(w, http.StatusOK, out)
}

func createBook(w http.ResponseWriter, r *http.Request) {
	var b Book
	if err := json.NewDecoder(r.Body).Decode(&b); err != nil { http.Error(w, err.Error(), 400); return }
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	res, err := coll.InsertOne(ctx, b)
	if err != nil { http.Error(w, err.Error(), 500); return }
	id, _ := res.InsertedID.(primitive.ObjectID)
	writeJSON(w, http.StatusCreated, map[string]any{"message":"created", "id": id.Hex()})
}

func updateBook(w http.ResponseWriter, r *http.Request) {
	idHex := r.URL.Path[len("/books/"):]
	id, err := primitive.ObjectIDFromHex(idHex)
	if err != nil { http.Error(w, "invalid id", 400); return }
	var b Book
	if err := json.NewDecoder(r.Body).Decode(&b); err != nil { http.Error(w, err.Error(), 400); return }
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	upd := bson.M{"$set": bson.M{"title": b.Title, "author": b.Author, "price": b.Price, "stock": b.Stock}}
	res, err := coll.UpdateByID(ctx, id, upd)
	if err != nil { http.Error(w, err.Error(), 500); return }
	if res.MatchedCount == 0 { http.Error(w, "not found", 404); return }
	writeJSON(w, http.StatusOK, map[string]string{"message":"updated"})
}

func deleteBook(w http.ResponseWriter, r *http.Request) {
	idHex := r.URL.Path[len("/books/"):]
	id, err := primitive.ObjectIDFromHex(idHex)
	if err != nil { http.Error(w, "invalid id", 400); return }
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	res, err := coll.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil { http.Error(w, err.Error(), 500); return }
	if res.DeletedCount == 0 { http.Error(w, "not found", 404); return }
	writeJSON(w, http.StatusOK, map[string]string{"message":"deleted"})
}

func main() {
	port := os.Getenv("PORT")
	if port == "" { port = "3002" }

	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" { mongoURI = "mongodb://localhost:27017" }
	dbName := os.Getenv("MONGO_DB")
	if dbName == "" { dbName = "catalogdb" }

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	c, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil { log.Fatalf("mongo connect: %v", err) }
	client = c
	coll = client.Database(dbName).Collection("books")

	http.HandleFunc("/health", health)
	http.HandleFunc("/books", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet { listBooks(w, r); return }
		if r.Method == http.MethodPost { createBook(w, r); return }
		http.Error(w, "method not allowed", 405)
	})
	http.HandleFunc("/books/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPut:
			updateBook(w, r); return
		case http.MethodDelete:
			deleteBook(w, r); return
		default:
			http.Error(w, "method not allowed", 405)
		}
	})

	log.Printf("catalog-service listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
