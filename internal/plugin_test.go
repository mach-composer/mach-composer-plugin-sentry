package internal

import (
	"github.com/mach-composer/mach-composer-plugin-sdk/v2/schema"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetValidationSchema(t *testing.T) {
	p := NewSentryPlugin()
	s, err := p.GetValidationSchema()
	assert.NoError(t, err)
	assert.IsType(t, schema.ValidationSchema{}, *s)
}

func TestSetGlobalConfigFull(t *testing.T) {
	p := NewSentryPlugin()

	err := p.SetGlobalConfig(map[string]any{
		"dsn":               "https://sentry.io/123",
		"rate_limit_window": 1000,
		"rate_limit_count":  10,
		"project":           "test",
		"track_deployments": true,
		"auth_token":        "foobar",
		"base_url":          "https://sentry.io",
		"organization":      "my-org",
	})
	assert.NoError(t, err)
	assert.Equal(t, GlobalConfig{
		BaseConfig: BaseConfig{
			DSN:              "https://sentry.io/123",
			RateLimitWindow:  intPtr(1000),
			RateLimitCount:   intPtr(10),
			TrackDeployments: boolPtr(true),
			ExposeKey:        boolPtr(false),
			Project:          "test",
		},
		AuthToken:    "foobar",
		BaseURL:      "https://sentry.io",
		Organization: "my-org",
	}, p.globalConfig)
}

func TestSetGlobalConfigInvalid(t *testing.T) {
	p := NewSentryPlugin()

	err := p.SetGlobalConfig(map[string]any{
		"foo": "bar",
		"bar": "baz",
	})
	assert.Error(t, err)
}

func TestSetGlobalConfigDefaults(t *testing.T) {
	p := NewSentryPlugin()

	err := p.SetGlobalConfig(map[string]any{})
	assert.NoError(t, err)
	assert.Equal(t, GlobalConfig{
		BaseConfig: BaseConfig{
			DSN:              "",
			RateLimitWindow:  nil,
			RateLimitCount:   nil,
			TrackDeployments: boolPtr(true),
			ExposeKey:        boolPtr(false),
		},
	}, p.globalConfig)
}

func TestSetGlobalConfigWithFalse(t *testing.T) {
	p := NewSentryPlugin()

	err := p.SetGlobalConfig(map[string]any{
		"track_deployments": false,
	})
	assert.NoError(t, err)
	assert.Equal(t, GlobalConfig{
		BaseConfig: BaseConfig{
			DSN:              "",
			RateLimitWindow:  nil,
			RateLimitCount:   nil,
			TrackDeployments: boolPtr(false),
			ExposeKey:        boolPtr(false),
		},
	}, p.globalConfig)
}

func TestSetSiteConfigFull(t *testing.T) {
	p := NewSentryPlugin()

	err := p.SetSiteConfig("my-site", map[string]any{
		"dsn":               "https://sentry.io/123",
		"rate_limit_window": 1000,
		"rate_limit_count":  10,
		"track_deployments": true,
	})
	assert.NoError(t, err)
	assert.Equal(t, 1, len(p.siteConfigs))

	assert.Equal(t, SiteConfig{
		BaseConfig: BaseConfig{
			DSN:              "https://sentry.io/123",
			RateLimitWindow:  intPtr(1000),
			RateLimitCount:   intPtr(10),
			TrackDeployments: boolPtr(true),
		},
		Components: map[string]SiteComponentConfig{},
	}, p.siteConfigs["my-site"])
}

func TestSetSiteConfigInvalid(t *testing.T) {
	p := NewSentryPlugin()

	err := p.SetSiteConfig("my-site", map[string]any{
		"foo": "bar",
		"bar": "baz",
	})
	assert.Error(t, err)
}

func TestSetSiteComponentConfigWithExistingSite(t *testing.T) {
	p := NewSentryPlugin()

	p.siteConfigs["my-site"] = SiteConfig{
		Components: map[string]SiteComponentConfig{},
	}

	err := p.SetSiteComponentConfig("my-site", "my-site-component", map[string]any{
		"dsn":               "https://sentry.io/123",
		"rate_limit_window": 1000,
		"rate_limit_count":  10,
		"track_deployments": true,
	})
	assert.NoError(t, err)
	assert.Equal(t, 1, len(p.siteConfigs["my-site"].Components))

	assert.Equal(t, SiteComponentConfig{
		BaseConfig: BaseConfig{
			DSN:              "https://sentry.io/123",
			RateLimitWindow:  intPtr(1000),
			RateLimitCount:   intPtr(10),
			TrackDeployments: boolPtr(true),
		},
	}, p.siteConfigs["my-site"].Components["my-site-component"])
}

func TestSetSiteComponentConfigWithNoSite(t *testing.T) {
	p := NewSentryPlugin()

	err := p.SetSiteComponentConfig("my-site", "my-site-component", map[string]any{
		"dsn":               "https://sentry.io/123",
		"rate_limit_window": 1000,
		"rate_limit_count":  10,
		"project":           "test",
		"track_deployments": true,
	})
	assert.NoError(t, err)
	assert.Equal(t, 1, len(p.siteConfigs["my-site"].Components))

	assert.Equal(t, SiteComponentConfig{
		BaseConfig: BaseConfig{
			DSN:              "https://sentry.io/123",
			RateLimitWindow:  intPtr(1000),
			RateLimitCount:   intPtr(10),
			TrackDeployments: boolPtr(true),
			Project:          "test",
		},
	}, p.siteConfigs["my-site"].Components["my-site-component"])
}

func TestSetSiteComponentConfigInvalid(t *testing.T) {
	p := NewSentryPlugin()

	err := p.SetSiteComponentConfig("my-site", "my-site-component", map[string]any{
		"foo": "bar",
		"bar": "baz",
	})
	assert.Error(t, err)
}

func TestRenderTerraformComponentWithoutAuthToken(t *testing.T) {
	p := NewSentryPlugin()

	p.SetGlobalConfig(map[string]any{})
	p.SetSiteComponentConfig("my-site", "my-component", map[string]any{
		"dsn": "https://sentry.io/123",
	})
	p.SetComponentConfig("my-component", "abc123", map[string]any{})

	result, err := p.RenderTerraformComponent("my-site", "my-component")
	assert.NoError(t, err)
	assert.Equal(t, `sentry_dsn = "https://sentry.io/123"`, result.Variables)
	assert.Empty(t, result.Resources)
}

func TestRenderTerraformComponentWithoutAuthTokenButExposeKey(t *testing.T) {
	p := NewSentryPlugin()

	p.SetGlobalConfig(map[string]any{
		"expose_key": true,
	})
	p.SetSiteComponentConfig("my-site", "my-component", map[string]any{
		"dsn": "https://sentry.io/123",
	})
	p.SetComponentConfig("my-component", "abc123", map[string]any{})

	result, err := p.RenderTerraformComponent("my-site", "my-component")
	assert.NoError(t, err)
	assert.Equal(t, `sentry_dsn = "https://sentry.io/123"`, result.Variables)
	assert.NotContains(t, result.Variables, "sentry_key =")
	assert.Empty(t, result.Resources)
}

func TestRenderTerraformComponentWithAuthToken(t *testing.T) {
	p := NewSentryPlugin()

	p.SetGlobalConfig(map[string]any{
		"auth_token":   "foobar",
		"organization": "my-org",
	})
	p.SetSiteComponentConfig("my-site", "my-component", map[string]any{
		"dsn": "https://sentry.io/123",
	})
	p.SetComponentConfig("my-component", "abc123", map[string]any{})

	result, err := p.RenderTerraformComponent("my-site", "my-component")
	assert.NoError(t, err)
	assert.Contains(t, result.Variables, "sentry_dsn = sentry_key.my-component.dsn_secret")
	assert.NotContains(t, result.Variables, "sentry_key =")
}

func TestRenderTerraformComponentWithExposeKeyGlobal(t *testing.T) {
	p := NewSentryPlugin()

	p.SetGlobalConfig(map[string]any{
		"auth_token":   "foobar",
		"organization": "my-org",
		"expose_key":   true,
	})
	p.SetSiteComponentConfig("my-site", "my-component", map[string]any{
		"dsn": "https://sentry.io/123",
	})
	p.SetComponentConfig("my-component", "abc123", map[string]any{})

	result, err := p.RenderTerraformComponent("my-site", "my-component")
	assert.NoError(t, err)
	assert.Contains(t, result.Variables, "sentry_dsn = sentry_key.my-component.dsn_secret")
	assert.Contains(t, result.Variables, "sentry_key = sentry_key.my-component.secret")
}

func TestRenderTerraformComponentWithExposeKeySiteOverride(t *testing.T) {
	p := NewSentryPlugin()

	p.SetGlobalConfig(map[string]any{
		"auth_token":   "foobar",
		"organization": "my-org",
		"expose_key":   true,
	})
	p.SetSiteConfig("my-site", map[string]any{
		"expose_key": false,
	})
	p.SetSiteComponentConfig("my-site", "my-component", map[string]any{
		"dsn": "https://sentry.io/123",
	})
	p.SetComponentConfig("my-component", "abc123", map[string]any{})

	result, err := p.RenderTerraformComponent("my-site", "my-component")
	assert.NoError(t, err)
	assert.Contains(t, result.Variables, "sentry_dsn = sentry_key.my-component.dsn_secret")
	assert.NotContains(t, result.Variables, "sentry_key =")
}

func TestRenderTerraformComponentWithExposeKeyComponentOverride(t *testing.T) {
	p := NewSentryPlugin()

	p.SetGlobalConfig(map[string]any{
		"auth_token":   "foobar",
		"organization": "my-org",
	})
	p.SetSiteComponentConfig("my-site", "my-component", map[string]any{
		"dsn":        "https://sentry.io/123",
		"expose_key": true,
	})
	p.SetComponentConfig("my-component", "abc123", map[string]any{})

	result, err := p.RenderTerraformComponent("my-site", "my-component")
	assert.NoError(t, err)
	assert.Contains(t, result.Variables, "sentry_dsn = sentry_key.my-component.dsn_secret")
	assert.Contains(t, result.Variables, "sentry_key = sentry_key.my-component.secret")
}

func TestSetComponentConfig(t *testing.T) {
	p := NewSentryPlugin()

	err := p.SetComponentConfig("my-component", "abc123", map[string]any{})
	assert.NoError(t, err)
	assert.Equal(t, 1, len(p.componentConfigs))

	assert.Equal(t, ComponentConfig{
		Version: "abc123",
	}, p.componentConfigs["my-component"])
}
