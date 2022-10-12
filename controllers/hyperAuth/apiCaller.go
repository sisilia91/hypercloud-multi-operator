/*
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package hyperAuth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"

	"github.com/tmax-cloud/hypercloud-multi-operator/controllers/util"
	coreV1 "k8s.io/api/core/v1"
)

func SetServiceDomainURI(serviceName string, urlParameter map[string]string) string {
	for key, value := range urlParameter {
		serviceName = strings.Replace(serviceName, "@@"+key+"@@", value, 1)
	}
	return "https://" + os.Getenv(util.AUTH_SUBDOMAIN) + "." + os.Getenv(util.HC_DOMAIN) + serviceName
}

func GetTokenAsAdmin(secret *coreV1.Secret) (string, error) {
	// Make Body for Content-Type (application/x-www-form-urlencoded)
	id, password := string(secret.Data["HYPERAUTH_ADMIN"]), string(secret.Data["HYPERAUTH_PASSWORD"])
	data := url.Values{}
	data.Set("grant_type", "password")
	data.Set("username", id)
	data.Set("password", password)
	data.Set("client_id", "admin-cli")

	// Make Request Object
	url := SetServiceDomainURI(KEYCLOAK_ADMIN_SERVICE_GET_TOKEN, nil)
	payload := strings.NewReader(data.Encode())
	req, err := http.NewRequest(http.MethodPost, url, payload)
	if err != nil {
		return "", err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	// Request with Client Object
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Result
	bytes, _ := ioutil.ReadAll(resp.Body)
	str := string(bytes)

	var resultJson map[string]interface{}
	if err := json.Unmarshal([]byte(str), &resultJson); err != nil {
		return "", err
	}
	accessToken := resultJson["access_token"].(string)

	return strings.Join([]string{"Bearer", accessToken}, " "), nil
}

func GetIdByClientId(clientId string, secret *coreV1.Secret) (string, error) {
	token, err := GetTokenAsAdmin(secret)
	if err != nil {
		return "", err
	}

	url := SetServiceDomainURI(KEYCLOAK_ADMIN_SERVICE_GET_CLIENTS, nil)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Add("Authorization", token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	reqDump, _ := httputil.DumpRequest(req, true)
	if !IsOK(resp.StatusCode) {
		return "", fmt.Errorf("failed to get client: " + string(reqDump) + "\n" + resp.Status)
	}

	respJson := &[]ClientConfig{}
	err = json.NewDecoder(resp.Body).Decode(respJson)
	if err != nil {
		return "", err
	}

	for _, data := range *respJson {
		if data.ClientId == clientId {
			return data.Id, nil
		}
	}

	return "", HyperAuthError{NotFound: true, Type: RESOURCE_TYPE_CLIENT, Name: clientId}
}

func CreateClient(config ClientConfig, secret *coreV1.Secret) error {
	token, err := GetTokenAsAdmin(secret)
	if err != nil {
		return err
	}

	jsonData, err := json.Marshal(config)
	if err != nil {
		return err
	}
	payload := bytes.NewBuffer(jsonData)

	url := SetServiceDomainURI(KEYCLOAK_ADMIN_SERVICE_CREATE_CLIENT, nil)
	req, err := http.NewRequest(http.MethodPost, url, payload)
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if !IsOK(resp.StatusCode) {
		reqDump, _ := httputil.DumpRequest(req, true)
		return fmt.Errorf("failed to create client: " + string(reqDump) + "\n" + resp.Status)
	}

	return nil
}

func CreateClientLevelProtocolMapper(config ClientLevelProtocolMapperConfig, secret *coreV1.Secret) error {
	token, err := GetTokenAsAdmin(secret)
	if err != nil {
		return err
	}

	id, err := GetIdByClientId(config.ClientId, secret)
	if err != nil {
		return err
	}

	jsonData, err := json.Marshal(config.ProtocolMapper)
	if err != nil {
		return err
	}
	payload := bytes.NewBuffer(jsonData)

	params := map[string]string{
		"id": id,
	}
	url := SetServiceDomainURI(KEYCLOAK_ADMIN_SERVICE_CREATE_CLIENT_PROTOCOL_MAPPERS, params)
	req, err := http.NewRequest(http.MethodPost, url, payload)
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if !IsOK(resp.StatusCode) {
		reqDump, _ := httputil.DumpRequest(req, true)
		return fmt.Errorf("failed to create protocol mapper: " + string(reqDump) + "\n" + resp.Status)
	}

	return nil
}

func CreateClientLevelRole(config ClientLevelRoleConfig, secret *coreV1.Secret) error {
	token, err := GetTokenAsAdmin(secret)
	if err != nil {
		return err
	}

	id, err := GetIdByClientId(config.ClientId, secret)
	if err != nil {
		return err
	}

	data := RoleConfig{
		Name: config.Role.Name,
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	payload := bytes.NewBuffer(jsonData)

	params := map[string]string{
		"id": id,
	}
	url := SetServiceDomainURI(KEYCLOAK_ADMIN_SERVICE_CREATE_CLIENT_ROLES, params)
	req, err := http.NewRequest(http.MethodPost, url, payload)
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if !IsOK(resp.StatusCode) {
		reqDump, _ := httputil.DumpRequest(req, true)
		return fmt.Errorf("failed to create client-level role: " + string(reqDump) + "\n" + resp.Status)
	}

	return nil
}

func GetUserIdByEmail(userEmail string, secret *coreV1.Secret) (string, error) {
	token, err := GetTokenAsAdmin(secret)
	if err != nil {
		return "", err
	}
	params := map[string]string{
		"userEmail": userEmail,
	}

	url := SetServiceDomainURI(KEYCLOAK_ADMIN_SERVICE_GET_USERS_BY_EMAIL, params)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Add("Authorization", token)
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	reqDump, _ := httputil.DumpRequest(req, true)
	if !IsOK(resp.StatusCode) {
		return "", fmt.Errorf("failed to get user by email: " + string(reqDump) + "\n" + resp.Status)
	}

	respJson := &[]UserConfig{}
	err = json.NewDecoder(resp.Body).Decode(respJson)
	if err != nil {
		return "", err
	}

	if respJson == nil || len(*respJson) == 0 {
		return "", HyperAuthError{NotFound: true, Type: RESOURCE_TYPE_USER_EMAIL, Name: userEmail}
	}

	return (*respJson)[0].Id, nil
}

func GetClientRoleIdByRoleName(clientId string, roleName string, secret *coreV1.Secret) (string, error) {
	token, err := GetTokenAsAdmin(secret)
	if err != nil {
		return "", err
	}

	id, err := GetIdByClientId(clientId, secret)
	if err != nil {
		return "", err
	}

	params := map[string]string{
		"id":       id,
		"roleName": roleName,
	}
	url := SetServiceDomainURI(KEYCLOAK_ADMIN_SERVICE_GET_CLIENT_ROLE_BY_NAME, params)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Add("Authorization", token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	reqDump, _ := httputil.DumpRequest(req, true)
	if !IsOK(resp.StatusCode) {
		return "", fmt.Errorf("failed to get role: " + string(reqDump) + "\n" + resp.Status)
	}

	respJson := &RoleConfig{}
	err = json.NewDecoder(resp.Body).Decode(respJson)
	if err != nil {
		return "", err
	}
	if respJson == nil {
		return "", HyperAuthError{NotFound: true, Type: RESOURCE_TYPE_CLIENT_ROLE, Name: roleName}
	}

	return respJson.Id, nil
}

func AddClientLevelRolesToUserRoleMapping(config ClientLevelRoleConfig, userEmail string, secret *coreV1.Secret) error {
	token, err := GetTokenAsAdmin(secret)
	if err != nil {
		return err
	}

	id, err := GetIdByClientId(config.ClientId, secret)
	if err != nil {
		return err
	}

	userId, err := GetUserIdByEmail(userEmail, secret)
	if err != nil {
		return err
	}

	roleId, err := GetClientRoleIdByRoleName(config.ClientId, config.Role.Name, secret)
	if err != nil {
		return err
	}
	data := []RoleConfig{
		{
			Id:   roleId,
			Name: config.Role.Name,
		},
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	payload := bytes.NewBuffer(jsonData)

	params := map[string]string{
		"userId": userId,
		"id":     id,
	}
	url := SetServiceDomainURI(KEYCLOAK_ADMIN_SERVICE_ADD_CLIENT_ROLE_TO_USER, params)
	req, err := http.NewRequest(http.MethodPost, url, payload)
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if !IsOK(resp.StatusCode) {
		reqDump, _ := httputil.DumpRequest(req, true)
		return fmt.Errorf("failed to add role to user: " + string(reqDump) + "\n" + resp.Status)
	}

	return nil
}

func GetRealmRoleIdByRoleName(roleName string, secret *coreV1.Secret) (string, error) {
	token, err := GetTokenAsAdmin(secret)
	if err != nil {
		return "", err
	}

	params := map[string]string{
		"roleName": roleName,
	}
	url := SetServiceDomainURI(KEYCLOAK_ADMIN_SERVICE_GET_REALM_ROLE_BY_NAME, params)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Add("Authorization", token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	reqDump, _ := httputil.DumpRequest(req, true)
	if !IsOK(resp.StatusCode) {
		return "", fmt.Errorf("failed to get role: " + string(reqDump) + "\n" + resp.Status)
	}

	respJson := &RoleConfig{}
	err = json.NewDecoder(resp.Body).Decode(respJson)
	if err != nil {
		return "", err
	}
	if respJson == nil {
		return "", HyperAuthError{NotFound: true, Type: RESOURCE_TYPE_REALM_ROLE, Name: roleName}
	}

	return respJson.Id, nil
}

func AddRealmLevelRolesToUserRoleMapping(roleName string, userEmail string, secret *coreV1.Secret) error {
	token, err := GetTokenAsAdmin(secret)
	if err != nil {
		return err
	}

	userId, err := GetUserIdByEmail(userEmail, secret)
	if err != nil {
		return err
	}

	roleId, err := GetRealmRoleIdByRoleName(roleName, secret)
	if err != nil {
		return err
	}
	data := []RoleConfig{
		{
			Id:   roleId,
			Name: roleName,
		},
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	payload := bytes.NewBuffer(jsonData)

	params := map[string]string{
		"userId": userId,
	}
	url := SetServiceDomainURI(KEYCLOAK_ADMIN_SERVICE_ADD_REALM_ROLE_TO_USER, params)
	req, err := http.NewRequest(http.MethodPost, url, payload)
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if !IsOK(resp.StatusCode) {
		reqDump, _ := httputil.DumpRequest(req, true)
		return fmt.Errorf("failed to add role to user: " + string(reqDump) + "\n" + resp.Status)
	}

	return nil
}

func GetClientScopesIdByName(name string, secret *coreV1.Secret) (string, error) {
	token, err := GetTokenAsAdmin(secret)
	if err != nil {
		return "", err
	}

	url := SetServiceDomainURI(KEYCLOAK_ADMIN_SERVICE_GET_CLIENT_SCOPES, nil)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Add("Authorization", token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if !IsOK(resp.StatusCode) {
		reqDump, _ := httputil.DumpRequest(req, true)
		return "", fmt.Errorf("failed to get client scope id: " + string(reqDump) + "\n" + resp.Status)
	}

	respJson := &[]ClientScopeConfig{}
	err = json.NewDecoder(resp.Body).Decode(respJson)
	if err != nil {
		return "", err
	}

	for _, data := range *respJson {
		if data.Name == name {
			return data.Id, nil
		}
	}

	return "", HyperAuthError{NotFound: true, Type: RESOURCE_TYPE_CLIENT_SCOPE, Name: name}
}

func AddClientScopeToClient(config ClientScopeMappingConfig, secret *coreV1.Secret) error {
	token, err := GetTokenAsAdmin(secret)
	if err != nil {
		return err
	}

	id, err := GetIdByClientId(config.ClientId, secret)
	if err != nil {
		return err
	}

	clientScopeId, err := GetClientScopesIdByName(config.ClientScope.Name, secret)
	if err != nil {
		return err
	}

	params := map[string]string{
		"id":            id,
		"clientScopeId": clientScopeId,
	}
	url := SetServiceDomainURI(KEYCLOAK_ADMIN_SERVICE_ADD_DEFAULT_CLIENT_SCOPE_TO_CLIENT, params)
	req, err := http.NewRequest(http.MethodPut, url, nil)
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if !IsOK(resp.StatusCode) {
		reqDump, _ := httputil.DumpRequest(req, true)
		return fmt.Errorf("failed to add client scope to client: " + string(reqDump) + "\n" + resp.Status)
	}

	return nil
}

func CreateGroup(config GroupConfig, secret *coreV1.Secret) error {
	token, err := GetTokenAsAdmin(secret)
	if err != nil {
		return err
	}

	jsonData, err := json.Marshal(config)
	if err != nil {
		return err
	}
	payload := bytes.NewBuffer(jsonData)

	url := SetServiceDomainURI(KEYCLOAK_ADMIN_SERVICE_CREATE_GROUP, nil)
	req, err := http.NewRequest(http.MethodPost, url, payload)
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if !IsOK(resp.StatusCode) {
		reqDump, _ := httputil.DumpRequest(req, true)
		return fmt.Errorf("failed to create group: " + string(reqDump) + "\n" + resp.Status)
	}

	return nil
}

func GetGroupIdByName(name string, secret *coreV1.Secret) (string, error) {
	token, err := GetTokenAsAdmin(secret)
	if err != nil {
		return "", err
	}

	url := SetServiceDomainURI(KEYCLOAK_ADMIN_SERVICE_GET_GROUP, nil)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Add("Authorization", token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	reqDump, _ := httputil.DumpRequest(req, true)
	if !IsOK(resp.StatusCode) {
		return "", fmt.Errorf("failed to get group: " + string(reqDump) + "\n" + resp.Status)
	}

	respJson := &[]GroupConfig{}
	err = json.NewDecoder(resp.Body).Decode(respJson)
	if err != nil {
		return "", err
	}

	for _, data := range *respJson {
		if data.Name == name {
			return data.Id, nil
		}
	}

	return "", HyperAuthError{NotFound: true, Type: RESOURCE_TYPE_GROUP, Name: name}
}

func AddGroupToUser(userEmail string, config GroupConfig, secret *coreV1.Secret) error {
	token, err := GetTokenAsAdmin(secret)
	if err != nil {
		return err
	}

	userId, err := GetUserIdByEmail(userEmail, secret)
	if err != nil {
		return err
	}

	groupId, err := GetGroupIdByName(config.Name, secret)
	if err != nil {
		return err
	}

	params := map[string]string{
		"userId":  userId,
		"groupId": groupId,
	}
	url := SetServiceDomainURI(KEYCLOAK_ADMIN_SERVICE_ADD_GROUP_TO_USER, params)
	req, err := http.NewRequest(http.MethodPut, url, nil)
	if err != nil {
		return err
	}
	req.Header.Add("Authorization", token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if !IsOK(resp.StatusCode) {
		reqDump, _ := httputil.DumpRequest(req, true)
		return fmt.Errorf("failed to add group to user: " + string(reqDump) + "\n" + resp.Status)
	}

	return nil
}

func DeleteClient(config ClientConfig, secret *coreV1.Secret) error {
	token, err := GetTokenAsAdmin(secret)
	if err != nil {
		return err
	}

	id, err := GetIdByClientId(config.ClientId, secret)
	if IsNotFound(err) {
		return nil
	} else if err != nil {
		return err
	}

	params := map[string]string{
		"id": id,
	}
	url := SetServiceDomainURI(KEYCLOAK_ADMIN_SERVICE_DELETE_CLIENT, params)
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return err
	}
	req.Header.Add("Authorization", token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if !IsOK(resp.StatusCode) {
		reqDump, _ := httputil.DumpRequest(req, true)
		return fmt.Errorf("failed to delete client: " + string(reqDump) + "\n" + resp.Status)
	}

	return nil
}

func DeleteGroup(config GroupConfig, secret *coreV1.Secret) error {
	token, err := GetTokenAsAdmin(secret)
	if err != nil {
		return err
	}

	groupId, err := GetGroupIdByName(config.Name, secret)
	if IsNotFound(err) {
		return nil
	} else if err != nil {
		return err
	}

	params := map[string]string{
		"groupId": groupId,
	}
	url := SetServiceDomainURI(KEYCLOAK_ADMIN_SERVICE_DELETE_GROUP, params)
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return err
	}
	req.Header.Add("Authorization", token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if !IsOK(resp.StatusCode) {
		reqDump, _ := httputil.DumpRequest(req, true)
		return fmt.Errorf("failed to delete client: " + string(reqDump) + "\n" + resp.Status)
	}

	return nil
}
