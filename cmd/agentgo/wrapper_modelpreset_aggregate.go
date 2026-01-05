package main

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/ppipada/flexigpt-app/internal/inferencewrapper"
	inferencewrapperSpec "github.com/ppipada/flexigpt-app/internal/inferencewrapper/spec"
	modelpresetSpec "github.com/ppipada/flexigpt-app/internal/modelpreset/spec"
	settingSpec "github.com/ppipada/flexigpt-app/internal/setting/spec"
	inferencegoSpec "github.com/ppipada/inference-go/spec"
)

func InitProviderSetUsingSettingsAndPresets(
	mpw *ModelPresetStoreWrapper,
	s *SettingStoreWrapper,
	p *ProviderSetWrapper,
) error {
	allProviders, err := GetAllProviderPresets(mpw)
	if err != nil {
		return err
	}
	keySecrets, err := getAllProviderSecrets(s)
	if err != nil {
		return err
	}

	if err := initProviders(p, allProviders, keySecrets); err != nil {
		return err
	}

	slog.Info("initProviderSetUsingSettingsAndPresets completed",
		"authKeys", len(keySecrets))

	return nil
}

func GetAllProviderPresets(
	mpw *ModelPresetStoreWrapper,
) ([]modelpresetSpec.ProviderPreset, error) {
	const maxSafetyHops = 16

	var (
		all   []modelpresetSpec.ProviderPreset
		token string
		hops  int
	)

	for {
		resp, err := mpw.ListProviderPresets(&modelpresetSpec.ListProviderPresetsRequest{
			IncludeDisabled: true,
			PageSize:        modelpresetSpec.MaxPageSize,
			PageToken:       token,
		})
		if err != nil {
			return nil, err
		}
		if resp.Body == nil {
			break
		}
		all = append(all, resp.Body.Providers...)

		if resp.Body.NextPageToken == nil || *resp.Body.NextPageToken == "" {
			break
		}
		if hops >= maxSafetyHops {
			return nil, fmt.Errorf("pagination exceeded %d hops - aborting", maxSafetyHops)
		}
		token = *resp.Body.NextPageToken
		hops++
	}
	return all, nil
}

// getAllProviderSecrets fetches every secret once and caches them in-mem.
func getAllProviderSecrets(
	s *SettingStoreWrapper,
) (map[string]string, error) {
	resp, err := s.GetSettings(&settingSpec.GetSettingsRequest{})
	if err != nil {
		return nil, err
	}
	if resp.Body == nil {
		return nil, errors.New("GetSettings: empty response body")
	}

	secrets := make(map[string]string, len(resp.Body.AuthKeys))
	for _, meta := range resp.Body.AuthKeys {
		if meta.Type != settingSpec.AuthKeyTypeProvider {
			continue
		}
		secResp, err := s.GetAuthKey(&settingSpec.GetAuthKeyRequest{
			Type:    meta.Type,
			KeyName: meta.KeyName,
		})
		if err != nil {
			return nil, err
		}
		if secResp.Body != nil && secResp.Body.Secret != "" {
			secrets[string(meta.KeyName)] = secResp.Body.Secret
		}
	}
	return secrets, nil
}

// BuildAddProviderRequests merges presets + secrets.
// Only providers that have a (valid) preset are considered.  If a matching
// secret exists its value is copied into the request.
func initProviders(
	providerAPI *ProviderSetWrapper,
	providers []modelpresetSpec.ProviderPreset,
	secrets map[string]string,
) error {
	providersAdded := 0
	providersWithAPIKey := 0
	for _, pp := range providers {
		if pp.Name == "" || pp.Origin == "" {
			slog.Warn("skipping provider with invalid preset", "name", pp.Name)
			continue
		}

		body := &inferencewrapperSpec.AddProviderRequestBody{
			SDKType:                  inferencewrapper.ConvertModelPresetToInferencegoSDKType(pp.SDKType),
			Origin:                   pp.Origin,
			ChatCompletionPathPrefix: pp.ChatCompletionPathPrefix,
			APIKeyHeaderKey:          pp.APIKeyHeaderKey,
			DefaultHeaders:           pp.DefaultHeaders,
		}
		r := &inferencewrapperSpec.AddProviderRequest{
			Provider: inferencegoSpec.ProviderName(string(pp.Name)),
			Body:     body,
		}
		if _, err := providerAPI.AddProvider(r); err != nil {
			return fmt.Errorf("add provider failed. name: %s, err: %w ", pp.Name, err)
		}
		providersAdded++
		if secret, ok := secrets[string(pp.Name)]; ok {
			_, err := providerAPI.SetProviderAPIKey(&inferencewrapperSpec.SetProviderAPIKeyRequest{
				Provider: inferencegoSpec.ProviderName(string(pp.Name)),
				Body: &inferencewrapperSpec.SetProviderAPIKeyRequestBody{
					APIKey: secret,
				},
			})
			if err != nil {
				return fmt.Errorf("set provider api key failed. name: %s, err: %w ", pp.Name, err)
			}
			providersWithAPIKey++
		}
	}

	if providersAdded == 0 {
		slog.Warn("no providers found - nothing to initialize")
	}
	if providersWithAPIKey == 0 {
		slog.Warn("no providers with APIKey")
	}

	return nil
}
