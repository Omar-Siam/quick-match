package util

import (
	"context"
	"fmt"
	"github.com/elastic/go-elasticsearch/v7"
	"log"
	"quick-match/internal/handlers/discover"
	"quick-match/internal/handlers/login"
	"quick-match/internal/handlers/swipe"
	"quick-match/internal/handlers/usercreate"
	"quick-match/internal/middleware/authentication"
	"quick-match/internal/repository"
	"quick-match/internal/services"
)

func NewLoginService(ddb repository.DynamoDBRepository) *login.LoginDeps {
	tokenService := authentication.NewJWTTokenService()
	passwordService := services.NewBcryptPasswordService()

	return &login.LoginDeps{
		UserRepo:        &ddb,
		TokenService:    tokenService,
		PasswordService: passwordService,
	}
}

func NewUserCreateService(ddb repository.DynamoDBRepository, es repository.ElasticSearchRepository) *usercreate.CreateUserDeps {
	return &usercreate.CreateUserDeps{
		UserRepoES: &es,
		UserRepo:   &ddb,
	}
}

func NewSwipeService(ddb repository.DynamoDBRepository) *swipe.SwipeDeps {
	return &swipe.SwipeDeps{
		SwipeRepo: &ddb,
	}
}

func NewDiscoverService(ddb repository.DynamoDBRepository, es repository.ElasticSearchRepository) *discover.DiscoverUserDeps {
	return &discover.DiscoverUserDeps{
		UserRepo:   &ddb,
		UserRepoES: &es,
	}
}

func PrintAllUsers(esClient *elasticsearch.Client) {
	res, err := esClient.Search(
		esClient.Search.WithContext(context.Background()),
		esClient.Search.WithIndex("users"),
		esClient.Search.WithPretty(),
	)
	if err != nil {
		log.Fatalf("Error getting response: %s", err)
	}
	defer res.Body.Close()

	fmt.Println(res.String())
}
