package repository

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"log"
	"quick-match/internal/models"
)

const usersTable = "quickmatch_users"
const swipesTable = "quickmatch_swipes"

type DynamoDBRepository struct {
	Client *dynamodb.DynamoDB
}

func NewDynamoDBRepository(client *dynamodb.DynamoDB) DynamoDBRepository {
	return DynamoDBRepository{
		Client: client,
	}
}

func (repo *DynamoDBRepository) InsertUser(user models.UserDetails) error {
	av, err := dynamodbattribute.MarshalMap(user)
	if err != nil {
		fmt.Printf("Failed to marshal user: %v\n", err)
		return err
	}

	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(usersTable),
	}

	_, err = repo.Client.PutItem(input)
	if err != nil {
		fmt.Printf("Failed to insert user into DynamoDB: %v\n", err)
		return err
	}

	return nil
}

func (repo *DynamoDBRepository) GetUserByEmail(email string) (*models.UserDetails, error) {
	f := expression.Name("email").Equal(expression.Value(email))
	e, err := expression.NewBuilder().WithFilter(f).Build()
	if err != nil {
		return nil, err
	}

	input := &dynamodb.ScanInput{
		ExpressionAttributeNames:  e.Names(),
		ExpressionAttributeValues: e.Values(),
		FilterExpression:          e.Filter(),
		TableName:                 aws.String(usersTable),
	}

	result, err := repo.Client.Scan(input)
	if err != nil {
		return nil, err
	}

	if len(result.Items) == 0 {
		return nil, nil // UserDetails not found
	}

	var user models.UserDetails
	err = dynamodbattribute.UnmarshalMap(result.Items[0], &user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (repo *DynamoDBRepository) InsertSwipeRecord(swipe models.Swipe) error {
	av, err := dynamodbattribute.MarshalMap(swipe)
	if err != nil {
		return err
	}

	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(swipesTable),
	}

	_, err = repo.Client.PutItem(input)
	return err
}

/*
CheckSwipeMatch queries the DynamoDB to check if a mutual "like" exists between two users, indicating a match.

This method performs a query using a Global Secondary Index (GSI) to efficiently find if the user identified by
`swipedUserID` has swiped right (liked) on the user identified by `currentUserID`. It specifically looks for records
where the `SwipedUserID` equals `currentUserID` and `UserID` equals `swipedUserID`, indicating that the swiped user
has previously swiped on the current user.

The method constructs a key condition expression to query the `swipesTable` using the `SwipedUserIndex` GSI. It
only retrieves the `preference` attribute of the swipe record to determine if the swiped user has liked the current
user. If a swipe record with `preference` set to true is found, the function returns true, indicating a match. If no
such record is found or if the query does not return any items, it assumes that no swipe has occurred yet and returns
false. Errors during the query execution or result unmarshalling are returned to the caller.

Parameters:
- swipedUserID: The ID of the user who is being checked to see if they have liked the current user.
- currentUserID: The ID of the current user performing the check.

Returns:
- A boolean indicating whether a mutual like exists (true if there is a match, false otherwise).
- An error if the query fails to execute or if there is an issue unmarshalling the query result.
*/
func (repo *DynamoDBRepository) CheckSwipeMatch(swipedUserID, currentUserID string) (bool, error) {
	// Using GSI to efficiently query swipes where swipedUserID has swiped on currentUserID
	keyCond := expression.Key("SwipedUserID").Equal(expression.Value(currentUserID)).
		And(expression.Key("UserID").Equal(expression.Value(swipedUserID)))

	expr, err := expression.NewBuilder().WithKeyCondition(keyCond).Build()
	if err != nil {
		return false, err
	}

	queryInput := &dynamodb.QueryInput{
		TableName:                 aws.String(swipesTable),
		IndexName:                 aws.String("SwipedUserIndex"), // Using the GSI
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		KeyConditionExpression:    expr.KeyCondition(),
		ProjectionExpression:      aws.String("preference"), // Assuming we're interested in the Preference attribute
	}

	result, err := repo.Client.Query(queryInput)
	if len(result.Items) == 0 {
		log.Println("No existing swipe record found for the swiped user on current user. Assuming no swipe yet.")
		return false, nil
	}
	if err != nil {
		return false, err
	}

	// Checking if there is a swipe record with Preference = true
	for _, item := range result.Items {
		var swipe models.Swipe
		err = dynamodbattribute.UnmarshalMap(item, &swipe)
		if err != nil {
			return false, err
		}

		if swipe.Preference { // Assuming Swipe struct has a Preference field of type bool
			return true, nil // Found a match
		}
	}

	return false, nil
}

func (repo *DynamoDBRepository) GetSwipedUserIDs(userID string) ([]string, error) {
	keyCond := expression.Key("UserID").Equal(expression.Value(userID))

	expr, err := expression.NewBuilder().WithKeyCondition(keyCond).Build()
	if err != nil {
		return nil, err
	}

	queryInput := &dynamodb.QueryInput{
		TableName:                 aws.String(swipesTable),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		KeyConditionExpression:    expr.KeyCondition(),
		ProjectionExpression:      aws.String("SwipedUserID"),
	}

	result, err := repo.Client.Query(queryInput)
	if err != nil {
		return nil, err
	}

	var swipedUserIDs []string
	for _, item := range result.Items {
		var swipe models.Swipe
		err = dynamodbattribute.UnmarshalMap(item, &swipe)
		if err != nil {
			return nil, err
		}
		swipedUserIDs = append(swipedUserIDs, swipe.SwipedUserID)
	}

	return swipedUserIDs, nil
}
