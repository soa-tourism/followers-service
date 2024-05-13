package main

import (
	//"followers/model"
	"context"
	"fmt"
	"followers/model"
	"followers/proto/follower"
	"followers/repo"
	"followers/service"
	"log"
	"net"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

func main() {
	lis, err := net.Listen("tcp", "0.0.0.0:8082")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	timeoutContext, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	storeLogger := log.New(os.Stdout, "[follower-store] ", log.LstdFlags)

	// NoSQL: Initialize  Repository store
	store, err := repo.New(storeLogger)
	if err != nil {
		storeLogger.Fatal(err)
	}
	defer store.CloseDriverConnection(timeoutContext)
	store.CheckConnection()

	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)

	follower.RegisterSocialProfileServer(grpcServer, Server{recommendationService: service.NewRecommendationService(store), repo: store})
	reflection.Register(grpcServer)
	grpcServer.Serve(lis)
}

type Server struct {
	follower.UnimplementedSocialProfileServer
	recommendationService *service.RecommendationService
	repo                  *repo.SocialProfileRepo
}

func (s Server) GetSocialProfile(ctx context.Context, request *follower.UserId) (*follower.SocialProfileResponse, error) {
	p, ok := s.repo.GetSocialProfileByUserId(request.UserId)
	if ok != nil {
		return nil, status.Error(codes.NotFound, "social profile not found")
	}
	response := &follower.SocialProfileResponse{
		UserId:   p.UserID,
		Username: p.Username,
	}
	return response, nil
}
func (s Server) GetFollowers(ctx context.Context, request *follower.UserId) (*follower.SocialProfilesResponse, error) {
	profiles, err := s.repo.GetAllFollowers(request.UserId)
	if err != nil {
		return nil, status.Error(codes.NotFound, "social profile not found")
	}
	var responses []*follower.SocialProfileResponse
	for _, profile := range profiles {
		response := &follower.SocialProfileResponse{
			UserId:   profile.UserID,
			Username: profile.Username,
		}
		responses = append(responses, response)
	}
	socialProfilesResponse := &follower.SocialProfilesResponse{
		SocialProfiles: responses,
	}

	return socialProfilesResponse, nil
}
func (s Server) GetFollowing(ctx context.Context, request *follower.UserId) (*follower.SocialProfilesResponse, error) {
	profiles, err := s.repo.GetAllFollowing(request.UserId)
	if err != nil {
		return nil, status.Error(codes.NotFound, "social profile not found")
	}
	var responses []*follower.SocialProfileResponse
	for _, profile := range profiles {
		response := &follower.SocialProfileResponse{
			UserId:   profile.UserID,
			Username: profile.Username,
		}
		responses = append(responses, response)
	}
	socialProfilesResponse := &follower.SocialProfilesResponse{
		SocialProfiles: responses,
	}

	return socialProfilesResponse, nil
}
func (s Server) GetRecommended(ctx context.Context, request *follower.UserId) (*follower.SocialProfilesResponse, error) {
	profiles := s.recommendationService.GetRecommendedAccounts(request.UserId)
	if profiles == nil {
		return nil, status.Error(codes.NotFound, "social profile not found")
	}
	var responses []*follower.SocialProfileResponse
	for _, profile := range profiles {
		response := &follower.SocialProfileResponse{
			UserId:   profile.UserID,
			Username: profile.Username,
		}
		responses = append(responses, response)
	}
	socialProfilesResponse := &follower.SocialProfilesResponse{
		SocialProfiles: responses,
	}

	return socialProfilesResponse, nil
}
func (s Server) Follow(ctx context.Context, request *follower.FollowRequest) (*follower.SocialProfilesResponse, error) {
	err := s.repo.Follow(request.UserId, request.FollowerId)
	if err != nil {
		return nil, status.Error(codes.NotFound, "social profile not found")
	}
	var responses []*follower.SocialProfileResponse
	socialProfilesResponse := &follower.SocialProfilesResponse{
		SocialProfiles: responses,
	}
	return socialProfilesResponse, nil
}
func (s Server) Unfollow(ctx context.Context, request *follower.UnfollowRequest) (*follower.SocialProfilesResponse, error) {
	err := s.repo.Unfollow(request.UserId, request.FollowedId)
	if err != nil {
		return nil, status.Error(codes.NotFound, "social profile not found")
		fmt.Println("EJ")
	}
	fmt.Println("JO")
	var responses []*follower.SocialProfileResponse
	socialProfilesResponse := &follower.SocialProfilesResponse{
		SocialProfiles: responses,
	}
	return socialProfilesResponse, nil
}
func (s Server) Search(ctx context.Context, request *follower.Username) (*follower.SocialProfilesResponse, error) {
	profiles, err := s.repo.SearchSocialProfilesByUsername(request.Username)
	if err != nil {
		return nil, status.Error(codes.NotFound, "social profile not found")
	}
	var responses []*follower.SocialProfileResponse
	for _, profile := range profiles {
		response := &follower.SocialProfileResponse{
			UserId:   profile.UserID,
			Username: profile.Username,
		}
		responses = append(responses, response)
	}
	socialProfilesResponse := &follower.SocialProfilesResponse{
		SocialProfiles: responses,
	}
	return socialProfilesResponse, nil
}
func (s Server) CreateSocialProfile(ctx context.Context, request *follower.SocialProfileRequest) (*follower.SocialProfileResponse, error) {
	profile := &model.SocialProfile{
		UserID:   request.UserId,
		Username: request.Username,
	}
	err := s.repo.WriteSocialProfile(profile)
	if err != nil {
		return nil, status.Error(codes.NotFound, "social profile not found")
	}
	response := &follower.SocialProfileResponse{
		UserId:   profile.UserID,
		Username: profile.Username,
	}
	return response, nil
}

