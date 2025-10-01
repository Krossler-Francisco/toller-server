package friends

import "database/sql"

type FriendRepository struct {
	DB *sql.DB
}

func NewFriendRepository(db *sql.DB) *FriendRepository {
	return &FriendRepository{DB: db}
}

func (r *FriendRepository) CreateFriendRequest(userID, friendID int) error {
	_, err := r.DB.Exec("INSERT INTO friends (user_id, friend_id) VALUES ($1, $2)", userID, friendID)
	return err
}

func (r *FriendRepository) UpdateFriendRequest(userID, friendID int, status string) error {
	_, err := r.DB.Exec("UPDATE friends SET status = $1 WHERE user_id = $2 AND friend_id = $3", status, friendID, userID)
	return err
}

func (r *FriendRepository) ListFriends(userID int) ([]Friend, error) {
	rows, err := r.DB.Query("SELECT user_id, friend_id, status FROM friends WHERE (user_id = $1 OR friend_id = $1) AND status = 'accepted'", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var friends []Friend
	for rows.Next() {
		var friend Friend
		if err := rows.Scan(&friend.UserID, &friend.FriendID, &friend.Status); err != nil {
			return nil, err
		}
		friends = append(friends, friend)
	}

	return friends, nil
}

func (r *FriendRepository) ListPendingRequests(userID int) ([]Friend, error) {
	rows, err := r.DB.Query("SELECT user_id, friend_id, status FROM friends WHERE friend_id = $1 AND status = 'pending'", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requests []Friend
	for rows.Next() {
		var req Friend
		if err := rows.Scan(&req.UserID, &req.FriendID, &req.Status); err != nil {
			return nil, err
		}
		requests = append(requests, req)
	}

	return requests, nil
}
