package news

import (
	"Timos-API/Newsroom/database"
	ctx "context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-playground/validator"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func collection() *mongo.Collection {
	return database.Database.Collection("news")
}

func printError(w http.ResponseWriter, message string) {
	w.WriteHeader(http.StatusUnprocessableEntity)
	json.NewEncoder(w).Encode(Exception{message})
}

func getNewsById(newsId string) *News {
	var news News
	oid, err := primitive.ObjectIDFromHex(newsId)

	if err != nil {
		fmt.Printf("Invalid ObjectID %v\n", newsId)
		return nil
	}

	err = collection().FindOne(ctx.Background(), bson.M{"_id": oid}).Decode(&news)

	if err != nil {
		fmt.Printf("News not found... (%v) %v \n", newsId, err)
		return nil
	}

	return &news
}

func getAll(w http.ResponseWriter, req *http.Request, filter primitive.M) {
	options := options.Find().SetSort(map[string]int{"timestamp": -1})

	query := req.URL.Query()
	qLimit, qSkip, qQuery := query.Get("limit"), query.Get("skip"), query.Get("query")

	if len(qLimit) > 0 {
		if limit, err := strconv.ParseInt(qLimit, 10, 64); err == nil {
			options.SetLimit(limit)
		}
	}

	if len(qSkip) > 0 {
		if skip, err := strconv.ParseInt(qSkip, 10, 64); err == nil {
			options.SetSkip(skip)
		}
	}

	if len(qQuery) > 0 {
		regex := primitive.Regex{Pattern: qQuery, Options: "i"}
		filter = bson.M{
			"$and": []bson.M{
				filter,
				{"$or": []bson.M{
					{"title": regex}, {"project": regex}, {"type": regex}, {"timestamp": regex}, {"content": regex}, {"thumbnail": regex}, {"featured": regex},
				}},
			},
		}
	}

	cursor, err := collection().Find(ctx.Background(), filter, options)

	if err != nil {
		printError(w, err.Error())
		return
	}

	defer cursor.Close(ctx.Background())

	allNews := []News{}
	for cursor.Next(ctx.Background()) {
		var news News
		cursor.Decode(&news)
		allNews = append(allNews, news)
	}

	if err := cursor.Err(); err != nil {
		printError(w, err.Error())
		return
	}

	json.NewEncoder(w).Encode(allNews)
}

func getAllNews(w http.ResponseWriter, req *http.Request) {
	getAll(w, req, bson.M{})
}

func getFeaturedNews(w http.ResponseWriter, req *http.Request) {
	getAll(w, req, bson.M{"$and": []bson.M{
		{"featured": bson.M{"$exists": "true"}},
		{"featured": bson.M{"$ne": nil}},
	}})
}

func getProjectNews(w http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)
	projectId := params["id"]
	oid, err := primitive.ObjectIDFromHex(projectId)

	if err != nil {
		printError(w, "Invalid ObjectID "+projectId)
		return
	}

	getAll(w, req, bson.M{"project": oid})

}

func getNews(w http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)
	news := getNewsById(params["id"])

	if news == nil {
		printError(w, "News not found")
		return
	}

	json.NewEncoder(w).Encode(news)
}

func deleteNews(w http.ResponseWriter, req *http.Request) {
	newsId, err := primitive.ObjectIDFromHex(mux.Vars(req)["id"])

	if err != nil {
		printError(w, "Invalid NewsID")
		return
	}

	result, _ := collection().DeleteOne(ctx.Background(), bson.M{"_id": newsId})

	if result.DeletedCount == 0 {
		printError(w, "News not found")
		return
	}

	json.NewEncoder(w).Encode(Exception{"News deleted"})
}

func postNews(w http.ResponseWriter, req *http.Request) {
	var news News
	json.NewDecoder(req.Body).Decode(&news)

	validate := validator.New()
	err := validate.Struct(news)

	if err != nil {
		printError(w, err.Error())
		return
	}

	t := time.Now()
	tUnixMilli := int64(time.Nanosecond) * t.UnixNano() / int64(time.Millisecond)

	news.NewsID = nil
	news.Timestamp = &tUnixMilli

	result, _ := collection().InsertOne(ctx.Background(), news)

	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		created := getNewsById(oid.Hex())
		json.NewEncoder(w).Encode(created)
	}
}

func patchNews(w http.ResponseWriter, req *http.Request) {
	newsId, err := primitive.ObjectIDFromHex(mux.Vars(req)["id"])

	if err != nil {
		printError(w, "Invalid NewsID")
		return
	}

	var news News
	json.NewDecoder(req.Body).Decode(&news)

	validate := validator.New()
	err = validate.Struct(news)

	if err != nil {
		printError(w, err.Error())
		return
	}

	news.NewsID = nil
	news.Timestamp = nil
	result, _ := collection().UpdateOne(ctx.Background(), bson.M{"_id": newsId}, bson.M{"$set": news})

	if result.MatchedCount == 0 {
		printError(w, "News not found")
		return
	}

	updated := getNewsById(newsId.Hex())
	json.NewEncoder(w).Encode(updated)
}
