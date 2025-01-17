package spec

type GetAllSettingsRequest struct {
	ForceFetch bool `query:"forceFetch" doc:"Force refresh the settings and get" required:"false"`
}

type GetAllSettingsResponse struct {
	Body *SettingsSchema
}

type SetSettingRequestBody struct {
	Value any `json:"value" required:"true" doc:"Value to be set"`
}

type SetSettingRequest struct {
	Key  string `path:"key" doc:"a dot separated setting key"`
	Body *SetSettingRequestBody
}

type SetSettingResponse struct{}
