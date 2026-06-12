// Copyright (c) ALTR Solutions, Inc.
// SPDX-License-Identifier: Apache-2.0

package repo

import (
	"errors"

	"github.com/altrsoftware/terraform-provider-altr/internal/client"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// Credential providers are shared by repo users and service users: both must
// configure exactly one of aws_secrets_manager, azure_key_vault,
// environment_variable, or secret_file. The attribute types, schema, and
// model<->client conversions live here so the two resources stay in sync.

var awsAttrTypes = map[string]attr.Type{
	"iam_role":     types.StringType,
	"secrets_path": types.StringType,
}

var azureAttrTypes = map[string]attr.Type{
	"key_vault_uri": types.StringType,
	"secret_name":   types.StringType,
}

var envVarAttrTypes = map[string]attr.Type{
	"variable_name": types.StringType,
}

var secretFileAttrTypes = map[string]attr.Type{
	"path": types.StringType,
}

// credentialProviderSchemaAttributes returns the four mutually-exclusive
// credential provider attributes shared by repo users and service users.
func credentialProviderSchemaAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"aws_secrets_manager": schema.SingleNestedAttribute{
			Description: "AWS Secrets Manager credential provider.",
			Optional:    true,
			Attributes: map[string]schema.Attribute{
				"secrets_path": schema.StringAttribute{
					Description: "Path or name of the secret in AWS Secrets Manager.",
					Required:    true,
					Validators: []validator.String{
						stringvalidator.LengthAtLeast(1),
					},
				},
				"iam_role": schema.StringAttribute{
					Description: "ARN of an IAM role to assume when retrieving the secret.",
					Optional:    true,
					Computed:    true,
				},
			},
		},
		"azure_key_vault": schema.SingleNestedAttribute{
			Description: "Azure Key Vault credential provider.",
			Optional:    true,
			Attributes: map[string]schema.Attribute{
				"key_vault_uri": schema.StringAttribute{
					Description: "HTTPS URL of the Azure Key Vault.",
					Required:    true,
					Validators: []validator.String{
						stringvalidator.LengthAtLeast(1),
					},
				},
				"secret_name": schema.StringAttribute{
					Description: "Name of the secret within the vault.",
					Required:    true,
					Validators: []validator.String{
						stringvalidator.LengthAtLeast(1),
					},
				},
			},
		},
		"environment_variable": schema.SingleNestedAttribute{
			Description: "Environment variable credential provider.",
			Optional:    true,
			Attributes: map[string]schema.Attribute{
				"variable_name": schema.StringAttribute{
					Description: "Name of the OS environment variable containing the secret.",
					Required:    true,
					Validators: []validator.String{
						stringvalidator.LengthAtLeast(1),
					},
				},
			},
		},
		"secret_file": schema.SingleNestedAttribute{
			Description: "Secret file credential provider. Reads from /altr/secrets/<path> at runtime.",
			Optional:    true,
			Attributes: map[string]schema.Attribute{
				"path": schema.StringAttribute{
					Description: "Simple filename (no path separators). Resolved under /altr/secrets/ at runtime.",
					Required:    true,
					Validators: []validator.String{
						stringvalidator.LengthAtLeast(1),
					},
				},
			},
		},
	}
}

// validateSingleCredentialProvider ensures exactly one of the four credential
// providers is configured.
func validateSingleCredentialProvider(aws, azure, envVar, secretFile basetypes.ObjectValue) error {
	count := 0

	if !aws.IsNull() {
		count++
	}

	if !azure.IsNull() {
		count++
	}

	if !envVar.IsNull() {
		count++
	}

	if !secretFile.IsNull() {
		count++
	}

	if count == 0 {
		return errors.New("exactly one credential provider must be specified (aws_secrets_manager, azure_key_vault, environment_variable, or secret_file)")
	}

	if count > 1 {
		return errors.New("only one credential provider can be specified at a time")
	}

	return nil
}

// credentialProvidersFromObjects extracts the active credential provider from
// the four nested objects, returning the client-side representations.
func credentialProvidersFromObjects(aws, azure, envVar, secretFile basetypes.ObjectValue) (
	*client.AWSSecretsManager,
	*client.AzureKeyVault,
	*client.EnvironmentVariable,
	*client.SecretFile,
) {
	var (
		awsOut    *client.AWSSecretsManager
		azureOut  *client.AzureKeyVault
		envVarOut *client.EnvironmentVariable
		fileOut   *client.SecretFile
	)

	if !aws.IsNull() {
		secretsPath := aws.Attributes()["secrets_path"].(types.String)
		awsOut = &client.AWSSecretsManager{SecretsPath: secretsPath.ValueString()}

		iamRole := aws.Attributes()["iam_role"].(types.String)
		if !iamRole.IsNull() && iamRole.ValueString() != "" {
			awsOut.IAMRole = iamRole.ValueString()
		}
	}

	if !azure.IsNull() {
		azureOut = &client.AzureKeyVault{
			KeyVaultURI: azure.Attributes()["key_vault_uri"].(types.String).ValueString(),
			SecretName:  azure.Attributes()["secret_name"].(types.String).ValueString(),
		}
	}

	if !envVar.IsNull() {
		envVarOut = &client.EnvironmentVariable{
			VariableName: envVar.Attributes()["variable_name"].(types.String).ValueString(),
		}
	}

	if !secretFile.IsNull() {
		fileOut = &client.SecretFile{
			Path: secretFile.Attributes()["path"].(types.String).ValueString(),
		}
	}

	return awsOut, azureOut, envVarOut, fileOut
}

// credentialProvidersToObjects converts the client-side credential providers
// back into the four nested object values for state, nulling out any that are
// absent from the API response.
func credentialProvidersToObjects(
	aws *client.AWSSecretsManager,
	azure *client.AzureKeyVault,
	envVar *client.EnvironmentVariable,
	secretFile *client.SecretFile,
) (basetypes.ObjectValue, basetypes.ObjectValue, basetypes.ObjectValue, basetypes.ObjectValue) {
	awsObj := basetypes.NewObjectNull(awsAttrTypes)
	if aws != nil && (aws.IAMRole != "" || aws.SecretsPath != "") {
		awsObj = basetypes.NewObjectValueMust(awsAttrTypes, map[string]attr.Value{
			"iam_role":     types.StringValue(aws.IAMRole),
			"secrets_path": types.StringValue(aws.SecretsPath),
		})
	}

	azureObj := basetypes.NewObjectNull(azureAttrTypes)
	if azure != nil && (azure.KeyVaultURI != "" || azure.SecretName != "") {
		azureObj = basetypes.NewObjectValueMust(azureAttrTypes, map[string]attr.Value{
			"key_vault_uri": types.StringValue(azure.KeyVaultURI),
			"secret_name":   types.StringValue(azure.SecretName),
		})
	}

	envVarObj := basetypes.NewObjectNull(envVarAttrTypes)
	if envVar != nil && envVar.VariableName != "" {
		envVarObj = basetypes.NewObjectValueMust(envVarAttrTypes, map[string]attr.Value{
			"variable_name": types.StringValue(envVar.VariableName),
		})
	}

	secretFileObj := basetypes.NewObjectNull(secretFileAttrTypes)
	if secretFile != nil && secretFile.Path != "" {
		secretFileObj = basetypes.NewObjectValueMust(secretFileAttrTypes, map[string]attr.Value{
			"path": types.StringValue(secretFile.Path),
		})
	}

	return awsObj, azureObj, envVarObj, secretFileObj
}
