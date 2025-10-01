package dms

import "errors"

type DMService struct {
	Repo *DMRepository
}

func NewDMService(repo *DMRepository) *DMService {
	return &DMService{Repo: repo}
}

func (s *DMService) CreateDM(user1ID, user2ID int) (int, error) {
	return s.Repo.CreateDMChannel(user1ID, user2ID)
}

// ListDMs devuelve la lista de conversaciones de DM de un usuario.
func (s *DMService) ListDMs(userID int) ([]DMChannelInfo, error) {
	return s.Repo.ListDMChannels(userID)
}

func (s *DMService) GetMessages(userID, channelID int) ([]Message, error) {
	isMember, err := s.Repo.IsUserInDMChannel(userID, channelID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, errors.New("user is not a member of this DM channel")
	}
	return s.Repo.GetMessagesByChannelID(channelID)
}

func (s *DMService) MarkAsRead(userID, channelID int) error {
	isMember, err := s.Repo.IsUserInDMChannel(userID, channelID)
	if err != nil {
		return err
	}
	if !isMember {
		return errors.New("user is not a member of this DM channel")
	}
	return s.Repo.MarkChannelAsRead(userID, channelID)
}