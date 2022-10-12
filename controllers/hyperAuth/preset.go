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
	"os"
	"strings"

	"github.com/tmax-cloud/hypercloud-multi-operator/controllers/util"
)

func GetClientConfigPreset(prefix string) []ClientConfig {
	AUTH_CLIENT_SECRET := util.AUTH_CLIENT_SECRET
	configs := []ClientConfig{
		{
			ClientId:                  strings.Join([]string{prefix, "kibana"}, "-"),
			Secret:                    os.Getenv(AUTH_CLIENT_SECRET),
			DirectAccessGrantsEnabled: true,
			ImplicitFlowEnabled:       false,
			RedirectUris:              []string{"*"},
		},
		{
			ClientId:                  strings.Join([]string{prefix, "grafana"}, "-"),
			Secret:                    os.Getenv(AUTH_CLIENT_SECRET),
			DirectAccessGrantsEnabled: true,
			ImplicitFlowEnabled:       false,
			RedirectUris:              []string{"*"},
		},
		{
			ClientId:                  strings.Join([]string{prefix, "kiali"}, "-"),
			Secret:                    os.Getenv(AUTH_CLIENT_SECRET),
			DirectAccessGrantsEnabled: true,
			ImplicitFlowEnabled:       true,
			RedirectUris:              []string{"*"},
		},
		{
			ClientId:                  strings.Join([]string{prefix, "jaeger"}, "-"),
			Secret:                    os.Getenv(AUTH_CLIENT_SECRET),
			DirectAccessGrantsEnabled: true,
			ImplicitFlowEnabled:       false,
			RedirectUris:              []string{"*"},
		},
		{
			ClientId:                  strings.Join([]string{prefix, "hyperregistry"}, "-"),
			Secret:                    os.Getenv(AUTH_CLIENT_SECRET),
			DirectAccessGrantsEnabled: true,
			ImplicitFlowEnabled:       false,
			RedirectUris:              []string{"*"},
		},
		{
			ClientId:                  strings.Join([]string{prefix, "opensearch"}, "-"),
			Secret:                    os.Getenv(AUTH_CLIENT_SECRET),
			DirectAccessGrantsEnabled: true,
			ImplicitFlowEnabled:       false,
			RedirectUris:              []string{"*"},
		},
	}

	return configs
}

func GetMappingProtocolMapperToClientConfigPreset(prefix string) []ClientLevelProtocolMapperConfig {
	configs := []ClientLevelProtocolMapperConfig{
		{
			ClientId: strings.Join([]string{prefix, "kibana"}, "-"),
			ProtocolMapper: ProtocolMapperConfig{
				Name:           "kibana",
				Protocol:       PROTOCOL_MAPPER_CONFIG_PROTOCOL_OPENID_CONNECT,
				ProtocolMapper: PROTOCOL_MAPPER_CONFIG_PROTOCOL_NAME_AUDIENCE,
				Config: MapperConfig{
					IncludedClientAudience: strings.Join([]string{prefix, "kibana"}, "-"),
					IdTokenClaim:           false,
					AccessTokenClaim:       true,
				},
			},
		},
		{
			ClientId: strings.Join([]string{prefix, "jaeger"}, "-"),
			ProtocolMapper: ProtocolMapperConfig{
				Name:           "jaeger",
				Protocol:       PROTOCOL_MAPPER_CONFIG_PROTOCOL_OPENID_CONNECT,
				ProtocolMapper: PROTOCOL_MAPPER_CONFIG_PROTOCOL_NAME_AUDIENCE,
				Config: MapperConfig{
					IncludedClientAudience: strings.Join([]string{prefix, "jaeger"}, "-"),
					IdTokenClaim:           false,
					AccessTokenClaim:       true,
				},
			},
		},
		{
			ClientId: strings.Join([]string{prefix, "hyperregistry"}, "-"),
			ProtocolMapper: ProtocolMapperConfig{
				Name:           "group",
				Protocol:       PROTOCOL_MAPPER_CONFIG_PROTOCOL_OPENID_CONNECT,
				ProtocolMapper: PROTOCOL_MAPPER_CONFIG_PROTOCOL_NAME_GROUP_MEMBERSHIP,
				Config: MapperConfig{
					ClaimName:          "group",
					FullPath:           true,
					IdTokenClaim:       true,
					AccessTokenClaim:   true,
					UserInfoTokenClaim: true,
				},
			},
		},
		{
			ClientId: strings.Join([]string{prefix, "opensearch"}, "-"),
			ProtocolMapper: ProtocolMapperConfig{
				Name:           "client roles",
				Protocol:       PROTOCOL_MAPPER_CONFIG_PROTOCOL_OPENID_CONNECT,
				ProtocolMapper: PROTOCOL_MAPPER_CONFIG_PROTOCOL_NAME_USER_CLIENT_ROLE,
				Config: MapperConfig{
					Multivalued:        true,
					ClaimName:          "roles",
					JsonType:           "String",
					IdTokenClaim:       true,
					AccessTokenClaim:   true,
					UserInfoTokenClaim: true,
				},
			},
		},
	}

	return configs
}

func GetClientLevelRoleConfigPreset(prefix string) []ClientLevelRoleConfig {
	configs := []ClientLevelRoleConfig{
		{
			ClientId: strings.Join([]string{prefix, "kibana"}, "-"),
			Role: RoleConfig{
				Name: "kibana-manager",
			},
		},
		{
			ClientId: strings.Join([]string{prefix, "jaeger"}, "-"),
			Role: RoleConfig{
				Name: "jaeger-manager",
			},
		},
		{
			ClientId: strings.Join([]string{prefix, "opensearch"}, "-"),
			Role: RoleConfig{
				Name: "opensearch-admin",
			},
		},
		{
			ClientId: strings.Join([]string{prefix, "opensearch"}, "-"),
			Role: RoleConfig{
				Name: "opensearch-developer",
			},
		},
		{
			ClientId: strings.Join([]string{prefix, "opensearch"}, "-"),
			Role: RoleConfig{
				Name: "opensearch-guest",
			},
		},
	}

	return configs
}

func GetClientScopeMappingPreset(prefix string) []ClientScopeMappingConfig {
	configs := []ClientScopeMappingConfig{
		{
			ClientId: strings.Join([]string{prefix, "kiali"}, "-"),
			ClientScope: ClientScopeConfig{
				Name: "kubernetes",
			},
		},
	}

	return configs
}

func GetGroupConfigPreset(prefix string) []GroupConfig {
	configs := []GroupConfig{
		{
			Name:      strings.Join([]string{prefix, "hyperregistry"}, "-"),
			Path:      "/" + prefix + "-hyperregistry",
			SubGroups: []string{},
		},
	}

	return configs
}
