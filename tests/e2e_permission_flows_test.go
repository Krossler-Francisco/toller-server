
package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"toller-server/modules/teams"
)

// Helper para registrar y loguear un usuario. Retorna el ID de usuario y el token JWT.
func registerAndLogin(t *testing.T, serverURL, username, email, password string) (int, string) {
	// Registrar
	registerData := map[string]string{"username": username, "email": email, "password": password}
	registerBody, _ := json.Marshal(registerData)
	resp, err := http.Post(serverURL+"/register", "application/json", bytes.NewBuffer(registerBody))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode, "Fallo al registrar usuario")
	var registerResp map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&registerResp)
	userID := int(registerResp["id"].(float64))
	resp.Body.Close()

	// Login
	loginData := map[string]string{"email": email, "password": password}
	loginBody, _ := json.Marshal(loginData)
	resp, err = http.Post(serverURL+"/login", "application/json", bytes.NewBuffer(loginBody))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Fallo al iniciar sesión")
	var loginResp map[string]string
	json.NewDecoder(resp.Body).Decode(&loginResp)
	token := loginResp["token"]
	resp.Body.Close()

	assert.NotZero(t, userID)
	assert.NotEmpty(t, token)

	return userID, token
}

// TestTeamMembershipFlow valida el flujo de añadir y quitar un miembro de un equipo.
func TestTeamMembershipFlow(t *testing.T) {
	server, _ := setupTestServer(t)
	client := &http.Client{}

	// 1. Crear Admin y Miembro
	adminEmail := fmt.Sprintf("admin_%d@test.com", time.Now().UnixNano())
	_, adminToken := registerAndLogin(t, server.URL, "adminuser", adminEmail, "password")

	memberEmail := fmt.Sprintf("member_%d@test.com", time.Now().UnixNano())
	memberID, memberToken := registerAndLogin(t, server.URL, "memberuser", memberEmail, "password")

	// 2. Admin crea un equipo
	teamData := map[string]string{"name": "Equipo de Membresía"}
	teamBody, _ := json.Marshal(teamData)
	req, _ := http.NewRequest("POST", server.URL+"/teams", bytes.NewBuffer(teamBody))
	req.Header.Set("Authorization", "Bearer "+adminToken)
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	var teamResp map[string]map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&teamResp)
	teamID := int(teamResp["team"]["id"].(float64))
	resp.Body.Close()

	// 3. Admin añade al Miembro al equipo
	addMemberData := map[string]int{"user_id": memberID}
	addMemberBody, _ := json.Marshal(addMemberData)
	addMemberURL := fmt.Sprintf("%s/teams/%d/members", server.URL, teamID)
	req, _ = http.NewRequest("POST", addMemberURL, bytes.NewBuffer(addMemberBody))
	req.Header.Set("Authorization", "Bearer "+adminToken)
	req.Header.Set("Content-Type", "application/json")
	resp, err = client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Admin debería poder añadir miembro")
	resp.Body.Close()

	// 4. Miembro verifica que pertenece al equipo
	req, _ = http.NewRequest("GET", server.URL+"/teams", nil)
	req.Header.Set("Authorization", "Bearer "+memberToken)
	resp, err = client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var memberTeams []teams.TeamWithRole
	json.NewDecoder(resp.Body).Decode(&memberTeams)
	assert.Len(t, memberTeams, 1, "Miembro debería estar en un equipo")
	assert.Equal(t, teamID, memberTeams[0].ID)
	resp.Body.Close()

	// 5. Admin remueve al Miembro del equipo
	removeMemberURL := fmt.Sprintf("%s/teams/%d/members/%d", server.URL, teamID, memberID)
	req, _ = http.NewRequest("DELETE", removeMemberURL, nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)
	resp, err = client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Admin debería poder remover miembro")
	resp.Body.Close()

	// 6. Miembro verifica que ya no pertenece al equipo
	req, _ = http.NewRequest("GET", server.URL+"/teams", nil)
	req.Header.Set("Authorization", "Bearer "+memberToken)
	resp, err = client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	json.NewDecoder(resp.Body).Decode(&memberTeams)
	assert.Len(t, memberTeams, 0, "Miembro ya no debería estar en ningún equipo")
	resp.Body.Close()
}

