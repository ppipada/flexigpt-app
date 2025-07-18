package store

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/ppipada/flexigpt-app/pkg/prompt/spec"
)

// Regular expressions.
var (
	placeholderRE = regexp.MustCompile(`\{\{([a-zA-Z_][a-zA-Z0-9_]*)\}\}`)
	nameRE        = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)
)

// validateTemplate performs a structural and referential integrity check.
func validateTemplate(tpl *spec.PromptTemplate) error {
	if tpl == nil {
		return errors.New("template is nil")
	}
	if strings.TrimSpace(tpl.DisplayName) == "" {
		return errors.New("displayName is empty")
	}
	if strings.TrimSpace(string(tpl.Slug)) == "" {
		return errors.New("slug is empty")
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
		if b.ID == "" {
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
		if !nameRE.MatchString(v.Name) {
			return fmt.Errorf("variables[%d]: invalid name %q", i, v.Name)
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
			if v.StaticVal == "" {
				return fmt.Errorf("variables[%d]: source 'static' requires staticVal", i)
			}
			if v.Required {
				return fmt.Errorf("variables[%d]: static variables cannot be required", i)
			}
		case spec.SourceTool:
			if v.ToolID == "" {
				return fmt.Errorf("variables[%d]: source 'tool' requires toolId", i)
			}
		}

		// Type-specific requirements.
		if v.Type == spec.VarEnum && len(v.EnumValues) == 0 {
			return fmt.Errorf("variables[%d]: enum type requires enumValues", i)
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
		if p.ToolID == "" {
			return fmt.Errorf("preProcessors[%d]: toolId is empty", i)
		}
		if !nameRE.MatchString(p.SaveAs) {
			return fmt.Errorf("preProcessors[%d]: invalid saveAs %q", i, p.SaveAs)
		}
		if _, ok := varNames[p.SaveAs]; !ok {
			return fmt.Errorf(
				"preProcessors[%d]: saveAs %q is not a declared variable",
				i,
				p.SaveAs,
			)
		}
		if varNames[p.SaveAs] != spec.SourceTool {
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
		if src == spec.SourceStatic && tpl.Version == "" {
			// Static variables may legitimately stay unused in blocks.
			continue
		}
		return fmt.Errorf("variable %q declared but never used", n)
	}

	// Validate tags.
	tagSeen := map[string]struct{}{}
	for i, t := range tpl.Tags {
		if !nameRE.MatchString(t) {
			return fmt.Errorf("tags[%d]: invalid tag %q", i, t)
		}
		if _, dup := tagSeen[t]; dup {
			return fmt.Errorf("duplicate tag %q", t)
		}
		tagSeen[t] = struct{}{}
	}

	return nil
}
