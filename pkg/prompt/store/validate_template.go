package store

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/ppipada/flexigpt-app/pkg/bundleitemutils"
	"github.com/ppipada/flexigpt-app/pkg/prompt/spec"
)

// Regular expressions.
var (
	placeholderRE = regexp.MustCompile(`\{\{([a-zA-Z_][a-zA-Z0-9_-]*)\}\}`)
)

// validateTemplate performs a structural and referential integrity check.
func validateTemplate(tpl *spec.PromptTemplate) error {
	if tpl == nil {
		return errors.New("template is nil")
	}
	if tpl.SchemaVersion != spec.SchemaVersion {
		return fmt.Errorf(
			"schemaVersion %q does not match expected %q",
			tpl.SchemaVersion,
			spec.SchemaVersion,
		)
	}
	if err := bundleitemutils.ValidateItemSlug(tpl.Slug); err != nil {
		return fmt.Errorf("invalid slug: %w", err)
	}
	if err := bundleitemutils.ValidateItemVersion(tpl.Version); err != nil {
		return fmt.Errorf("invalid version: %w", err)
	}
	if strings.TrimSpace(tpl.DisplayName) == "" {
		return errors.New("displayName is empty")
	}
	if tpl.CreatedAt.IsZero() {
		return errors.New("createdAt is zero")
	}
	if tpl.ModifiedAt.IsZero() {
		return errors.New("modifiedAt is zero")
	}
	if len(tpl.Blocks) == 0 {
		return errors.New("at least one message block required")
	}

	// Validate message blocks and collect placeholders.
	allowedRoles := map[spec.PromptRoleEnum]struct{}{
		spec.System:    {},
		spec.Developer: {},
		spec.User:      {},
		spec.Assistant: {},
	}
	blockIDs := map[spec.MessageBlockID]struct{}{}
	placeholders := map[string]struct{}{}

	for i, b := range tpl.Blocks {
		if _, ok := allowedRoles[b.Role]; !ok {
			return fmt.Errorf("blocks[%d]: invalid role %q", i, b.Role)
		}
		if strings.TrimSpace(b.Content) == "" {
			return fmt.Errorf("blocks[%d]: content is empty", i)
		}
		if strings.TrimSpace(string(b.ID)) == "" {
			return fmt.Errorf("blocks[%d]: id is empty", i)
		}
		if _, dup := blockIDs[b.ID]; dup {
			return fmt.Errorf("duplicate block id %q", b.ID)
		}
		blockIDs[b.ID] = struct{}{}

		for _, m := range placeholderRE.FindAllStringSubmatch(b.Content, -1) {
			if len(m) == 2 {
				placeholders[m[1]] = struct{}{}
			}
		}
	}

	// Validate variables.
	allowedVarTypes := map[spec.VarType]struct{}{
		spec.VarString:  {},
		spec.VarNumber:  {},
		spec.VarBoolean: {},
		spec.VarEnum:    {},
		spec.VarDate:    {},
	}
	allowedVarSources := map[spec.VarSource]struct{}{
		spec.SourceUser:   {},
		spec.SourceStatic: {},
		spec.SourceTool:   {},
	}

	varNames := map[string]spec.VarSource{}
	for i, v := range tpl.Variables {
		v.Name = strings.TrimSpace(v.Name)
		if err := bundleitemutils.ValidateTag(v.Name); err != nil {
			return fmt.Errorf("variables[%d]: invalid name %q, err %w", i, v.Name, err)
		}

		if _, dup := varNames[v.Name]; dup {
			return fmt.Errorf("duplicate variable %q", v.Name)
		}
		if _, ok := allowedVarTypes[v.Type]; !ok {
			return fmt.Errorf("variables[%d]: invalid type %q", i, v.Type)
		}
		if _, ok := allowedVarSources[v.Source]; !ok {
			return fmt.Errorf("variables[%d]: invalid source %q", i, v.Source)
		}

		// Source-specific requirements.
		switch v.Source {
		case spec.SourceStatic:
			if strings.TrimSpace(v.StaticVal) == "" {
				return fmt.Errorf("variables[%d]: source 'static' requires staticVal", i)
			}
			if v.Required {
				return fmt.Errorf("variables[%d]: static variables cannot be required", i)
			}
		case spec.SourceTool:

			if strings.TrimSpace(string(v.ToolBundleID)) == "" {
				return fmt.Errorf("variables[%d]: source 'tool' requires toolBundleID", i)
			}
			if err := bundleitemutils.ValidateItemSlug(v.ToolSlug); err != nil {
				return fmt.Errorf("variables[%d]: invalid toolSlug %q err %w", i, v.ToolSlug, err)
			}
			if err := bundleitemutils.ValidateItemVersion(v.ToolVersion); err != nil {
				return fmt.Errorf(
					"variables[%d]: invalid toolVersion %q err %w",
					i,
					v.ToolVersion,
					err,
				)
			}
		case spec.SourceUser:
			// No validations needed.
		}

		// Type-specific requirements.
		if v.Type == spec.VarEnum {
			if len(v.EnumValues) == 0 {
				return fmt.Errorf("variables[%d]: enum type requires enumValues", i)
			}
			seenEnum := map[string]struct{}{}
			for j, ev := range v.EnumValues {
				ev = strings.TrimSpace(ev)
				if ev == "" {
					return fmt.Errorf("variables[%d]: enumValues[%d] is empty", i, j)
				}
				if _, dup := seenEnum[ev]; dup {
					return fmt.Errorf("variables[%d]: duplicate enum value %q", i, ev)
				}
				seenEnum[ev] = struct{}{}
			}
		}

		varNames[v.Name] = v.Source
	}

	// All placeholders must be declared.
	for ph := range placeholders {
		if _, ok := varNames[ph]; !ok {
			return fmt.Errorf("variable %q used in blocks but not declared", ph)
		}
	}

	// Validate pre-processors.
	allowedOnError := map[spec.PreProcessorOnError]struct{}{
		"":                {},
		spec.OnErrorEmpty: {},
		spec.OnErrorFail:  {},
	}
	saveTargets := map[string]struct{}{}

	for i, p := range tpl.PreProcessors {

		if strings.TrimSpace(string(p.ToolBundleID)) == "" {
			return fmt.Errorf("preProcessors[%d]: toolBundleID is empty", i)
		}

		if err := bundleitemutils.ValidateItemSlug(p.ToolSlug); err != nil {
			return fmt.Errorf("preProcessors[%d]: invalid toolSlug %q err %w", i, p.ToolSlug, err)
		}
		if err := bundleitemutils.ValidateItemVersion(p.ToolVersion); err != nil {
			return fmt.Errorf(
				"preProcessors[%d]: invalid toolVersion %q err %w",
				i,
				p.ToolVersion,
				err,
			)
		}
		if err := bundleitemutils.ValidateTag(p.SaveAs); err != nil {
			return fmt.Errorf("preProcessors[%d]: invalid saveAs %q err %w", i, p.SaveAs, err)
		}

		src, ok := varNames[p.SaveAs]
		if !ok {
			return fmt.Errorf(
				"preProcessors[%d]: saveAs %q is not a declared variable",
				i,
				p.SaveAs,
			)
		}
		if src != spec.SourceTool {
			return fmt.Errorf("preProcessors[%d]: variable %q must have source 'tool'", i, p.SaveAs)
		}
		if _, dup := saveTargets[p.SaveAs]; dup {
			return fmt.Errorf("multiple preProcessors write to %q", p.SaveAs)
		}
		saveTargets[p.SaveAs] = struct{}{}

		if _, ok := allowedOnError[p.OnError]; !ok {
			return fmt.Errorf("preProcessors[%d]: invalid onError %q", i, p.OnError)
		}
	}

	// Detect unused variables.
	for n, src := range varNames {
		if _, used := placeholders[n]; used {
			continue
		}
		if _, produced := saveTargets[n]; produced {
			continue
		}
		if src == spec.SourceStatic {
			// Static variables may legitimately stay unused in blocks.
			continue
		}
		return fmt.Errorf("variable %q declared but never used", n)
	}

	if err := bundleitemutils.ValidateTags(tpl.Tags); err != nil {
		return err
	}
	return nil
}
