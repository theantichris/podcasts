package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type podcast struct {
	ID     primitive.ObjectID `bson:"_id,omitempty"`
	Title  string             `bson:"title,omitempty"`
	Author string             `bson:"author,omitempty"`
	Tags   []string           `bson:"tags,omitempty"`
}

type episode struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	Podcast     primitive.ObjectID `bson:"podcast,omitempty"`
	Title       string             `bson:"title,omitempty"`
	Description string             `bson:"description,omitempty"`
	Duration    int32              `bson:"duration,omitempty"`
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file.")
	}

	client := initClient(os.Getenv("ATLAS_URI"))
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	err = client.Connect(ctx)
	if err != nil {
		cancel()
		log.Fatalf("could not connect to database server: %v:", err)
	}
	defer client.Disconnect(ctx)

	database := client.Database("podcasts_app")
	// podcastsCollection := database.Collection("podcasts")
	episodesCollection := database.Collection("episodes")

	var episodeResults []episode
	cursor, err := episodesCollection.Find(ctx, bson.M{"duration": bson.D{{"$gt", 25}}})
	if err != nil {
		log.Fatal(err)
	}

	if err = cursor.All(ctx, &episodeResults); err != nil {
		log.Fatal(err)
	}

	fmt.Println(episodeResults)
}

func initClient(uri string) *mongo.Client {
	client, err := mongo.NewClient(options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatal(err)
	}

	return client
}

func insertData(ctx context.Context, podcastsCollection, episodesCollection *mongo.Collection) {
	podcastResult, err := podcastsCollection.InsertOne(ctx, bson.D{
		{Key: "title", Value: "The Polygot Developer Podcst"},
		{Key: "author", Value: "Nic Raboy"},
		{Key: "tags", Value: bson.A{"development", "programming", "coding"}},
	})
	if err != nil {
		log.Fatal(err)
	}

	episodeResult, err := episodesCollection.InsertMany(ctx, []interface{}{
		bson.D{
			{Key: "podcast", Value: podcastResult.InsertedID},
			{Key: "title", Value: "GraphQL for API Development"},
			{Key: "description", Value: "Learn about GraphQL from the go-creator of GraphQL, Lee Byron."},
			{Key: "duration", Value: 25},
		},
		bson.D{
			{Key: "podcast", Value: podcastResult.InsertedID},
			{Key: "title", Value: "Progressive Web Application Development"},
			{Key: "description", Value: "Learn about PWA development with Tara Manicsic."},
			{Key: "duration", Value: 32},
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Inserted %v documents into episode collection!\n", len(episodeResult.InsertedIDs))
}

func readAll(ctx context.Context, collection *mongo.Collection) {
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		log.Fatal(err)
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var result bson.M
		if err = cursor.Decode(&result); err != nil {
			log.Fatal(err)
		}

		fmt.Println(result)
	}
}
