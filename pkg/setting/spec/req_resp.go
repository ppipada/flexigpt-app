package spec

type SetAppThemeRequestBody struct {
	Type ThemeType `json:"type" required:"true"`
	Name string    `json:"name" required:"true"`
}
type SetAppThemeRequest struct {
	Body *SetAppThemeRequestBody
}

type SetAppThemeResponse struct{}

// AuthKeyMeta is the public view of one stored key (no secret, only SHA).
type AuthKeyMeta struct {
	Type     AuthKeyType `json:"type"`
	KeyName  AuthKeyName `json:"keyName"`
	SHA256   string      `json:"sha256"`
	NonEmpty bool        `json:"nonEmpty"`
}

// GetAuthKeyRequest fetches one decrypted secret.
type GetAuthKeyRequest struct {
	Type    AuthKeyType `path:"type"`
	KeyName AuthKeyName `path:"keyName"`
}

// GetAuthKeyResponse returns the secret and its SHA.
type GetAuthKeyResponseBody struct {
	Secret   string `json:"secret"`
	SHA256   string `json:"sha256"`
	NonEmpty bool   `json:"nonEmpty"`
}
type GetAuthKeyResponse struct {
	Body *GetAuthKeyResponseBody
}

type SetAuthKeyRequestBody struct {
	Secret string `json:"secret" required:"true"`
}

// SetAuthKeyRequest creates or updates a key (idempotent).
type SetAuthKeyRequest struct {
	Type    AuthKeyType `path:"type"`
	KeyName AuthKeyName `path:"keyName"`
	Body    *SetAuthKeyRequestBody
}

type SetAuthKeyResponse struct{}

// DeleteAuthKeyRequest removes a key (if not built-in).
type DeleteAuthKeyRequest struct {
	Type    AuthKeyType `path:"type"`
	KeyName AuthKeyName `path:"keyName"`
}

type DeleteAuthKeyResponse struct{}

// GetSettingsRequest fetches everything (theme + keys).  Secrets are omitted.
type GetSettingsRequest struct {
	ForceFetch bool `query:"forceFetch" doc:"Refresh from disk before reading." required:"false"`
}

type GetSettingsResponseBody struct {
	AppTheme AppTheme      `json:"appTheme"`
	AuthKeys []AuthKeyMeta `json:"authKeys"`
}

// GetSettingsResponse returns the current settings without secrets.
type GetSettingsResponse struct {
	Body *GetSettingsResponseBody
}
