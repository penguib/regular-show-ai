package main

import (
	"fmt"
	"net/http"
	"os"
	"regular-show-ai/endpoints"
	"regular-show-ai/util"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/joho/godotenv"
	openai "github.com/sashabaranov/go-openai"
)

var chatGPT *openai.Client

func main() {

	generateTopics := false
	args := os.Args[1:]
	if len(args) > 0 {
		if args[0] == "-g" {
			generateTopics = true
		}
	}

	fmt.Println(generateTopics)

	r := chi.NewRouter()
	if err := godotenv.Load(); err != nil {
		panic(err)
	}

	chatGPT = openai.NewClient(os.Getenv("OPENAI_API_KEY"))

	r.Use(middleware.Logger)
	r.Use(middleware.RequestID)

	util.ScenesMetadata.Init()

	if generateTopics {
		go GenerateTopics()
	}

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, world"))
	})

	r.Route("/api", func(r chi.Router) {
		r.Route("/scene", func(r chi.Router) {
			r.Get("/", endpoints.GETScene)
		})
	})

	http.ListenAndServe(":3000", r)
}
