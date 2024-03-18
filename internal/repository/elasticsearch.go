package repository

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/elastic/go-elasticsearch/v7"
	"github.com/mitchellh/mapstructure"
	"log"
	"quick-match/internal/models"
	"strings"
)

type ElasticSearchRepository struct {
	EsClient *elasticsearch.Client
}

func NewElasticSearchClient(client *elasticsearch.Client) ElasticSearchRepository {
	return ElasticSearchRepository{
		EsClient: client,
	}
}

func CreateElasticSearchUser(user models.UserDetails) models.UserDetailsES {
	return models.UserDetailsES{
		UserID: user.UserID,
		Name:   user.Name,
		Gender: user.Gender,
		Age:    user.Age,
		Location: models.UserLocationES{
			Lat: user.Latitude,
			Lon: user.Longitude,
		},
	}
}

type BoolQuery struct {
	MustNot []interface{} `json:"must_not,omitempty"`
	Filter  []interface{} `json:"filter,omitempty"`
}

type Query struct {
	Query struct {
		Bool BoolQuery `json:"bool"`
	} `json:"query"`
}

func NewQuery() *Query {
	return &Query{}
}

/*
EnsureElasticsearchSetup checks and ensures the necessary Elasticsearch index setup for user data.
This method specifically checks if the "users" index exists in the Elasticsearch database. If it does not exist,
it creates the index with predefined mappings for the user properties such as UserID, name, gender, age, and location.
These mappings help in optimizing search queries and aggregations on the user data.
*/
func (repo *ElasticSearchRepository) EnsureElasticsearchSetup() {
	indexName := "users"
	mappings := `{
		"mappings": {
			"properties": {
				"UserID": { "type": "keyword" },
				"name": { "type": "text" },
				"gender": { "type": "keyword" },
				"age": { "type": "integer" },
				"location": { "type": "geo_point" }
			}
		}
	}`

	// Check if index exists
	res, err := repo.EsClient.Indices.Exists([]string{indexName})
	if err != nil || res.StatusCode == 404 {
		res, err = repo.EsClient.Indices.Create(
			indexName,
			repo.EsClient.Indices.Create.WithBody(strings.NewReader(mappings)),
			repo.EsClient.Indices.Create.WithContext(context.Background()),
		)
		if err != nil || res.IsError() {
			log.Fatalf("Failed to create Elasticsearch index: %v", err)
		}
		log.Println("Elasticsearch index created successfully")
	} else {
		log.Println("Elasticsearch index already exists")
	}
}

func (repo *ElasticSearchRepository) InsertUserES(user models.UserDetailsES) error {
	userJSON, err := json.Marshal(user)
	if err != nil {
		return err
	}

	res, err := repo.EsClient.Index(
		"users",
		strings.NewReader(string(userJSON)),
		repo.EsClient.Index.WithDocumentID(user.UserID), // Using UserID as the document ID
		repo.EsClient.Index.WithRefresh("true"),
	)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("error indexing document ID=%s", user.UserID)
	}
	log.Println("Document indexed successfully, ID:", user.UserID)
	return nil
}

func (repo *ElasticSearchRepository) GetUserByID(userID string) (models.UserDetailsES, error) {
	var user models.UserDetailsES

	res, err := repo.EsClient.Get(
		"users",
		userID, // The document ID, which is the UserID in this case
	)
	if err != nil {
		// Handle the error
		return user, err
	}
	defer res.Body.Close()

	if res.IsError() {
		// Handle the error, possibly the user not found
		return user, fmt.Errorf("Error fetching user with ID %s: %s", userID, res.String())
	}

	var r map[string]any
	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		return user, fmt.Errorf("Error parsing the response body: %s", err)
	}

	// Extract the source which contains the user data
	if source, found := r["_source"]; found {
		if err := mapstructure.Decode(source, &user); err != nil {
			return user, fmt.Errorf("Error decoding source into user model: %v", err)
		}
	}

	return user, nil
}

func (q *Query) AddGenderFilter(gender string) {
	if gender != "" {
		q.Query.Bool.Filter = append(q.Query.Bool.Filter, map[string]any{
			"term": map[string]any{
				"gender": gender,
			},
		})
	}
}

