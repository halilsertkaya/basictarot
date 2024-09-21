package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client
var cardsCollection *mongo.Collection
var resultsCollection *mongo.Collection

// Card ve Result yapıları
type Card struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Image string `json:"image"`
}

type Result struct {
	ID   int    `json:"id"`
	C1ID int    `json:"c1id"`
	C2ID int    `json:"c2id"`
	C3ID int    `json:"c3id"`
	Text string `json:"result"`
}

// Kartlar ve sonuçlar
var tarotCards = []Card{
	{1, "Mecnun", "1"},
	{2, "Büyücü", "2"},
	{3, "Azize", "3"},
	{4, "İmparatoriçe", "4"},
	{5, "İmparator", "5"},
	{6, "Aziz", "6"},
	{7, "Aşıklar", "7"},
	{8, "Savaş Arabası", "8"},
	{9, "Güç", "9"},
	{10, "Ermiş", "10"},
	{11, "Kader Çarkı", "11"},
	{12, "Adalet", "12"},
	{13, "Asılan Adam", "13"},
	{14, "Ölüm", "14"},
	{15, "Denge", "15"},
	{16, "Şeytan", "16"},
	{17, "Kale", "17"},
	{18, "Yıldız", "18"},
	{19, "Ay", "19"},
	{20, "Güneş", "20"},
	{21, "Mahkeme", "21"},
	{22, "Dünya", "22"},
}

var tarotResults = []Result{
	{1, 1, 2, 3, "123 - result text"},
	{2, 1, 3, 2, "132 - result text"},
	{3, 2, 1, 3, "213 - result text"},
	{4, 2, 3, 1, "231 - result text"},
	{5, 3, 1, 2, "312 - result text"},
	{6, 3, 2, 1, "321 - result text"},
}

func connectMongo() {
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	var err error
	client, err = mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("MongoDB connection is OK.")
	cardsCollection = client.Database("tarot").Collection("cards")
	resultsCollection = client.Database("tarot").Collection("results")
}

func insertCards() {
	for _, card := range tarotCards {
		filter := map[string]interface{}{"id": card.ID}
		update := map[string]interface{}{"$set": card}

		// UpdateOne ile upsert işlemi yap
		_, err := cardsCollection.UpdateOne(context.TODO(), filter, update, options.Update().SetUpsert(true))
		if err != nil {
			log.Fatal(err)
		}

		// Güncelleme veya ekleme işlemi sonucunu logla
		fmt.Printf("Inserted or updated card with ID: %d, Name: %s\n", card.ID, card.Name)
	}
}

func insertResults() {
	for _, result := range tarotResults {
		filter := map[string]interface{}{"id": result.ID}
		update := map[string]interface{}{"$set": result}

		// UpdateOne ile upsert işlemi yap
		_, err := resultsCollection.UpdateOne(context.TODO(), filter, update, options.Update().SetUpsert(true))
		if err != nil {
			log.Fatal(err)
		}

		// Güncelleme veya ekleme işlemi sonucunu kontrol et
		if err == nil {
			fmt.Printf("Combination [%d %d %d] fetched on ID: %d\n", result.ID, result.C1ID, result.C2ID, result.C3ID)
		}
	}
}

func getCards(w http.ResponseWriter, r *http.Request) {
	// Konsola istek detayını yazdır
	log.Printf("Received request: %s %s", r.Method, r.URL.Path)

	w.Header().Set("Content-Type", "application/json")
	cursor, err := cardsCollection.Find(context.TODO(), map[string]interface{}{})
	if err != nil {
		http.Error(w, `{"status": "error", "message": "`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}
	defer cursor.Close(context.TODO())

	var cards []Card
	if err = cursor.All(context.TODO(), &cards); err != nil {
		http.Error(w, `{"status": "error", "message": "`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(cards)
}

func getResults(w http.ResponseWriter, r *http.Request) {
	// Konsola istek detayını yazdır
	log.Printf("Received request: %s %s", r.Method, r.URL.Path)
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var ids struct {
		C1ID int `json:"c1id"`
		C2ID int `json:"c2id"`
		C3ID int `json:"c3id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&ids); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	filter := map[string]interface{}{
		"c1id": ids.C1ID,
		"c2id": ids.C2ID,
		"c3id": ids.C3ID,
	}

	cursor, err := resultsCollection.Find(context.TODO(), filter)
	if err != nil {
		http.Error(w, `{"status": "error", "message": "`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}
	defer cursor.Close(context.TODO())

	var results []Result
	if err = cursor.All(context.TODO(), &results); err != nil {
		http.Error(w, `{"status": "error", "message": "`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	response := make(map[string]interface{})
	if len(results) > 0 {
		randomIndex := rand.Intn(len(results))
		randomResult := results[randomIndex]
		response["status"] = "success"
		response["result"] = randomResult
	} else {
		response["status"] = "error"
		response["result"] = "No Results found."
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func main() {
	connectMongo()

	insertCards()
	insertResults()

	http.HandleFunc("/cards", getCards)
	http.HandleFunc("/results", getResults)
	log.Fatal(http.ListenAndServe(":9999", nil))
}
