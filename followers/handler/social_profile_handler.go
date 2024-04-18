package handler

import (
	"context"
	"followers/model"
	"followers/repo"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type KeyProduct struct{}

type SocialProfileHandler struct {
	logger *log.Logger
	// NoSQL: injecting social profile repository
	repo *repo.SocialProfileRepo
}

// Injecting the logger makes this code much more testable.
func NewSocialProfileHandler(l *log.Logger, r *repo.SocialProfileRepo) *SocialProfileHandler {
	return &SocialProfileHandler{l, r}
}

// Social Profile Handler Features
func (m *SocialProfileHandler) CreateSocialProfile(rw http.ResponseWriter, h *http.Request) {
	socialProfile := h.Context().Value(KeyProduct{}).(*model.SocialProfile)
	err := m.repo.WriteSocialProfile(socialProfile)
	if err != nil {
		m.logger.Print("Database exception: ", err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
	rw.WriteHeader(http.StatusCreated)
}

func (m *SocialProfileHandler) GetAllSocialProfiles(rw http.ResponseWriter, h *http.Request) {
	vars := mux.Vars(h)
	limit, err := strconv.Atoi(vars["limit"])
	if err != nil {
		m.logger.Printf("Expected integer, got: %d", limit)
		http.Error(rw, "Unable to convert limit to integer", http.StatusBadRequest)
		return
	}

	profiles, err := m.repo.GetAllSocialProfiles(limit)
	if err != nil {
		m.logger.Print("Database exception: ", err)
	}

	if profiles == nil {
		return
	}

	err = profiles.ToJSON(rw)
	if err != nil {
		http.Error(rw, "Unable to convert to json", http.StatusInternalServerError)
		m.logger.Fatal("Unable to convert to json :", err)
		return
	}
}

func (m *SocialProfileHandler) GetSocialProfileByUserId(rw http.ResponseWriter, h *http.Request) {
	vars := mux.Vars(h)
	userId, err := strconv.ParseInt(vars["userId"], 10, 64)
	if err != nil {
		http.Error(rw, "Invalid userId", http.StatusBadRequest)
		m.logger.Println("Invalid userId:", err)
		return
	}

	profile, err := m.repo.GetSocialProfileByUserId(userId)
	if err != nil {
		m.logger.Print("Database exception: ", err)
		http.Error(rw, "Error retrieving social profile", http.StatusInternalServerError)
		return
	}

	if profile == nil {
		http.NotFound(rw, h)
		return
	}

	// Convert the profile to a slice containing a single profile
	socialProfile := model.SocialProfiles{profile}
	err = socialProfile.ToJSON(rw)
	if err != nil {
		http.Error(rw, "Unable to convert to json", http.StatusInternalServerError)
		m.logger.Fatal("Unable to convert to json :", err)
		return
	}
}

func (m *SocialProfileHandler) GetAllFollowers(rw http.ResponseWriter, h *http.Request) {
	vars := mux.Vars(h)
	userId, err := strconv.ParseInt(vars["userId"], 10, 64)
	if err != nil {
		http.Error(rw, "Invalid userId", http.StatusBadRequest)
		m.logger.Println("Invalid userId:", err)
		return
	}

	profiles, err := m.repo.GetAllFollowers(userId)
	if err != nil {
		m.logger.Print("Database exception: ", err)
	}

	if profiles == nil {
		return
	}

	err = profiles.ToJSON(rw)
	if err != nil {
		http.Error(rw, "Unable to convert to json", http.StatusInternalServerError)
		m.logger.Fatal("Unable to convert to json :", err)
		return
	}
}

func (m *SocialProfileHandler) GetAllFollowing(rw http.ResponseWriter, h *http.Request) {
	vars := mux.Vars(h)
	userId, err := strconv.ParseInt(vars["userId"], 10, 64)
	if err != nil {
		http.Error(rw, "Invalid userId", http.StatusBadRequest)
		m.logger.Println("Invalid userId:", err)
		return
	}

	profiles, err := m.repo.GetAllFollowing(userId)
	if err != nil {
		m.logger.Print("Database exception: ", err)
	}

	if profiles == nil {
		return
	}

	err = profiles.ToJSON(rw)
	if err != nil {
		http.Error(rw, "Unable to convert to json", http.StatusInternalServerError)
		m.logger.Fatal("Unable to convert to json :", err)
		return
	}
}

func (m *SocialProfileHandler) Follow(rw http.ResponseWriter, h *http.Request) {
	vars := mux.Vars(h)
	userId, err := strconv.ParseInt(vars["userId"], 10, 64)
	if err != nil {
		http.Error(rw, "Invalid userId", http.StatusBadRequest)
		m.logger.Println("Invalid userId:", err)
		return
	}

	followerId, err := strconv.ParseInt(vars["followerId"], 10, 64)
	if err != nil {
		http.Error(rw, "Invalid followerId", http.StatusBadRequest)
		m.logger.Println("Invalid followerId:", err)
		return
	}

	err = m.repo.Follow(userId, followerId)
	if err != nil {
		m.logger.Println("Error following user:", err)
		http.Error(rw, "Error following user", http.StatusInternalServerError)
		return
	}

	rw.WriteHeader(http.StatusOK)
}

func (m *SocialProfileHandler) Unfollow(rw http.ResponseWriter, h *http.Request) {
	vars := mux.Vars(h)
	userId, err := strconv.ParseInt(vars["userId"], 10, 64)
	if err != nil {
		http.Error(rw, "Invalid userId", http.StatusBadRequest)
		m.logger.Println("Invalid userId:", err)
		return
	}

	followerId, err := strconv.ParseInt(vars["followerId"], 10, 64)
	if err != nil {
		http.Error(rw, "Invalid followerId", http.StatusBadRequest)
		m.logger.Println("Invalid followerId:", err)
		return
	}

	err = m.repo.Unfollow(userId, followerId)
	if err != nil {
		m.logger.Println("Error unfollowing user:", err)
		http.Error(rw, "Error unfollowing user", http.StatusInternalServerError)
		return
	}

	rw.WriteHeader(http.StatusOK)
}

func (m *SocialProfileHandler) MiddlewareSocialProfileDeserialization(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, h *http.Request) {
		profile := &model.SocialProfile{}
		err := profile.FromJSON(h.Body)
		if err != nil {
			http.Error(rw, "Unable to decode json", http.StatusBadRequest)
			m.logger.Fatal(err)
			return
		}
		ctx := context.WithValue(h.Context(), KeyProduct{}, profile)
		h = h.WithContext(ctx)
		next.ServeHTTP(rw, h)
	})
}

func (m *SocialProfileHandler) MiddlewareContentTypeSet(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, h *http.Request) {
		m.logger.Println("Method [", h.Method, "] - Hit path :", h.URL.Path)
		rw.Header().Add("Content-Type", "application/json")
		next.ServeHTTP(rw, h)
	})
}
