// Copyright (c) ALTR Solutions, Inc.
// SPDX-License-Identifier: Apache-2.0

package client

import "encoding/json"

// Sidecar structures
type Sidecar struct {
	ID                       string     `json:"id"`
	Name                     string     `json:"name"`
	Description              string     `json:"description"`
	Hostname                 string     `json:"hostname"`
	OrgID                    string     `json:"org_id"`
	DataPlaneURL             string     `json:"data_plane_url"`
	ListenerRepoBindingCount int        `json:"listener_repo_binding_count"`
	ListenerCount            int        `json:"listener_count"`
	PublicKey1               *PublicKey `json:"public_key_1,omitempty"`
	PublicKey2               *PublicKey `json:"public_key_2,omitempty"`
	UnsupportedQueryBypass   bool       `json:"unsupported_query_bypass"`
	CreatedAt                string     `json:"created_at"`
	UpdatedAt                string     `json:"updated_at"`
}

type PublicKey struct {
	RSAKey       string `json:"rsa_key"`
	RegisteredAt string `json:"registered_at"`
}

type CreateSidecarInput struct {
	Name                   string `json:"name"`
	Description            string `json:"description"`
	Hostname               string `json:"hostname"`
	PublicKey1             string `json:"public_key_1,omitempty"`
	PublicKey2             string `json:"public_key_2,omitempty"`
	UnsupportedQueryBypass bool   `json:"unsupported_query_bypass"`
}

type UpdateSidecarInput struct {
	Name                   *string `json:"name,omitempty"`
	Description            *string `json:"description,omitempty"`
	Hostname               *string `json:"hostname,omitempty"`
	PublicKey1             *string `json:"public_key_1,omitempty"`
	PublicKey2             *string `json:"public_key_2,omitempty"`
	UnsupportedQueryBypass *bool   `json:"unsupported_query_bypass,omitempty"`
}

// Repo structures
type Repo struct {
	Name         string `json:"name"`
	Description  string `json:"description"`
	Hostname     string `json:"hostname"`
	Port         int    `json:"port"`
	Type         string `json:"type"`
	UserCount    int    `json:"user_count"`
	BindingCount int    `json:"binding_count"`
	OrgID        string `json:"org_id"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

type CreateRepoInput struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type"`
	Hostname    string `json:"hostname"`
	Port        int    `json:"port"`
}

type UpdateRepoInput struct {
	Description string `json:"description"`
}

type RepoUser struct {
	Username          string             `json:"username"`
	RepoName          string             `json:"repo_name"`
	AWSSecretsManager *AWSSecretsManager `json:"aws_secrets_manager"`
	AzureKeyVault     *AzureKeyVault     `json:"azure_key_vault"`
	CreatedAt         string             `json:"created_at"`
	UpdatedAt         string             `json:"updated_at"`
}

type UpdateRepoUserInput struct {
	AWSSecretsManager *AWSSecretsManager `json:"aws_secrets_manager,omitempty"`
	AzureKeyVault     *AzureKeyVault     `json:"azure_key_vault,omitempty"`
}

type AWSSecretsManager struct {
	IAMRole     string `json:"iam_role"`
	SecretsPath string `json:"secrets_path"`
}

type AzureKeyVault struct {
	KeyVaultURI string `json:"key_vault_uri"`
	SecretName  string `json:"secret_name"`
}

type CreateRepoUserInput struct {
	Username          string             `json:"username"`
	AWSSecretsManager *AWSSecretsManager `json:"aws_secrets_manager,omitempty"`
	AzureKeyVault     *AzureKeyVault     `json:"azure_key_vault,omitempty"`
}

// Sidecar Listener structures
type ListenerPort struct {
	Port              int    `json:"port"`
	DatabaseType      string `json:"database_type"`
	AdvertisedVersion string `json:"advertised_version"`
}

type RegisterSidecarListenerInput struct {
	Port              int    `json:"port"`
	DatabaseType      string `json:"database_type"`
	AdvertisedVersion string `json:"advertised_version,omitempty"`
}

type ListSidecarListenersOutput struct {
	SidecarListeners []ListenerPort `json:"sidecar_listeners"`
	ContiguousID     string         `json:"contiguous_id"`
}

// Repo Sidecar Binding structures
type RepoSidecarBinding struct {
	Port      int    `json:"port"`
	SidecarID string `json:"sidecar_id"`
	RepoName  string `json:"repo_name"`
}

type GetRepoBindOutput struct {
	RepoSidecarBinding RepoSidecarBinding `json:"repo_sidecar_binding"`
}

type ListBindingsOutput struct {
	RepoBindings []RepoSidecarBinding `json:"repo_bindings"`
	ContiguousID string               `json:"contiguous_id"`
}

type APIResponse struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data,omitempty"`
}

