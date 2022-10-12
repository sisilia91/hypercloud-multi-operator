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

const (
	// admin api
	KEYCLOAK_ADMIN_SERVICE_GET_TOKEN                          = "/auth/realms/master/protocol/openid-connect/token"
	KEYCLOAK_ADMIN_SERVICE_GET_CLIENTS                        = "/auth/admin/realms/tmax/clients"
	KEYCLOAK_ADMIN_SERVICE_CREATE_CLIENT                      = "/auth/admin/realms/tmax/clients"
	KEYCLOAK_ADMIN_SERVICE_DELETE_CLIENT                      = "/auth/admin/realms/tmax/clients/@@id@@"
	KEYCLOAK_ADMIN_SERVICE_CREATE_CLIENT_PROTOCOL_MAPPERS     = "/auth/admin/realms/tmax/clients/@@id@@/protocol-mappers/models"
	KEYCLOAK_ADMIN_SERVICE_CREATE_CLIENT_ROLES                = "/auth/admin/realms/tmax/clients/@@id@@/roles"
	KEYCLOAK_ADMIN_SERVICE_GET_CLIENT_ROLE_BY_NAME            = "/auth/admin/realms/tmax/clients/@@id@@/roles/@@roleName@@"
	KEYCLOAK_ADMIN_SERVICE_ADD_CLIENT_ROLE_TO_USER            = "/auth/admin/realms/tmax/users/@@userId@@/role-mappings/clients/@@id@@"
	KEYCLOAK_ADMIN_SERVICE_GET_CLIENT_SCOPES                  = "/auth/admin/realms/tmax/client-scopes"
	KEYCLOAK_ADMIN_SERVICE_ADD_DEFAULT_CLIENT_SCOPE_TO_CLIENT = "/auth/admin/realms/tmax/clients/@@id@@/default-client-scopes/@@clientScopeId@@"
	KEYCLOAK_ADMIN_SERVICE_GET_REALM_ROLE_BY_NAME             = "/auth/admin/realms/tmax/roles/@@roleName@@"
	KEYCLOAK_ADMIN_SERVICE_ADD_REALM_ROLE_TO_USER             = "/auth/admin/realms/tmax/users/@@userId@@/role-mappings/realm"
	KEYCLOAK_ADMIN_SERVICE_GET_GROUP                          = "/auth/admin/realms/tmax/groups"
	KEYCLOAK_ADMIN_SERVICE_CREATE_GROUP                       = "/auth/admin/realms/tmax/groups"
	KEYCLOAK_ADMIN_SERVICE_DELETE_GROUP                       = "/auth/admin/realms/tmax/groups/@@groupId@@"
	KEYCLOAK_ADMIN_SERVICE_ADD_GROUP_TO_USER                  = "/auth/admin/realms/tmax/users/@@userId@@/groups/@@groupId@@"
	KEYCLOAK_ADMIN_SERVICE_GET_USERS_BY_EMAIL                 = "/auth/admin/realms/tmax/users?exact=true&email=@@userEmail@@"
)

const (
	PROTOCOL_MAPPER_CONFIG_PROTOCOL_OPENID_CONNECT        = "openid-connect"
	PROTOCOL_MAPPER_CONFIG_PROTOCOL_NAME_AUDIENCE         = "oidc-audience-mapper"
	PROTOCOL_MAPPER_CONFIG_PROTOCOL_NAME_GROUP_MEMBERSHIP = "oidc-group-membership-mapper"
	PROTOCOL_MAPPER_CONFIG_PROTOCOL_NAME_USER_REALM_ROLE  = "oidc-usermodel-realm-role-mapper"
	PROTOCOL_MAPPER_CONFIG_PROTOCOL_NAME_USER_CLIENT_ROLE = "oidc-usermodel-client-role-mapper"
)

const (
	RESOURCE_TYPE_USER_EMAIL   = "UserEmail"
	RESOURCE_TYPE_CLIENT       = "Client"
	RESOURCE_TYPE_CLIENT_ROLE  = "ClientRole"
	RESOURCE_TYPE_REALM_ROLE   = "RealmRole"
	RESOURCE_TYPE_CLIENT_SCOPE = "ClientScop"
	RESOURCE_TYPE_GROUP        = "Group"
)

// const (
// 	HYPERAUTH_HTTPS_SECRET = "hyperauth-https-secret"
// 	HYPERAUTH_NAMESPACE    = "hyperauth"
// )
