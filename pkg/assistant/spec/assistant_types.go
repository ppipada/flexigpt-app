package spec

import "time"

type Assistant struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"displayName,omitempty"`
	Description string `json:"description,omitempty"`
	InvokeSlug  string `json:"invokeSlug,omitempty"`
	IsEnabled   bool   `json:"isEnabled"`

	TemplateID               string            `json:"templateId"`
	ModelPresetID            string            `json:"modelPresetId"`
	ToolBundleIDs            []string          `json:"toolBundleIds,omitempty"`
	DefaultTemplateVarValues map[string]string `json:"defaultTemplateVarValues,omitempty"`

	Version    string    `json:"version"`
	CreatedAt  time.Time `json:"createdAt"`
	ModifiedAt time.Time `json:"modifiedAt"`
}
