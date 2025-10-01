package friends

import "errors"

type FriendService struct {
	Repo *FriendRepository
}

func NewFriendService(repo *FriendRepository) *FriendService {
	return &FriendService{Repo: repo}
}

func (s *FriendService) SendFriendRequest(userID, friendID int) error {
	if userID == friendID {
		return errors.New("cannot send friend request to yourself")
	}
	return s.Repo.CreateFriendRequest(userID, friendID)
}

func (s *FriendService) UpdateFriendRequest(userID, friendID int, status string) error {
	if status != "accepted" && status != "blocked" {
		return errors.New("invalid status")
	}
	return s.Repo.UpdateFriendRequest(userID, friendID, status)
}

func (s *FriendService) ListFriends(userID int) ([]Friend, error) {
	return s.Repo.ListFriends(userID)
}

func (s *FriendService) ListPendingRequests(userID int) ([]Friend, error) {
	return s.Repo.ListPendingRequests(userID)
}