// TestChannelAccessPermissions valida la lógica de permisos de acceso a canales.
func TestChannelAccessPermissions(t *testing.T) {
	server, _ := setupTestServer(t)
	client := &http.Client{}

	// 1. Crear usuarios: Admin del equipo y un Miembro regular
	adminEmail := fmt.Sprintf("channel_admin_%d@test.com", time.Now().UnixNano())
	_, adminToken := registerAndLogin(t, server.URL, "channeladmin", adminEmail, "password")

	memberEmail := fmt.Sprintf("channel_member_%d@test.com", time.Now().UnixNano())
	memberID, memberToken := registerAndLogin(t, server.URL, "channelmember", memberEmail, "password")

	// 2. Admin crea un equipo y un canal
	teamData := map[string]string{"name": "Equipo de Permisos"}
	teamBody, _ := json.Marshal(teamData)
	req, _ := http.NewRequest("POST", server.URL+"/teams", bytes.NewBuffer(teamBody))
	req.Header.Set("Authorization", "Bearer "+adminToken)
	req.Header.Set("Content-Type", "application/json")
	resp, _ := client.Do(req)
	var teamResponse map[string]map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&teamResponse)
	teamID := int(teamResponse["team"]["id"].(float64))
	resp.Body.Close()

	channelData := map[string]string{"name": "Canal Secreto"}
	channelBody, _ := json.Marshal(channelData)
	channelURL := fmt.Sprintf(server.URL+"/api/v1/teams/%d/channels", teamID)
	req, _ = http.NewRequest("POST", channelURL, bytes.NewBuffer(channelBody))
	req.Header.Set("Authorization", "Bearer "+adminToken)
	req.Header.Set("Content-Type", "application/json")
	resp, _ = client.Do(req)
	var channelResp map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&channelResp)
	channelID := int(channelResp["id"].(float64))
	resp.Body.Close()

	// 3. Admin añade al Miembro al EQUIPO (pero no al canal)
	addMemberData := map[string]int{"user_id": memberID}
	addMemberBody, _ := json.Marshal(addMemberData)
	addMemberURL := fmt.Sprintf("%s/teams/%d/members", server.URL, teamID)
	req, _ = http.NewRequest("POST", addMemberURL, bytes.NewBuffer(addMemberBody))
	req.Header.Set("Authorization", "Bearer "+adminToken)
	req.Header.Set("Content-Type", "application/json")
	client.Do(req)
	resp.Body.Close()

	// 4. Miembro intenta acceder al canal y DEBERÍA FALLAR (Forbidden)
	getChannelURL := fmt.Sprintf(server.URL+"/api/v1/channels/%d", channelID)
	req, _ = http.NewRequest("GET", getChannelURL, nil)
	req.Header.Set("Authorization", "Bearer "+memberToken)
	resp, _ = client.Do(req)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode, "Miembro no debería poder acceder al canal sin ser parte de él")
	resp.Body.Close()

	// 5. Admin añade al Miembro al CANAL
	addMemberToChannelData := map[string]interface{}{"user_id": memberID, "role": "user"}
	addMemberToChannelBody, _ := json.Marshal(addMemberToChannelData)
	addMemberToChannelURL := fmt.Sprintf(server.URL+"/api/v1/channels/%d/members", channelID)
	req, _ = http.NewRequest("POST", addMemberToChannelURL, bytes.NewBuffer(addMemberToChannelBody))
	req.Header.Set("Authorization", "Bearer "+adminToken)
	req.Header.Set("Content-Type", "application/json")
	resp, _ = client.Do(req)
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Admin debería poder añadir miembro al canal")
	resp.Body.Close()

	// 6. Miembro intenta acceder al canal y DEBERÍA FUNCIONAR
	req, _ = http.NewRequest("GET", getChannelURL, nil)
	req.Header.Set("Authorization", "Bearer "+memberToken)
	resp, _ = client.Do(req)
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Miembro ahora debería tener acceso al canal")
	resp.Body.Close()

	// 7. Miembro (rol 'user') intenta añadir a otra persona y DEBERÍA FALLAR
	otherUserID, _ := registerAndLogin(t, server.URL, "otheruser", fmt.Sprintf("other_%d@test.com", time.Now().UnixNano()), "password")
	addOtherUserData := map[string]int{"user_id": otherUserID}
	addOtherUserBody, _ := json.Marshal(addOtherUserData)
	req, _ = http.NewRequest("POST", addMemberToChannelURL, bytes.NewBuffer(addOtherUserBody))
	req.Header.Set("Authorization", "Bearer "+memberToken) // Usando el token del miembro
	req.Header.Set("Content-Type", "application/json")
	resp, _ = client.Do(req)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode, "Un miembro normal no debería poder añadir a otros al canal")
	resp.Body.Close()
}
