package internal

// BaseConfig is the base sentry config.
type BaseConfig struct {
	DSN              string `mapstructure:"dsn"`
	RateLimitWindow  *int   `mapstructure:"rate_limit_window"`
	RateLimitCount   *int   `mapstructure:"rate_limit_count"`
	Project          string `mapstructure:"project"`
	TrackDeployments *bool  `mapstructure:"track_deployments"`
}

// GlobalConfig global Sentry configuration.
type GlobalConfig struct {
	BaseConfig   `mapstructure:",squash"`
	AuthToken    string `mapstructure:"auth_token"`
	BaseURL      string `mapstructure:"base_url"`
	Organization string `mapstructure:"organization"`
}

var defaultGlobalConfig = GlobalConfig{
	BaseConfig: BaseConfig{
		TrackDeployments: boolPtr(true),
	},
}

// SiteConfig is for site specific sentry DSN settings
type SiteConfig struct {
	BaseConfig `mapstructure:",squash"`
	Components map[string]SiteComponentConfig `mapstructure:"-"`
}

var defaultSiteConfig = SiteConfig{
	Components: map[string]SiteComponentConfig{},
}

// SiteComponentConfig is for component specific sentry DSN settings
type SiteComponentConfig struct {
	BaseConfig `mapstructure:",squash"`
}

var defaultSiteComponentConfig = SiteComponentConfig{}

// ComponentConfig is for general component information
type ComponentConfig struct {
	Version string `mapstructure:"-"`
}

func (c *SiteConfig) extendGlobalConfig(g GlobalConfig) SiteConfig {
	cfg := SiteConfig{
		BaseConfig: g.BaseConfig,
		Components: c.Components,
	}
	if c.DSN != "" {
		cfg.DSN = c.DSN
	}
	if c.RateLimitWindow != nil {
		cfg.RateLimitWindow = c.RateLimitWindow
	}
	if c.RateLimitCount != nil {
		cfg.RateLimitCount = c.RateLimitCount
	}
	if c.Project != "" {
		cfg.Project = c.Project
	}
	if c.TrackDeployments != nil {
		cfg.TrackDeployments = c.TrackDeployments
	}
	return cfg
}

func (c *SiteComponentConfig) extendSiteConfig(s SiteConfig) SiteComponentConfig {
	cfg := SiteComponentConfig{
		BaseConfig: s.BaseConfig,
	}

	if c.DSN != "" {
		cfg.DSN = c.DSN
	}
	if c.RateLimitWindow != nil {
		cfg.RateLimitWindow = c.RateLimitWindow
	}
	if c.RateLimitCount != nil {
		cfg.RateLimitCount = c.RateLimitCount
	}
	if c.Project != "" {
		cfg.Project = c.Project
	}
	if c.TrackDeployments != nil {
		cfg.TrackDeployments = c.TrackDeployments
	}
	return cfg
}

func (c *SiteConfig) getSiteComponentConfig(name string) SiteComponentConfig {
	compConfig, ok := c.Components[name]
	if !ok {
		compConfig = defaultSiteComponentConfig
	}
	return compConfig.extendSiteConfig(*c)
}
