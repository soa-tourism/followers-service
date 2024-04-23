package main

import (
	//"followers/model"
	"context"
	"followers/handler"
	"followers/repo"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	gorillaHandlers "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

func main() {
	//Reading from environment, if not set we will default it to 8080.
	//This allows flexibility in different environments (for eg. when running multiple docker api's and want to override the default port)
	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = "8082"
	}

	// Initialize context
	timeoutContext, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	//Initialize the logger we are going to use, with prefix and datetime for every log
	logger := log.New(os.Stdout, "[follower-api] ", log.LstdFlags)
	storeLogger := log.New(os.Stdout, "[follower-store] ", log.LstdFlags)

	// NoSQL: Initialize  Repository store
	store, err := repo.New(storeLogger)
	if err != nil {
		logger.Fatal(err)
	}
	defer store.CloseDriverConnection(timeoutContext)
	store.CheckConnection()

	//Initialize the handler and inject said logger
	handler := handler.NewSocialProfileHandler(logger, store)

	//Initialize the router and add a middleware for all the requests
	router := mux.NewRouter()

	router.Use(handler.MiddlewareContentTypeSet)

	//Handle requests
	// postSocialProfileNode := router.Methods(http.MethodPost).Subrouter()
	// postSocialProfileNode.HandleFunc("/profile", handler.CreateSocialProfile)
	// postSocialProfileNode.Use(handler.MiddlewareSocialProfileDeserialization)

	putSocialProfileNode := router.Methods(http.MethodPut).Subrouter()
	putSocialProfileNode.HandleFunc("/profiles/add/{userId}/{username}", handler.CreateSocialProfile)

	getAllSocialProfiles := router.Methods(http.MethodGet).Subrouter()
	getAllSocialProfiles.HandleFunc("/profiles/{limit}", handler.GetAllSocialProfiles)

	getSocialProfileByUserId := router.Methods(http.MethodGet).Subrouter()
	getSocialProfileByUserId.HandleFunc("/profile/{userId}", handler.GetSocialProfileByUserId)

	getFollowers := router.Methods(http.MethodGet).Subrouter()
	getFollowers.HandleFunc("/profiles/followers/{userId}", handler.GetAllFollowers)

	getFollowing := router.Methods(http.MethodGet).Subrouter()
	getFollowing.HandleFunc("/profiles/following/{userId}", handler.GetAllFollowing)

	getRecommended := router.Methods(http.MethodGet).Subrouter()
	getRecommended.HandleFunc("/profiles/recommend/{userId}", handler.GetFollowRecommendations)

	follow := router.Methods(http.MethodPut).Subrouter()
	follow.HandleFunc("/profiles/follow/{userId}/{followerId}", handler.Follow)

	unfollow := router.Methods(http.MethodDelete).Subrouter()
	unfollow.HandleFunc("/profiles/unfollow/{userId}/{followerId}", handler.Unfollow)

	cors := gorillaHandlers.CORS(gorillaHandlers.AllowedOrigins([]string{"*"}))

	//Initialize the server
	server := http.Server{
		Addr:         ":" + port,
		Handler:      cors(router),
		IdleTimeout:  120 * time.Second,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	logger.Println("Server listening on port", port)
	//Distribute all the connections to goroutines
	go func() {
		err := server.ListenAndServe()
		if err != nil {
			logger.Fatal(err)
		}
	}()

	sigCh := make(chan os.Signal)
	signal.Notify(sigCh, os.Interrupt)
	signal.Notify(sigCh, os.Kill)

	sig := <-sigCh
	logger.Println("Received terminate, graceful shutdown", sig)

	//Try to shutdown gracefully
	if server.Shutdown(timeoutContext) != nil {
		logger.Fatal("Cannot gracefully shutdown...")
	}
	logger.Println("Server stopped")
}
