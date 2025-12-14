package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/Johnnycyan/elevenlabs/client"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var dbClient *mongo.Client

var (
	dbName    string
	mongoUser string
	mongoPass string
	mongoHost string
	mongoPort string
)

const (
	collectionName = "data"
)

type Data struct {
	Date          time.Time `json:"date" bson:"date"`
	Channel       string    `json:"channel" bson:"channel"`
	NumCharacters int       `json:"num_characters" bson:"num_characters"`
	EstimatedCost float64   `json:"estimated_cost" bson:"estimated_cost"`
}

func setupDB() {
	var err error
	var mongoURI string
	if mongoUser == "" || mongoPass == "" {
		mongoURI = fmt.Sprintf("mongodb://%s:%s", mongoHost, mongoPort)
	} else {
		mongoURI = fmt.Sprintf("mongodb://%s:%s@%s:%s", mongoUser, mongoPass, mongoHost, mongoPort)
	}

	// Establish MongoDB connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	dbClient, err = mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
}

func createData(request Request) (*Data, error) {
	numCharacters := len(request.Text)
	elevenPriceStr := os.Getenv("ELEVENLABS_PRICE")

	elevenPrice, err := strconv.ParseFloat(elevenPriceStr, 64)
	if err != nil {
		log.Fatalf("Failed to parse ELEVENLABS_PRICE: %v", err)
		return nil, err
	}

	ctx := context.Background()
	client := client.New(elevenKey)

	clientData, err := client.GetUserInfo(ctx)
	if err != nil {
		logger("Error getting user info: "+err.Error(), logError, request.Channel)
		return nil, err
	}

	elevenChars := float64(clientData.Subscription.CharacterLimit)

	estimatedCost := float64(numCharacters) * elevenPrice / elevenChars

	data := Data{
		Date:          time.Now(),
		Channel:       request.Channel,
		NumCharacters: numCharacters,
		EstimatedCost: estimatedCost,
	}

	return &data, nil
}

func addData(data *Data) error {
	collection := dbClient.Database(dbName).Collection(collectionName)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if _, err := collection.InsertOne(ctx, data); err != nil {
		return err
	}

	return nil
}

func viewDataHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	channel := vars["channel"]

	startDate := r.URL.Query().Get("start")
	endDate := r.URL.Query().Get("end")

	filter := bson.M{"channel": channel}

	if startDate != "" && endDate != "" {
		startTime, err := time.Parse("2006-01-02", startDate)
		if err != nil {
			http.Error(w, "Invalid start date format", http.StatusBadRequest)
			return
		}
		endTime, err := time.Parse("2006-01-02", endDate)
		if err != nil {
			http.Error(w, "Invalid end date format", http.StatusBadRequest)
			return
		}
		endTime = endTime.Add(24 * time.Hour) // Add 24 hours to include the entire day
		filter["date"] = bson.M{
			"$gte": startTime,
			"$lt":  endTime,
		}
	}

	collection := dbClient.Database(dbName).Collection(collectionName)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cur, err := collection.Find(ctx, filter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer cur.Close(ctx)

	var results []Data
	for cur.Next(ctx) {
		var data Data
		if err := cur.Decode(&data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		results = append(results, data)
	}
	if err := cur.Err(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(results); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
