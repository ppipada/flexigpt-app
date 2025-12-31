package store

import (
	"errors"
	"fmt"
	"strings"

	"github.com/ppipada/flexigpt-app/pkg/bundleitemutils"
	"github.com/ppipada/flexigpt-app/pkg/modelpreset/spec"
	inferencegoSpec "github.com/ppipada/inference-go/spec"
)

// validateProviderPreset performs structural and referential checks for a
// provider together with its embedded model presets.
func validateProviderPreset(pp *spec.ProviderPreset) error {
	if pp == nil {
		return spec.ErrNilProvider
	}
	if pp.SchemaVersion != spec.SchemaVersion {
		return fmt.Errorf("provider %q: schemaVersion %q not equal to %q",
			pp.Name, pp.SchemaVersion, spec.SchemaVersion)
	}
	if err := validateProviderName(pp.Name); err != nil {
		return fmt.Errorf("provider %q: %w", pp.Name, err)
	}
	if strings.TrimSpace(string(pp.DisplayName)) == "" {
		return fmt.Errorf("provider %q: displayName is empty", pp.Name)
	}
	if pp.CreatedAt.IsZero() || pp.ModifiedAt.IsZero() {
		return fmt.Errorf("provider %q: %w", pp.Name, spec.ErrInvalidTimestamp)
	}
	if strings.TrimSpace(pp.Origin) == "" {
		return fmt.Errorf("provider %q: origin is empty", pp.Name)
	}
	if strings.TrimSpace(pp.ChatCompletionPathPrefix) == "" {
		return fmt.Errorf("provider %q: chatCompletionPathPrefix is empty", pp.Name)
	}
	// Per-model validation and duplicate ID detection.
	seenModel := map[spec.ModelPresetID]string{}
	for mid, mp := range pp.ModelPresets {
		if err := validateModelPreset(&mp); err != nil {
			return fmt.Errorf("provider %q, model %q: %w", pp.Name, mid, err)
		}
		if prev := seenModel[mid]; prev != "" {
			return fmt.Errorf("provider %q: duplicate modelPresetID %q (also in %s)",
				pp.Name, mid, prev)
		}
		seenModel[mid] = string(mid)
	}

	// DefaultModelPresetID must exist if set.
	if pp.DefaultModelPresetID != "" {
		if _, ok := pp.ModelPresets[pp.DefaultModelPresetID]; !ok {
			return fmt.Errorf("provider %q: defaultModelPresetID %q not present",
				pp.Name, pp.DefaultModelPresetID)
		}
	}
	return nil
}

// validateModelPreset performs structural validation for a single model preset.
func validateModelPreset(mp *spec.ModelPreset) error {
	if mp == nil {
		return spec.ErrNilModelPreset
	}
	if mp.SchemaVersion != spec.SchemaVersion {
		return fmt.Errorf("schemaVersion %q not equal to %q",
			mp.SchemaVersion, spec.SchemaVersion)
	}
	if err := validateModelPresetID(mp.ID); err != nil {
		return err
	}
	if err := validateModelName(mp.Name); err != nil {
		return err
	}
	if err := validateModelSlug(mp.Slug); err != nil {
		return err
	}
	if strings.TrimSpace(string(mp.DisplayName)) == "" {
		return errors.New("displayName is empty")
	}
	if mp.CreatedAt.IsZero() || mp.ModifiedAt.IsZero() {
		return spec.ErrInvalidTimestamp
	}

	// Either Reasoning or Temperature must be provided (both cannot be nil).
	if mp.Reasoning == nil && mp.Temperature == nil {
		return errors.New("either reasoning or temperature must be set")
	}

	// Reasoning checks (optional).
	if mp.Reasoning != nil {
		if err := validateReasoning(mp.Reasoning); err != nil {
			return fmt.Errorf("invalid reasoning: %w", err)
		}
	}
	return nil
}

// validateProviderName currently only trims blanks; extend as required.
func validateProviderName(n spec.ProviderName) error {
	if strings.TrimSpace(string(n)) == "" {
		return errors.New("name is empty")
	}
	return nil
}

// validateModelName is similar free-form stub.
func validateModelName(n spec.ModelName) error {
	if strings.TrimSpace(string(n)) == "" {
		return errors.New("name is empty")
	}
	return nil
}

// validateModelSlug uses existing tag validator.
func validateModelSlug(s spec.ModelSlug) error {
	return bundleitemutils.ValidateTag(string(s))
}

// validateModelPresetID uses the same rule set as slugs for now.
func validateModelPresetID(id spec.ModelPresetID) error {
	return bundleitemutils.ValidateTag(string(id))
}

// validateReasoning verifies the type/level/tokens combos.
func validateReasoning(r *inferencegoSpec.ReasoningParam) error {
	switch r.Type {
	case inferencegoSpec.ReasoningTypeHybridWithTokens:
		if r.Tokens <= 0 {
			return errors.New("tokens must be >0 for hybridWithTokens")
		}
	case inferencegoSpec.ReasoningTypeSingleWithLevels:
		switch r.Level {
		case
			inferencegoSpec.ReasoningLevelNone,
			inferencegoSpec.ReasoningLevelMinimal,
			inferencegoSpec.ReasoningLevelLow,
			inferencegoSpec.ReasoningLevelMedium,
			inferencegoSpec.ReasoningLevelHigh,
			inferencegoSpec.ReasoningLevelXHigh:
			// Valid.
		default:
			return fmt.Errorf("invalid level %q for singleWithLevels", r.Level)
		}
	default:
		return fmt.Errorf("unknown type %q", r.Type)
	}
	return nil
}
