package spec

import "time"

type AssistantID string

type Assistant struct {
	ID          AssistantID `json:"id"`
	Name        string      `json:"name"`
	DisplayName string      `json:"displayName,omitempty"`
	Description string      `json:"description,omitempty"`

	TemplateID    string   `json:"templateId"`
	ModelPresetID string   `json:"modelPresetId"`
	ToolBundleIDs []string `json:"toolBundleIds,omitempty"`

	// Default values for template variables (optional override).
	DefaultTemplateVars map[string]string `json:"defaultTemplateVars,omitempty"`

	Version    int       `json:"version"`
	IsEnabled  bool      `json:"isEnabled"`
	CreatedAt  time.Time `json:"createdAt"`
	ModifiedAt time.Time `json:"modifiedAt"`
	Tags       []string  `json:"tags,omitempty"`
}
