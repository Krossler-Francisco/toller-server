package friends

// Friend represents a friend relationship.
type Friend struct {
	UserID   int    `json:"user_id"`
	FriendID int    `json:"friend_id"`
	Status   string `json:"status"`
}
