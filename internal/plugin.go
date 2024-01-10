package internal

import (
	_ "dario.cat/mergo"
	"embed"
	"fmt"
	"github.com/hashicorp/go-hclog"
	"github.com/mach-composer/mach-composer-plugin-helpers/helpers"
	"github.com/mach-composer/mach-composer-plugin-sdk/v2/schema"
	"github.com/mitchellh/mapstructure"
	"sync"
)

//go:embed templates/*
var templates embed.FS

//go:embed schemas/*
var schemas embed.FS

type SentryPlugin struct {
	environment      string
	provider         string
	globalConfig     GlobalConfig
	siteConfigs      map[string]SiteConfig
	componentConfigs map[string]ComponentConfig
}

func NewSentryPlugin() *SentryPlugin {
	state := &SentryPlugin{
		provider:         "1.0.2",
		siteConfigs:      map[string]SiteConfig{},
		componentConfigs: map[string]ComponentConfig{},
	}

	return state
}

func (p *SentryPlugin) Configure(environment string, provider string) error {
	p.environment = environment
	if provider != "" {
		p.provider = provider
	}
	return nil
}

func (p *SentryPlugin) IsEnabled() bool {
	return p.globalConfig.AuthToken != ""
}

func (p *SentryPlugin) GetValidationSchema() (*schema.ValidationSchema, error) {
	s := &schema.ValidationSchema{}

	if err := loadSchemaNode("schemas/global-config.json", &s.GlobalConfigSchema); err != nil {
		return nil, err
	}

	if err := loadSchemaNode("schemas/site-config.json", &s.SiteConfigSchema); err != nil {
		return nil, err
	}

	if err := loadSchemaNode("schemas/site-component-config.json", &s.SiteComponentConfigSchema); err != nil {
		return nil, err
	}

	return s, nil
}

func (p *SentryPlugin) SetGlobalConfig(data map[string]any) error {
	if err := validate("schemas/global-config.json", data); err != nil {
		return fmt.Errorf("invalid global config: %w", err)
	}

	cfg := defaultGlobalConfig
	if err := mapstructure.Decode(data, &cfg); err != nil {
		return err
	}
	p.globalConfig = cfg
	return nil
}

func (p *SentryPlugin) SetSiteConfig(site string, data map[string]any) error {
	if err := validate("schemas/site-config.json", data); err != nil {
		return fmt.Errorf("invalid site config for site %s: %w", site, err)
	}

	cfg := defaultSiteConfig
	if err := mapstructure.Decode(data, &cfg); err != nil {
		return err
	}
	p.siteConfigs[site] = cfg
	return nil
}

func (p *SentryPlugin) SetSiteComponentConfig(site string, component string, data map[string]any) error {
	if err := validate("schemas/site-component-config.json", data); err != nil {
		return err
	}

	cfg := defaultSiteComponentConfig
	if err := mapstructure.Decode(data, &cfg); err != nil {
		return err
	}

	siteCfg, ok := p.siteConfigs[site]
	if !ok {
		siteCfg = defaultSiteConfig
		p.siteConfigs[site] = siteCfg
	}
	siteCfg.Components[component] = cfg

	return nil
}

func (p *SentryPlugin) SetComponentConfig(component, version string, data map[string]any) error {
	cfg := ComponentConfig{
		Version: version,
	}
	if err := mapstructure.Decode(data, &cfg); err != nil {
		return err
	}

	p.componentConfigs[component] = cfg

	return nil
}

func (p *SentryPlugin) RenderTerraformProviders(_ string) (string, error) {
	if !p.IsEnabled() {
		hclog.Default().Warn("Sentry plugin provider rendering is disabled. Set auth_token to enable")
		return "", nil
	}
	result := fmt.Sprintf(`
		sentry = {
			source = "labd/sentry"
			version = "%s"
		}`, helpers.VersionConstraint(p.provider))
	return result, nil
}

func (p *SentryPlugin) RenderTerraformResources(_ string) (string, error) {
	if !p.IsEnabled() {
		hclog.Default().Warn("Sentry plugin resource rendering is disabled. Set auth_token to enable")
		return "", nil
	}

	templateContext := struct {
		Token string
		URL   string
	}{
		Token: p.globalConfig.AuthToken,
		URL:   p.globalConfig.BaseURL,
	}

	tpl, err := templates.ReadFile("templates/provider.tmpl")
	if err != nil {
		return "", err
	}

	return helpers.RenderGoTemplate(string(tpl), templateContext)
}

var (
	warnOnce sync.Once
)

func (p *SentryPlugin) RenderTerraformComponent(site string, component string) (*schema.ComponentSchema, error) {
	siteComponentConfig := p.getSiteComponentConfig(site, component)
	componentConfig, err := p.getComponentConfig(component)
	if err != nil {
		return nil, err
	}

	vars := fmt.Sprintf("sentry_dsn = \"%s\"", siteComponentConfig.DSN)
	if p.globalConfig.AuthToken != "" {
		vars = fmt.Sprintf("sentry_dsn = sentry_key.%s.dsn_secret", component)
	}

	result := &schema.ComponentSchema{
		Variables: vars,
	}

	if !p.IsEnabled() {
		warnOnce.Do(func() {
			hclog.Default().Warn("Sentry plugin component rendering is disabled. Set auth_token to enable")
		})
		return result, nil
	}

	resources, err := terraformRenderComponentResources(site, component, componentConfig.Version, p.environment, p.globalConfig, siteComponentConfig)
	if err != nil {
		return nil, err
	}
	result.Resources = resources

	return result, nil
}

func (p *SentryPlugin) getSiteConfig(site string) SiteConfig {
	cfg, ok := p.siteConfigs[site]
	if !ok {
		cfg = defaultSiteConfig
	}
	return cfg.extendGlobalConfig(p.globalConfig)
}

func (p *SentryPlugin) getSiteComponentConfig(site, name string) SiteComponentConfig {
	siteCfg := p.getSiteConfig(site)
	cfg := siteCfg.getSiteComponentConfig(name)
	return cfg
}

func (p *SentryPlugin) getComponentConfig(component string) (ComponentConfig, error) {
	cfg, ok := p.componentConfigs[component]
	if !ok {
		return ComponentConfig{}, fmt.Errorf("component %s not found", component)
	}
	return cfg, nil
}

func terraformRenderComponentResources(site, component, componentVersion, environment string, globalCfg GlobalConfig,
	cfg SiteComponentConfig) (string, error) {

	trackDeployments := false
	if cfg.TrackDeployments != nil {
		trackDeployments = *cfg.TrackDeployments
	}

	templateContext := struct {
		SiteName         string
		ComponentName    string
		ComponentVersion string
		Environment      string
		TrackDeployments bool
		Global           GlobalConfig
		Config           SiteComponentConfig
	}{
		SiteName:         site,
		ComponentName:    component,
		ComponentVersion: componentVersion,
		Environment:      environment,
		TrackDeployments: trackDeployments,
		Global:           globalCfg,
		Config:           cfg,
	}

	tpl, err := templates.ReadFile("templates/resources.tmpl")
	if err != nil {
		return "", err
	}

	return helpers.RenderGoTemplate(string(tpl), templateContext)
}
