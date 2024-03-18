package models

type Swipe struct {
	UserID       string `json:"UserID" dynamodbav:"UserID"`
	SwipedUserID string `json:"SwipedUserID" dynamodbav:"SwipedUserID"`
	Preference   bool   `json:"preference" dynamodbav:"preference"`
	Matched      bool   `json:"matched" dynamodbav:"matched"`
	MatchID      string `json:"matchId,omitempty" dynamodbav:"matchId"`
}

type SwipeResponse struct {
	Matched bool   `json:"matched"`
	MatchID string `json:"matchId,omitempty"`
}
