package repo

import (
	"context"
	"followers/model"
	"log"
	"os"

	// NoSQL: module containing Neo4J api client
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

// NoSQL: SocialProfileRepo struct encapsulating Neo4J api client
type SocialProfileRepo struct {
	// Thread-safe instance which maintains a database connection pool
	driver neo4j.DriverWithContext
	logger *log.Logger
}

// NoSQL: Constructor which reads db configuration from environment and creates a keyspace
func New(logger *log.Logger) (*SocialProfileRepo, error) {
	// Local instance
	uri := os.Getenv("NEO4J_DB")
	user := os.Getenv("NEO4J_USERNAME")
	pass := os.Getenv("NEO4J_PASS")
	auth := neo4j.BasicAuth(user, pass, "")

	driver, err := neo4j.NewDriverWithContext(uri, auth)
	if err != nil {
		logger.Panic(err)
		return nil, err
	}

	// Return repository with logger and DB session
	return &SocialProfileRepo{
		driver: driver,
		logger: logger,
	}, nil
}

// Check if connection is established
func (spr *SocialProfileRepo) CheckConnection() {
	ctx := context.Background()
	err := spr.driver.VerifyConnectivity(ctx)
	if err != nil {
		spr.logger.Panic(err)
		return
	}
	// Print Neo4J server address
	spr.logger.Printf(`Neo4J server address: %s`, spr.driver.Target().Host)
}

// Disconnect from database
func (spr *SocialProfileRepo) CloseDriverConnection(ctx context.Context) {
	spr.driver.Close(ctx)
}

// Social Profile Repo Features
func (mr *SocialProfileRepo) WriteSocialProfile(profile *model.SocialProfile) error {
	// Neo4J Sessions are lightweight so we create one for each transaction (Cassandra sessions are not lightweight!)
	// Sessions are NOT thread safe
	ctx := context.Background()
	session := mr.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: "neo4j"})
	defer session.Close(ctx)

	// ExecuteWrite for write transactions (Create/Update/Delete)
	savedSocialProfile, err := session.ExecuteWrite(ctx,
		func(transaction neo4j.ManagedTransaction) (any, error) {
			result, err := transaction.Run(ctx,
				"CREATE (p:SocialProfile) SET p.userId = $userID, p.username = $username RETURN p.username + ', from node ' + id(p)",
				map[string]any{"userID": profile.UserID, "username": profile.Username})
			if err != nil {
				return nil, err
			}

			if result.Next(ctx) {
				return result.Record().Values[0], nil
			}

			return nil, result.Err()
		})
	if err != nil {
		mr.logger.Println("Error inserting Social Profile:", err)
		return err
	}
	mr.logger.Println(savedSocialProfile.(string))
	return nil
}

func (mr *SocialProfileRepo) GetAllSocialProfiles(limit int) (model.SocialProfiles, error) {
	ctx := context.Background()
	session := mr.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: "neo4j"})
	defer session.Close(ctx)

	// ExecuteRead for read transactions (Read and queries)
	profileResults, err := session.ExecuteRead(ctx,
		func(transaction neo4j.ManagedTransaction) (any, error) {
			result, err := transaction.Run(ctx,
				`MATCH (profile:SocialProfile)
				RETURN profile.userId as userId, profile.username as username
				LIMIT $limit`,
				map[string]any{"limit": limit})
			if err != nil {
				return nil, err
			}

			// Option 1: we iterate over result while there are records
			var profiles model.SocialProfiles
			for result.Next(ctx) {
				record := result.Record()
				userId, ok := record.Get("userId")
				if !ok || userId == nil {
					userId = 0
				}
				username, _ := record.Get("username")
				profiles = append(profiles, &model.SocialProfile{
					UserID:   userId.(int64),
					Username: username.(string),
				})
			}
			return profiles, nil
			// Option 2: we collect all records from result and iterate and map outside of the transaction
			// return result.Collect(ctx)
		})
	if err != nil {
		mr.logger.Println("Error querying search:", err)
		return nil, err
	}
	return profileResults.(model.SocialProfiles), nil
}

