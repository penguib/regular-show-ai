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

const (
	port = 3000
)

func main() {
	fmt.Print("\033[H\033[2J")
	fmt.Print(util.Mordecai)

	generateTopics := false
	args := os.Args[1:]
    dpath := ""

    // dubious. should be using an argument parser package
	if len(args) > 0 {
        dpath = args[0]
        if len(args) > 1 {
            if args[1] == "-g" {
                generateTopics = true
            }
        }
	} else {
        util.Log("Not enough arguments")
        return
    }
    util.Log(fmt.Sprintf("Set disk path to %s", dpath))

	r := chi.NewRouter()

    if generateTopics {
		util.Log("Generating topics enabled")
        if err := godotenv.Load(); err != nil {
            panic(err)
        }
        chatGPT = openai.NewClient(os.Getenv("OPENAI_API_KEY"))
    }


	r.Use(middleware.Logger)
	r.Use(middleware.RequestID)

    err := util.ScenesMetadata.Init(dpath)
    if err != nil {
        util.Log("You have no scenes on disk")
        return
    }


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

	util.Log(fmt.Sprintf("Listening at port %d", port))
	http.ListenAndServe(fmt.Sprintf(":%d", port), r)
}