// func main() {
// 	//Reading from environment, if not set we will default it to 8080.
// 	//This allows flexibility in different environments (for eg. when running multiple docker api's and want to override the default port)
// 	port := os.Getenv("PORT")
// 	if len(port) == 0 {
// 		port = "8082"
// 	}

// 	// Initialize context
// 	timeoutContext, cancel := context.WithTimeout(context.Background(), 30*time.Second)
// 	defer cancel()

// 	//Initialize the logger we are going to use, with prefix and datetime for every log
// 	logger := log.New(os.Stdout, "[follower-api] ", log.LstdFlags)
// 	storeLogger := log.New(os.Stdout, "[follower-store] ", log.LstdFlags)

// 	// NoSQL: Initialize  Repository store
// 	store, err := repo.New(storeLogger)
// 	if err != nil {
// 		logger.Fatal(err)
// 	}
// 	defer store.CloseDriverConnection(timeoutContext)
// 	store.CheckConnection()

// 	//Initialize the handler and inject said logger
// 	handler := handler.NewSocialProfileHandler(logger, store)

// 	//Initialize the router and add a middleware for all the requests
// 	router := mux.NewRouter()

// 	router.Use(handler.MiddlewareContentTypeSet)

// 	//Handle requests
// 	putSocialProfileNode := router.Methods(http.MethodPut).Subrouter()
// 	putSocialProfileNode.HandleFunc("/profiles/add/{userId}/{username}", handler.CreateSocialProfile)

// 	getAllSocialProfiles := router.Methods(http.MethodGet).Subrouter()
// 	getAllSocialProfiles.HandleFunc("/profiles/{limit}", handler.GetAllSocialProfiles)

// 	getSocialProfileByUserId := router.Methods(http.MethodGet).Subrouter()
// 	getSocialProfileByUserId.HandleFunc("/profile/{userId}", handler.GetSocialProfileByUserId)

// 	getFollowers := router.Methods(http.MethodGet).Subrouter()
// 	getFollowers.HandleFunc("/profiles/followers/{userId}", handler.GetAllFollowers)

// 	getFollowing := router.Methods(http.MethodGet).Subrouter()
// 	getFollowing.HandleFunc("/profiles/following/{userId}", handler.GetAllFollowing)

// 	getRecommended := router.Methods(http.MethodGet).Subrouter()
// 	getRecommended.HandleFunc("/profiles/recommend/{userId}", handler.GetFollowRecommendations)

// 	follow := router.Methods(http.MethodPut).Subrouter()
// 	follow.HandleFunc("/profiles/follow/{userId}/{followerId}", handler.Follow)

// 	unfollow := router.Methods(http.MethodDelete).Subrouter()
// 	unfollow.HandleFunc("/profiles/unfollow/{userId}/{followerId}", handler.Unfollow)

// 	search := router.Methods(http.MethodGet).Subrouter()
// 	search.HandleFunc("/profiles/search/{username}", handler.SearchSocialProfilesByUsername)

// 	cors := gorillaHandlers.CORS(gorillaHandlers.AllowedOrigins([]string{"*"}))

// 	//Initialize the server
// 	server := http.Server{
// 		Addr:         ":" + port,
// 		Handler:      cors(router),
// 		IdleTimeout:  120 * time.Second,
// 		ReadTimeout:  5 * time.Second,
// 		WriteTimeout: 5 * time.Second,
// 	}

// 	logger.Println("Server listening on port", port)
// 	//Distribute all the connections to goroutines
// 	go func() {
// 		err := server.ListenAndServe()
// 		if err != nil {
// 			logger.Fatal(err)
// 		}
// 	}()

// 	sigCh := make(chan os.Signal)
// 	signal.Notify(sigCh, os.Interrupt)
// 	signal.Notify(sigCh, os.Kill)

// 	sig := <-sigCh
// 	logger.Println("Received terminate, graceful shutdown", sig)

// 	//Try to shutdown gracefully
// 	if server.Shutdown(timeoutContext) != nil {
// 		logger.Fatal("Cannot gracefully shutdown...")
// 	}
// 	logger.Println("Server stopped")
// }
