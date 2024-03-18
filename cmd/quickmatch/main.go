package main

import (
	"errors"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"quick-match/cmd/util"
	"quick-match/internal/clients"
	"quick-match/internal/handlers/discover"
	"quick-match/internal/handlers/login"
	"quick-match/internal/handlers/swipe"
	"quick-match/internal/handlers/usercreate"
	"quick-match/internal/middleware/authentication"
	"quick-match/internal/repository"
	"time"
)

const port = "8080"
const EsDomainName = "quickmatch-discover"

func main() {
	r := mux.NewRouter()

	dynamoDBClient := clients.NewDynamoDBClient()
	dc := repository.NewDynamoDBRepository(dynamoDBClient)

	esClient := clients.NewElasticsearchClient(EsDomainName)
	esc := repository.NewElasticSearchClient(esClient)

	// Ensuring Elasticsearch index and mappings are correctly set up
	esc.EnsureElasticsearchSetup()

	ud := util.NewUserCreateService(dc, esc)
	r.HandleFunc("/user/create", usercreate.CreateUserHandler(ud)).Methods("POST")

	ld := util.NewLoginService(dc)
	r.HandleFunc("/login", login.LoginHandler(ld)).Methods("POST")

	sd := util.NewSwipeService(dc)
	r.Handle("/swipe", authentication.JWTMiddleware(swipe.SwipeHandler(sd))).Methods("POST")

	dd := util.NewDiscoverService(dc, esc)
	r.Handle("/discover", authentication.JWTMiddleware(discover.DiscoverUserInsert(dd))).Methods("POST")

	// Server Setup
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	log.Printf("Starting server on port %s", port)
	if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("Server failed: %v", err)
	}
}
