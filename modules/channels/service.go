// modules/channels/service.go
package channels

import (
	"errors"
)

type ChannelService struct {
	Repo *ChannelRepository
}

// CreateChannel cria um novo canal
func (s *ChannelService) CreateChannel(name string, teamID, creatorID int) (*Channel, error) {
	if name == "" {
		return nil, errors.New("nome do canal é obrigatório")
	}

	// Verificar se o usuário pertence ao time
	inTeam, err := s.Repo.IsUserInTeam(creatorID, teamID)
	if err != nil {
		return nil, err
	}
	if !inTeam {
		return nil, errors.New("usuário não pertence ao time")
	}

	return s.Repo.CreateChannel(name, teamID, creatorID)
}

// GetChannelsByTeam retorna os canais de um time para o usuário
func (s *ChannelService) GetChannelsByTeam(teamID, userID int) ([]ChannelWithRole, error) {
	// Verificar se o usuário pertence ao time
	inTeam, err := s.Repo.IsUserInTeam(userID, teamID)
	if err != nil {
		return nil, err
	}
	if !inTeam {
		return nil, errors.New("usuário não pertence ao time")
	}

	return s.Repo.GetChannelsByTeam(teamID, userID)
}

// GetChannelByID retorna um canal específico
func (s *ChannelService) GetChannelByID(channelID, userID int) (*Channel, error) {
	// Verificar se o usuário tem acesso ao canal
	inChannel, err := s.Repo.IsUserInChannel(userID, channelID)
	if err != nil {
		return nil, err
	}
	if !inChannel {
		return nil, errors.New("usuário não tem acesso a este canal")
	}

	return s.Repo.GetChannelByID(channelID)
}

// AddMemberToChannel adiciona um membro ao canal
func (s *ChannelService) AddMemberToChannel(channelID, userID, requestingUserID int, role string) error {
	// Verificar se o usuário que está fazendo a requisição é admin do canal
	isAdmin, err := s.Repo.IsUserChannelAdmin(requestingUserID, channelID)
	if err != nil {
		return err
	}
	if !isAdmin {
		return errors.New("apenas administradores podem adicionar membros")
	}

	// Verificar se o canal existe
	channel, err := s.Repo.GetChannelByID(channelID)
	if err != nil {
		return err
	}

	// Verificar se o usuário a ser adicionado pertence ao time
	inTeam, err := s.Repo.IsUserInTeam(userID, channel.TeamID)
	if err != nil {
		return err
	}
	if !inTeam {
		return errors.New("usuário não pertence ao time deste canal")
	}

	// Validar role
	if role != "admin" && role != "user" {
		role = "user"
	}

	return s.Repo.AddUserToChannel(userID, channelID, role)
}

// RemoveMemberFromChannel remove um membro do canal
func (s *ChannelService) RemoveMemberFromChannel(channelID, userID, requestingUserID int) error {
	// Verificar se o usuário que está fazendo a requisição é admin do canal
	isAdmin, err := s.Repo.IsUserChannelAdmin(requestingUserID, channelID)
	if err != nil {
		return err
	}
	if !isAdmin {
		return errors.New("apenas administradores podem remover membros")
	}

	// Não permitir que o último admin se remova
	members, err := s.Repo.GetChannelMembers(channelID)
	if err != nil {
		return err
	}

	adminCount := 0
	for _, member := range members {
		if member.Role == "admin" {
			adminCount++
		}
	}

	if adminCount == 1 && userID == requestingUserID {
		return errors.New("não é possível remover o último administrador do canal")
	}

	return s.Repo.RemoveUserFromChannel(userID, channelID)
}

// GetChannelMembers retorna os membros de um canal
func (s *ChannelService) GetChannelMembers(channelID, requestingUserID int) ([]ChannelMember, error) {
	// Verificar se o usuário tem acesso ao canal
	inChannel, err := s.Repo.IsUserInChannel(requestingUserID, channelID)
	if err != nil {
		return nil, err
	}
	if !inChannel {
		return nil, errors.New("usuário não tem acesso a este canal")
	}

	return s.Repo.GetChannelMembers(channelID)
}

// DeleteChannel remove um canal
func (s *ChannelService) DeleteChannel(channelID, userID int) error {
	// Verificar se o usuário é admin do canal
	isAdmin, err := s.Repo.IsUserChannelAdmin(userID, channelID)
	if err != nil {
		return err
	}
	if !isAdmin {
		return errors.New("apenas administradores podem deletar o canal")
	}

	return s.Repo.DeleteChannel(channelID)
}

// UpdateChannelName atualiza o nome do canal
func (s *ChannelService) UpdateChannelName(channelID int, newName string, userID int) error {
	if newName == "" {
		return errors.New("nome do canal não pode ser vazio")
	}

	// Verificar se o usuário é admin do canal
	isAdmin, err := s.Repo.IsUserChannelAdmin(userID, channelID)
	if err != nil {
		return err
	}
	if !isAdmin {
		return errors.New("apenas administradores podem renomear o canal")
	}

	query := `UPDATE channels SET name = $1 WHERE id = $2`
	_, err = s.Repo.DB.Exec(query, newName, channelID)
	return err
}
