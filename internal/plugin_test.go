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

func TestSetComponentConfig(t *testing.T) {
	p := NewSentryPlugin()

	err := p.SetComponentConfig("my-component", "abc123", map[string]any{})
	assert.NoError(t, err)
	assert.Equal(t, 1, len(p.componentConfigs))

	assert.Equal(t, ComponentConfig{
		Version: "abc123",
	}, p.componentConfigs["my-component"])
}
