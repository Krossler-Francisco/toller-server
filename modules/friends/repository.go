package friends

import (
	"database/sql"
	"log"
)

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
	log.Printf("[FRIENDS] UpdateFriendRequest: user_id=%d, friend_id=%d, status=%s\n", userID, friendID, status)
	res, err := r.DB.Exec("UPDATE friends SET status = $1 WHERE user_id = $2 AND friend_id = $3", status, userID, friendID)
	if err != nil {
		log.Printf("[FRIENDS] Update error: %v\n", err)
		return err
	}
	rows, _ := res.RowsAffected()
	log.Printf("[FRIENDS] Rows affected (direct): %d\n", rows)
	if rows == 0 {
		// Intentar el update en el sentido inverso
		log.Printf("[FRIENDS] Intentando update inverso: user_id=%d, friend_id=%d\n", friendID, userID)
		res2, err2 := r.DB.Exec("UPDATE friends SET status = $1 WHERE user_id = $2 AND friend_id = $3", status, friendID, userID)
		if err2 != nil {
			log.Printf("[FRIENDS] Update inverso error: %v\n", err2)
			return err2
		}
		rows2, _ := res2.RowsAffected()
		log.Printf("[FRIENDS] Rows affected (inverso): %d\n", rows2)
		if rows2 == 0 {
			return sql.ErrNoRows
		}
	}
	return nil
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