func (mr *SocialProfileRepo) GetSocialProfileByUserId(userId int64) (*model.SocialProfileData, error) {
	ctx := context.Background()
	session := mr.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: "neo4j"})
	defer session.Close(ctx)

	// ExecuteRead for read transactions (Read and queries)
	result, err := session.ExecuteRead(ctx,
		func(transaction neo4j.ManagedTransaction) (any, error) {
			result, err := transaction.Run(ctx,
				`MATCH (profile:SocialProfile {userId: $userId})
                OPTIONAL MATCH (profile)<-[:FOLLOWS]-(follower:SocialProfile)
                OPTIONAL MATCH (profile)-[:FOLLOWS]->(following:SocialProfile)
                RETURN profile.userId as userId, profile.username as username,
                collect(DISTINCT { userId: follower.userId, username: follower.username }) as followers,
                collect(DISTINCT { userId: following.userId, username: following.username }) as following`,
				map[string]any{"userId": userId})
			if err != nil {
				return nil, err
			}

			if result.Next(ctx) {
				record := result.Record()
				userID, ok := record.Get("userId")
				if !ok || userID == nil {
					return nil, nil
				}
				username, _ := record.Get("username")

				// Parse followers list
				followersValue, ok := record.Get("followers")
				var followers []*model.SocialProfile
				if ok && followersValue != nil {
					followersMapList := followersValue.([]interface{})
					for _, followerMap := range followersMapList {
						if followerMap != nil {
							userID, userIDOK := followerMap.(map[string]interface{})["userId"].(int64)
							username, usernameOK := followerMap.(map[string]interface{})["username"].(string)
							if userIDOK && usernameOK {
								follower := &model.SocialProfile{
									UserID:   userID,
									Username: username,
								}
								followers = append(followers, follower)
							}
						}
					}
				}

				// Parse following list
				followingValue, ok := record.Get("following")
				var following []*model.SocialProfile
				if ok && followingValue != nil {
					followingMapList := followingValue.([]interface{})
					for _, followingMap := range followingMapList {
						if followingMap != nil {
							userID, userIDOK := followingMap.(map[string]interface{})["userId"].(int64)
							username, usernameOK := followingMap.(map[string]interface{})["username"].(string)
							if userIDOK && usernameOK {
								followingUser := &model.SocialProfile{
									UserID:   userID,
									Username: username,
								}
								following = append(following, followingUser)
							}
						}
					}
				}

				socialProfile := &model.SocialProfileData{
					UserID:    userID.(int64),
					Username:  username.(string),
					Followers: followers,
					Following: following,
				}
				return socialProfile, nil
			}

			return nil, nil
		})
	if err != nil {
		mr.logger.Println("Error querying social profile by userId:", err)
		return nil, err
	}

	if result == nil {
		return nil, nil
	}

	return result.(*model.SocialProfileData), nil
}

func (mr *SocialProfileRepo) GetAllFollowers(userId int64) (model.SocialProfiles, error) {
	ctx := context.Background()
	session := mr.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: "neo4j"})
	defer session.Close(ctx)

	// ExecuteRead for read transactions (Read and queries)
	followerResults, err := session.ExecuteRead(ctx,
		func(transaction neo4j.ManagedTransaction) (any, error) {
			result, err := transaction.Run(ctx,
				`MATCH (follower:SocialProfile)-[:FOLLOWS]->(user:SocialProfile {userId: $userId})
				RETURN follower.userId as userId, follower.username as username`,
				map[string]any{"userId": userId})
			if err != nil {
				return nil, err
			}

			var followers model.SocialProfiles
			for result.Next(ctx) {
				record := result.Record()
				followerUserID, ok := record.Get("userId")
				if !ok || followerUserID == nil {
					followerUserID = 0
				}
				username, _ := record.Get("username")
				followers = append(followers, &model.SocialProfile{
					UserID:   followerUserID.(int64),
					Username: username.(string),
				})
			}
			return followers, nil
		})
	if err != nil {
		mr.logger.Println("Error querying followers:", err)
		return nil, err
	}
	return followerResults.(model.SocialProfiles), nil
}