func (q *Query) AddAgeRangeFilter(minAge, maxAge int) {
	if maxAge > 0 || minAge > 0 {
		ageQuery := map[string]any{
			"range": map[string]any{
				"age": map[string]interface{}{},
			},
		}
		if minAge > 0 {
			ageQuery["range"].(map[string]any)["age"].(map[string]any)["gte"] = minAge
		}
		if maxAge > 0 {
			ageQuery["range"].(map[string]any)["age"].(map[string]any)["lte"] = maxAge
		}
		q.Query.Bool.Filter = append(q.Query.Bool.Filter, ageQuery)
	}
}

func (q *Query) AddGeoDistanceFilter(location models.UserLocationES, maxDistanceKM int) {
	if maxDistanceKM > 0 {
		q.Query.Bool.Filter = append(q.Query.Bool.Filter, map[string]any{
			"geo_distance": map[string]any{
				"distance": fmt.Sprintf("%dkm", maxDistanceKM),
				"location": location,
			},
		})
	}
}

func (q *Query) AddExclusionFilter(ids []string) {
	if len(ids) > 0 {
		q.Query.Bool.MustNot = append(q.Query.Bool.MustNot, map[string]any{
			"ids": map[string]any{
				"type":   "_doc",
				"values": ids,
			},
		})
	}
}

/*
SearchUsers performs a filtered search on the user data stored in Elasticsearch based on the given filters:
currentUserLocation, swipedUserIDs, and discover filters. It constructs a query that excludes users already swiped on,
matches the specified gender and age range, and is within the maximum distance from the currentUserLocation. Any combination of
filters can be added. This function returns a list of users that match the specified criteria or an error if the search fails.

Parameters:
- currentUserLocation: The geographical location of the current user performing the discovery.
- swipedUserIDs: A list of user IDs that the current user has already swiped on, to be excluded from the search results.
- discover: Filters specifying the criteria for the user discovery such as gender preference, age range, and maximum distance.

Returns:
- A slice of UserDetailsES models representing the users who match the search criteria.
- An error if the search operation fails or if there is an issue parsing the response from Elasticsearch.
*/
func (repo *ElasticSearchRepository) SearchUsers(currentUserLocation models.UserLocationES, swipedUserIDs []string, discover models.DiscoverFilters) ([]models.UserDetailsES, error) {
	var buf bytes.Buffer

	query := NewQuery()
	query.AddExclusionFilter(swipedUserIDs)
	query.AddGenderFilter(discover.Gender)
	query.AddAgeRangeFilter(discover.MinAge, discover.MaxAge)
	query.AddGeoDistanceFilter(currentUserLocation, discover.MaxLocation)

	// Encode the query into JSON
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		return nil, fmt.Errorf("error encoding query: %v", err)
	}

	res, err := repo.EsClient.Search(
		repo.EsClient.Search.WithContext(context.Background()),
		repo.EsClient.Search.WithIndex("users"),
		repo.EsClient.Search.WithBody(&buf),
		repo.EsClient.Search.WithTrackTotalHits(true),
	)
	if err != nil {
		return nil, fmt.Errorf("error getting response: %s", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		var e map[string]any
		if err = json.NewDecoder(res.Body).Decode(&e); err != nil {
			return nil, fmt.Errorf("error parsing the response body: %s", err)
		}
		return nil, fmt.Errorf("[%s] %s: %s", res.Status(), e["error"].(map[string]interface{})["type"], e["error"].(map[string]interface{})["reason"])
	}

	// Parse the response
	var r map[string]any
	if err = json.NewDecoder(res.Body).Decode(&r); err != nil {
		return nil, fmt.Errorf("error parsing the response body: %s", err)
	}
	fmt.Printf("Successful query. Number of hits: %d\n", int(r["hits"].(map[string]interface{})["total"].(map[string]interface{})["value"].(float64)))

	// Deserialize the hits into model
	var users []models.UserDetailsES
	for _, hit := range r["hits"].(map[string]any)["hits"].([]any) {
		var user models.UserDetailsES
		hitSource := hit.(map[string]any)["_source"].(map[string]any)
		if err = mapstructure.Decode(hitSource, &user); err != nil {
			return nil, fmt.Errorf("error decoding hit source: %v", err)
		}
		users = append(users, user)
	}

	return users, nil
}
