package teams

import "errors"

type TeamService struct {
	Repo *TeamRepository
}

// Crear un team (el creador se convierte en admin automáticamente)
func (s *TeamService) CreateTeam(name, description string, creatorID int) (*Team, error) {
	if name == "" {
		return nil, errors.New("el nombre del equipo es requerido")
	}

	team := &Team{
		Name:        name,
		Description: description,
	}

	// Crear el team
	err := s.Repo.CreateTeam(team)
	if err != nil {
		return nil, err
	}

	// Agregar al creador como admin
	err = s.Repo.AddUserToTeam(creatorID, team.ID, "admin")
	if err != nil {
		return nil, err
	}

	return team, nil
}

// Obtener los teams de un usuario
func (s *TeamService) GetUserTeams(userID int) ([]TeamWithRole, error) {
	return s.Repo.GetUserTeams(userID)
}

// Obtener un team por ID (solo si el usuario es miembro)
func (s *TeamService) GetTeam(teamID, userID int) (*Team, error) {
	// Verificar que el usuario sea miembro
	isMember, _, err := s.Repo.IsUserInTeam(userID, teamID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, errors.New("no tienes acceso a este equipo")
	}

	return s.Repo.GetTeamByID(teamID)
}

// Agregar miembro al team (solo admins pueden hacerlo)
func (s *TeamService) AddMember(teamID, newUserID, requestingUserID int) error {
	// Verificar que quien lo solicita sea admin
	isMember, role, err := s.Repo.IsUserInTeam(requestingUserID, teamID)
	if err != nil {
		return err
	}
	if !isMember || role != "admin" {
		return errors.New("solo los admins pueden agregar miembros")
	}

	// Verificar que el nuevo usuario no esté ya en el team
	alreadyMember, _, err := s.Repo.IsUserInTeam(newUserID, teamID)
	if err != nil {
		return err
	}
	if alreadyMember {
		return errors.New("el usuario ya es miembro del equipo")
	}

	return s.Repo.AddUserToTeam(newUserID, teamID, "member")
}

// Obtener miembros del team
func (s *TeamService) GetTeamMembers(teamID, userID int) ([]TeamMember, error) {
	// Verificar que el usuario sea miembro
	isMember, _, err := s.Repo.IsUserInTeam(userID, teamID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, errors.New("no tienes acceso a este equipo")
	}

	return s.Repo.GetTeamMembers(teamID)
}

// Remover miembro del team (solo admins)
func (s *TeamService) RemoveMember(teamID, userToRemove, requestingUserID int) error {
	// Verificar que quien lo solicita sea admin
	isMember, role, err := s.Repo.IsUserInTeam(requestingUserID, teamID)
	if err != nil {
		return err
	}
	if !isMember || role != "admin" {
		return errors.New("solo los admins pueden remover miembros")
	}

	// No puede removerse a sí mismo si es admin
	if userToRemove == requestingUserID {
		return errors.New("no puedes removerte a ti mismo del equipo")
	}

	return s.Repo.RemoveUserFromTeam(userToRemove, teamID)
}

// Actualizar rol de un miembro (solo admins)
func (s *TeamService) UpdateMemberRole(teamID, targetUserID, requestingUserID int, newRole string) error {
	// Verificar que quien lo solicita sea admin
	isMember, role, err := s.Repo.IsUserInTeam(requestingUserID, teamID)
	if err != nil {
		return err
	}
	if !isMember || role != "admin" {
		return errors.New("solo los admins pueden cambiar roles")
	}

	// Validar el nuevo rol
	if newRole != "admin" && newRole != "member" {
		return errors.New("rol inválido, debe ser 'admin' o 'member'")
	}

	return s.Repo.UpdateUserRole(targetUserID, teamID, newRole)
}

// Actualizar team (solo admins)
func (s *TeamService) UpdateTeam(teamID, userID int, name, description string) error {
	// Verificar que el usuario sea admin
	isMember, role, err := s.Repo.IsUserInTeam(userID, teamID)
	if err != nil {
		return err
	}
	if !isMember || role != "admin" {
		return errors.New("solo los admins pueden actualizar el equipo")
	}

	if name == "" {
		return errors.New("el nombre del equipo es requerido")
	}

	team := &Team{
		ID:          teamID,
		Name:        name,
		Description: description,
	}

	return s.Repo.UpdateTeam(team)
}

// Salir del team (dejar el equipo)
func (s *TeamService) LeaveTeam(teamID, userID int) error {
	// Verificar que el usuario sea miembro
	isMember, _, err := s.Repo.IsUserInTeam(userID, teamID)
	if err != nil {
		return err
	}
	if !isMember {
		return errors.New("no eres miembro de este equipo")
	}

	return s.Repo.RemoveUserFromTeam(userID, teamID)
}
