package validation

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// UniqueStringListValidator validates that all strings in a list are unique
type UniqueStringListValidator struct{}

// Description returns a description of the validator
func (v UniqueStringListValidator) Description(_ context.Context) string {
	return "Ensures all string values in the list are unique"
}

// MarkdownDescription returns a markdown description of the validator
func (v UniqueStringListValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

// ValidateList performs the validation
func (v UniqueStringListValidator) ValidateList(ctx context.Context, req validator.ListRequest, resp *validator.ListResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	elements := req.ConfigValue.Elements()
	if len(elements) <= 1 {
		// No duplicates possible with 0 or 1 elements
		return
	}

	seen := make(map[string]bool)
	duplicates := make(map[string]bool)

	for i, element := range elements {
		strValue, ok := element.(types.String)
		if !ok {
			resp.Diagnostics.AddAttributeError(
				req.Path.AtListIndex(i),
				"Invalid Element Type",
				"Expected string element in list",
			)
			continue
		}

		if strValue.IsNull() || strValue.IsUnknown() {
			continue
		}

		value := strValue.ValueString()
		if seen[value] {
			duplicates[value] = true
		} else {
			seen[value] = true
		}
	}

	if len(duplicates) > 0 {
		var duplicateValues []string
		for duplicate := range duplicates {
			duplicateValues = append(duplicateValues, duplicate)
		}

		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Duplicate Values Found",
			fmt.Sprintf("List contains duplicate values: %v. All values must be unique.", duplicateValues),
		)
	}
}

// UniqueStringList creates a new unique string list validator
func UniqueStringList() validator.List {
	return UniqueStringListValidator{}
}
