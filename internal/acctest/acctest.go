// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package acctest

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"testing"

	"github.com/altrsoftware/terraform-provider-altr/internal/provider"
	"github.com/altrsoftware/terraform-provider-altr/internal/version"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-mux/tf6muxserver"
)

// ProtoV6ProviderFactories provides a Provider Factory to be used within
// acceptance tests.
var ProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"altr": func() (tfprotov6.ProviderServer, error) {
		providers := []func() tfprotov6.ProviderServer{
			providerserver.NewProtocol6(provider.New(version.ProviderVersion)()),
		}

		return tf6muxserver.NewMuxServer(context.Background(), providers...)
	},
}

// PreCheck verifies that the required provider testing configuration is set.
//
// This PreCheck function should be present in every acceptance test. It ensures
// credentials and other test environment settings are configured.
func PreCheck(t *testing.T) {
	t.Helper()

	if os.Getenv("ALTR_ORG_ID") == "" {
		t.Fatal("ALTR_ORG_ID must be set for acceptance tests")
	}

	if os.Getenv("ALTR_API_KEY") == "" {
		t.Fatal("ALTR_API_KEY must be set for acceptance tests")
	}

	if os.Getenv("ALTR_SECRET") == "" {
		t.Fatal("ALTR_SECRET must be set for acceptance tests")
	}
}

// Helper function to get environment variables with defaults
func TestGetEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return defaultValue
}

// RandomWithPrefixUnderScore is used to generate a unique name with a prefix,
func RandomWithPrefixUnderscore(prefix string) string {
	return fmt.Sprintf("%s_%d", prefix, RandInt())
}

// RandomWithPrefixUnderScoreMaxLength generates a random string with a prefix and ensures it does not exceed the max length.
func RandomWithPrefixUnderscoreMaxLength(prefix string, maxLength int) string {
	if len(prefix) >= maxLength {
		return prefix[:maxLength]
	}

	randomSuffix := fmt.Sprintf("_%d", RandInt())
	if len(prefix)+len(randomSuffix) > maxLength {
		randomSuffix = randomSuffix[:maxLength-len(prefix)]
	}

	return fmt.Sprintf("%s%s", prefix, randomSuffix)
}

// RandInt generates a random integer
func RandInt() int {
	return rand.Int()
}
