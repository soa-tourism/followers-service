package service

import (
	"followers/model"
	"followers/repo"
)

type RecommendationService struct {
	repo *repo.SocialProfileRepo
}

func NewRecommendationService(r *repo.SocialProfileRepo) *RecommendationService {
	return &RecommendationService{r}
}

func (service *RecommendationService) GetRecommendedAccounts(userId int64) model.SocialProfiles {
	following, _ := service.repo.GetAllFollowing(userId)
	recommended := make(model.SocialProfiles, 0)
	for i := 0; i < len(following); i++ {
		profile := following[i]
		profile_following, _ := service.repo.GetAllFollowing(profile.UserID)
		contains := false
		for j := 0; j < len(profile_following); j++ {
			for k := 0; k < len(recommended); k++ {
				if recommended[k].UserID == profile_following[j].UserID {
					contains = true
				}
			}
			if !contains {
				recommended = append(recommended, profile_following[j])
			}
			contains = false
		}
	}
	return recommended
}