func (mr *SocialProfileRepo) GetAllFollowing(userId int64) (model.SocialProfiles, error) {
	ctx := context.Background()
	session := mr.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: "neo4j"})
	defer session.Close(ctx)

	// ExecuteRead for read transactions (Read and queries)
	followingResults, err := session.ExecuteRead(ctx,
		func(transaction neo4j.ManagedTransaction) (any, error) {
			result, err := transaction.Run(ctx,
				`MATCH (user:SocialProfile {userId: $userId})-[:FOLLOWS]->(following:SocialProfile)
				RETURN following.userId as userId, following.username as username`,
				map[string]any{"userId": userId})
			if err != nil {
				return nil, err
			}

			var following model.SocialProfiles
			for result.Next(ctx) {
				record := result.Record()
				followingUserID, ok := record.Get("userId")
				if !ok || followingUserID == nil {
					followingUserID = 0
				}
				username, _ := record.Get("username")
				following = append(following, &model.SocialProfile{
					UserID:   followingUserID.(int64),
					Username: username.(string),
				})
			}
			return following, nil
		})
	if err != nil {
		mr.logger.Println("Error querying following:", err)
		return nil, err
	}
	return followingResults.(model.SocialProfiles), nil
}

func (mr *SocialProfileRepo) CheckFollowRelationshipExists(userId int64, followerId int64) (bool, error) {
	ctx := context.Background()
	session := mr.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: "neo4j"})
	defer session.Close(ctx)

	result, err := session.Run(ctx,
		"MATCH (user:SocialProfile {userId: $userId})<-[:FOLLOWS]-(follower:SocialProfile {userId: $followerId}) "+
			"RETURN COUNT(*)",
		map[string]any{"userId": userId, "followerId": followerId})
	if err != nil {
		return false, err
	}

	if result.Next(ctx) {
		record := result.Record()
		countValue, _ := record.Get("COUNT(*)")
		return countValue.(int64) > 0, nil
	}

	return false, nil
}

func (mr *SocialProfileRepo) Follow(userId int64, followerId int64) error {
	ctx := context.Background()
	session := mr.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: "neo4j"})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx,
		func(transaction neo4j.ManagedTransaction) (any, error) {
			_, err := transaction.Run(ctx,
				"MATCH (user:SocialProfile {userId: $userId}), (follower:SocialProfile {userId: $followerId}) "+
					"CREATE (follower)-[:FOLLOWS]->(user)",
				map[string]any{"userId": userId, "followerId": followerId})
			if err != nil {
				return nil, err
			}

			return nil, nil
		})

	if err != nil {
		mr.logger.Println("Error creating 'follows' relationship:", err)
		return err
	}

	return nil
}

func (mr *SocialProfileRepo) Unfollow(userId int64, followerId int64) error {
	ctx := context.Background()
	session := mr.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: "neo4j"})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx,
		func(transaction neo4j.ManagedTransaction) (any, error) {
			//result, err := transaction.Run(ctx,
			_, err := transaction.Run(ctx,
				"MATCH (user:SocialProfile {userId: $userId})<-[rel:FOLLOWS]-(follower:SocialProfile {userId: $followerId}) "+
					"DELETE rel",
				map[string]any{"userId": userId, "followerId": followerId})
			if err != nil {
				return nil, err
			}

			// // Check if the relationship was deleted
			// if result.Summary().Counters().RelationshipsDeleted() == 0 {
			// 	// Relationship didn't exist
			// 	return nil, fmt.Errorf("relationship not found")
			// }

			return nil, nil
		})

	if err != nil {
		mr.logger.Println("Error deleting 'FOLLOWS' relationship:", err)
		return err
	}

	return nil
}

func (mr *SocialProfileRepo) SearchSocialProfilesByUsername(searchedUsername string) (model.SocialProfiles, error) {
	ctx := context.Background()
	session := mr.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: "neo4j"})
	defer session.Close(ctx)

	// ExecuteRead for read transactions (Read and queries)
	searchResults, err := session.ExecuteRead(ctx,
		func(transaction neo4j.ManagedTransaction) (any, error) {
			result, err := transaction.Run(ctx,
				`MATCH (profile:SocialProfile)
                WHERE TOLOWER(profile.username) CONTAINS TOLOWER($searchedUsername)
                RETURN profile.userId as userId, profile.username as username`,
				map[string]any{"searchedUsername": searchedUsername})
			if err != nil {
				return nil, err
			}

			var profiles model.SocialProfiles
			for result.Next(ctx) {
				record := result.Record()
				userID, ok := record.Get("userId")
				if !ok || userID == nil {
					userID = 0
				}
				username, _ := record.Get("username")
				profiles = append(profiles, &model.SocialProfile{
					UserID:   userID.(int64),
					Username: username.(string),
				})
			}
			return profiles, nil
		})
	if err != nil {
		mr.logger.Println("Error searching social profiles:", err)
		return nil, err
	}
	return searchResults.(model.SocialProfiles), nil
}