// Agent structures
type Agent struct {
	ID           string     `json:"id"`
	Type         string     `json:"type"`
	Name         string     `json:"name"`
	Description  string     `json:"description"`
	DataPlaneURL string     `json:"data_plane_url"`
	PublicKey1   *PublicKey `json:"public_key_1,omitempty"`
	PublicKey2   *PublicKey `json:"public_key_2,omitempty"`
	TaskCount    int        `json:"task_count"`
	CreatedAt    string     `json:"created_at"`
	UpdatedAt    string     `json:"updated_at"`
}

type CreateAgentInput struct {
	Type        string  `json:"type"`
	Name        string  `json:"name"`
	Description string  `json:"description,omitempty"`
	PublicKey1  *string `json:"public_key_1,omitempty"`
	PublicKey2  *string `json:"public_key_2,omitempty"`
}

type UpdateAgentInput struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	PublicKey1  *string `json:"public_key_1,omitempty"`
	PublicKey2  *string `json:"public_key_2,omitempty"`
}

// AgentTask structures
type SslConfig struct {
	Enabled                bool   `json:"enabled"`
	HostnameInCertificate  string `json:"hostname_in_certificate,omitempty"`
	TrustServerCertificate bool   `json:"trust_server_certificate"`
	TrustStorePasswordARN  string `json:"trust_store_password_arn,omitempty"`
	TrustStorePath         string `json:"trust_store_path,omitempty"`
}

type AgentTaskConfiguration struct {
	CollectionName     string     `json:"collection_name,omitempty"`
	ClassificationType *int       `json:"classification_type,omitempty"`
	SampleStrategy     string     `json:"sample_strategy,omitempty"`
	SslConfig          *SslConfig `json:"ssl_config,omitempty"`
}

type AgentTaskSchedule struct {
	Type        string `json:"type"`
	Value       string `json:"value"`
	MaxDuration string `json:"max_duration,omitempty"`
	Timezone    string `json:"timezone,omitempty"`
}

type AgentTask struct {
	ID            string                 `json:"id"`
	AgentID       string                 `json:"agent_id"`
	Name          string                 `json:"name"`
	Description   string                 `json:"description"`
	RepoName      string                 `json:"repo_name"`
	ServiceUser   string                 `json:"service_user,omitempty"`
	Configuration AgentTaskConfiguration `json:"configuration"`
	Schedule      AgentTaskSchedule      `json:"schedule"`
	CreatedAt     string                 `json:"created_at"`
	UpdatedAt     string                 `json:"updated_at"`
}

type CreateAgentTaskInput struct {
	Name          string                 `json:"name"`
	Description   string                 `json:"description,omitempty"`
	RepoName      string                 `json:"repo_name"`
	ServiceUser   string                 `json:"service_user,omitempty"`
	Configuration AgentTaskConfiguration `json:"configuration"`
	Schedule      AgentTaskSchedule      `json:"schedule"`
}

type UpdateAgentTaskInput struct {
	Name          *string                 `json:"name,omitempty"`
	Description   *string                 `json:"description,omitempty"`
	Configuration *AgentTaskConfiguration `json:"configuration,omitempty"`
	Schedule      *AgentTaskSchedule      `json:"schedule,omitempty"`
}

// ServiceUser structures
type EnvironmentVariable struct {
	VariableName string `json:"variable_name"`
}

type SecretFile struct {
	Path string `json:"path"`
}

type ServiceUser struct {
	Username            string               `json:"username"`
	RepoName            string               `json:"repo_name"`
	AWSSecretsManager   *AWSSecretsManager   `json:"aws_secrets_manager,omitempty"`
	AzureKeyVault       *AzureKeyVault       `json:"azure_key_vault,omitempty"`
	EnvironmentVariable *EnvironmentVariable `json:"environment_variable,omitempty"`
	SecretFile          *SecretFile          `json:"secret_file,omitempty"`
	TaskCount           int                  `json:"task_count"`
	CreatedAt           string               `json:"created_at"`
	UpdatedAt           string               `json:"updated_at"`
}

type CreateServiceUserInput struct {
	Username            string               `json:"username"`
	AWSSecretsManager   *AWSSecretsManager   `json:"aws_secrets_manager,omitempty"`
	AzureKeyVault       *AzureKeyVault       `json:"azure_key_vault,omitempty"`
	EnvironmentVariable *EnvironmentVariable `json:"environment_variable,omitempty"`
	SecretFile          *SecretFile          `json:"secret_file,omitempty"`
}

type UpdateServiceUserInput struct {
	AWSSecretsManager   *AWSSecretsManager   `json:"aws_secrets_manager,omitempty"`
	AzureKeyVault       *AzureKeyVault       `json:"azure_key_vault,omitempty"`
	EnvironmentVariable *EnvironmentVariable `json:"environment_variable,omitempty"`
	SecretFile          *SecretFile          `json:"secret_file,omitempty"`
}
