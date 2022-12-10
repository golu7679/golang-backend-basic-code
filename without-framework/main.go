package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/thedevsaddam/renderer"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var rnd *renderer.Render
var collection *mongo.Collection

const (
	dbUrl          string = "mongodb://localhost:27017"
	dbName         string = "demo_todo"
	collectionName string = "todo"
	port           string = ":9000"
)

type (
	todo struct {
		ID        string    `json:"_id"`
		Title     string    `json:"title"`
		Completed bool      `json:"completed"`
		CreatedAt time.Time `json:"created_at"`
	}
)

func init() {

	clientOptions := options.Client().ApplyURI(dbUrl)

	client, err := mongo.Connect(context.TODO(), clientOptions)
	checkErr(err)

	err = client.Ping(context.TODO(), nil)
	checkErr(err)
	fmt.Println("connected to mongodb")
	collection = client.Database(dbName).Collection(collectionName)
	fmt.Println("collection instance created")

	rnd = renderer.New()
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	err := rnd.Template(w, http.StatusOK, []string{"static/home.tpl"}, nil)
	checkErr(err)

}

func fetchTodos(w http.ResponseWriter, r *http.Request) {

	cur, err := collection.Find(context.Background(), bson.D{{}})
	if err != nil {
		log.Fatal(err)
	}

	var results []primitive.M
	for cur.Next(context.Background()) {
		var result bson.M
		e := cur.Decode(&result)
		checkErr(e)

		results = append(results, result)
	}

	if err := cur.Err(); err != nil {
		log.Fatal(err)
	}
	cur.Close(context.Background())

	rnd.JSON(w, http.StatusOK, renderer.M{
		"data": results,
	})
}

func createTodo(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Context-Type", "application/x-www-form-urlencoded")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	var t todo

	json.NewDecoder(r.Body).Decode(&t)

	insertResult, err := collection.InsertOne(context.Background(), t)
	log.Println(insertResult.InsertedID)

	if err != nil {
		log.Fatal(err)
	}

	rnd.JSON(w, http.StatusCreated, renderer.M{
		"message": "todo created successfully",
		"todo_id": insertResult.InsertedID,
	})

}

func deleteTodo(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/x-www-form-urlencoded")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	id := strings.TrimSpace(chi.URLParam(r, "id"))

	hexId, _ := primitive.ObjectIDFromHex(id)
	filter := bson.M{"_id": hexId}
	_, err := collection.DeleteOne(context.Background(), filter)
	if err != nil {
		rnd.JSON(w, http.StatusProcessing, renderer.M{
			"message": "Unable to delete task",
		})
		log.Fatal(err)
	}

	rnd.JSON(w, http.StatusOK, renderer.M{
		"message": "todo deleted successfully",
	})

}

func updateTodo(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(chi.URLParam(r, "id"))
	if !primitive.IsValidObjectID(id) {
		rnd.JSON(w, http.StatusBadRequest, renderer.M{
			"message": "The id is invalid",
		})
		return
	}

	var t todo

	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		rnd.JSON(w, http.StatusProcessing, err)
		return
	}

	if t.Title == "" {
		rnd.JSON(w, http.StatusBadRequest, renderer.M{
			"message": "The title filed is required",
		})
		return
	}

	hexId, _ := primitive.ObjectIDFromHex(id)
	filter := bson.M{"_id": hexId}
	update := bson.M{"$set": bson.M{"title": t.Title, "completed": t.Completed}}
	_, err := collection.UpdateOne(context.Background(), filter, update)

	if err != nil {
		rnd.JSON(w, http.StatusProcessing, renderer.M{
			"message": "failed to update todo ",
			"error":   err,
		})
		return
	}

	rnd.JSON(w, http.StatusOK, renderer.M{
		"message": "Todo updated successfully",
	})

}

func main() {
	// for server stop code start
	stopChan := make(chan os.Signal)
	signal.Notify(stopChan, os.Interrupt)
	// end

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/", homeHandler)
	r.Mount("/todo", todoHandlers())

	srv := &http.Server{
		Addr:         port,
		Handler:      r,
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Println("Listening on port", port)
		if err := srv.ListenAndServe(); err != nil {
			log.Printf("listen:%s\n", err)
		}
	}()

	// to stop server code start
	<-stopChan
	log.Println("shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	srv.Shutdown(ctx)
	defer cancel()
	log.Println("server gracefully stopped")
	//end
}

func todoHandlers() http.Handler {
	rg := chi.NewRouter()
	rg.Group(func(r chi.Router) {
		r.Get("/", fetchTodos)
		r.Post("/", createTodo)
		r.Put("/{id}", updateTodo)
		r.Delete("/{id}", deleteTodo)
	})
	return rg
}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
