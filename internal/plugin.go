package internal

import (
	"embed"
	"fmt"
	"log"

	"github.com/mach-composer/mach-composer-plugin-helpers/helpers"
	"github.com/mach-composer/mach-composer-plugin-sdk/v2/plugin"
	"github.com/mach-composer/mach-composer-plugin-sdk/v2/schema"
	"github.com/mitchellh/mapstructure"
)

//go:embed templates/*
var templates embed.FS

type SentryPlugin struct {
	environment      string
	provider         string
	globalConfig     GlobalConfig
	siteConfigs      map[string]*SiteConfig
	componentConfigs map[string]*ComponentConfig
}

func NewSentryPlugin() schema.MachComposerPlugin {
	state := &SentryPlugin{
		provider:         "1.0.2",
		siteConfigs:      map[string]*SiteConfig{},
		componentConfigs: map[string]*ComponentConfig{},
	}

	return plugin.NewPlugin(&schema.PluginSchema{
		Identifier: "sentry",

		Configure: state.Configure,

		GetValidationSchema: state.GetValidationSchema,

		// Config
		SetGlobalConfig:        state.SetGlobalConfig,
		SetSiteConfig:          state.SetSiteConfig,
		SetSiteComponentConfig: state.SetSiteComponentConfig,
		SetComponentConfig:     state.SetComponentConfig,

		// Renders
		RenderTerraformProviders: state.TerraformRenderProviders,
		RenderTerraformResources: state.TerraformRenderResources,
		RenderTerraformComponent: state.RenderTerraformComponent,
	})
}

func (p *SentryPlugin) Configure(environment string, provider string) error {
	p.environment = environment
	if provider != "" {
		p.provider = provider
	}
	return nil
}

func (p *SentryPlugin) IsEnabled() bool {
	return p.globalConfig.AuthToken != "" || p.globalConfig.DSN != ""
}

func (p *SentryPlugin) GetValidationSchema() (*schema.ValidationSchema, error) {
	log.Println("GetValidationSchema")
	result := getSchema()
	return result, nil
}

func (p *SentryPlugin) SetGlobalConfig(data map[string]any) error {
	if err := mapstructure.Decode(data, &p.globalConfig); err != nil {
		return err
	}
	return nil
}

func (p *SentryPlugin) SetSiteConfig(site string, data map[string]any) error {
	cfg := SiteConfig{
		Components: map[string]SiteComponentConfig{},
	}
	if err := mapstructure.Decode(data, &cfg); err != nil {
		return err
	}
	p.siteConfigs[site] = &cfg
	return nil
}

func (p *SentryPlugin) SetSiteComponentConfig(site string, component string, data map[string]any) error {
	cfg := SiteComponentConfig{}
	if err := mapstructure.Decode(data, &cfg); err != nil {
		return err
	}

	siteCfg, ok := p.siteConfigs[site]
	if !ok {
		siteCfg = &SiteConfig{
			Components: map[string]SiteComponentConfig{},
		}
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

	p.componentConfigs[component] = &cfg

	return nil
}

func (p *SentryPlugin) TerraformRenderProviders(_ string) (string, error) {
	if !p.IsEnabled() {
		return "", nil
	}
	result := fmt.Sprintf(`
		sentry = {
			source = "labd/sentry"
			version = "%s"
		}`, helpers.VersionConstraint(p.provider))
	return result, nil
}

func (p *SentryPlugin) TerraformRenderResources(site string) (string, error) {
	if !p.IsEnabled() {
		return "", nil
	}

	cfg := p.getSiteConfig(site)
	if cfg == nil {
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

func (p *SentryPlugin) RenderTerraformComponent(site string, component string) (*schema.ComponentSchema, error) {
	siteComponentConfig := p.getSiteComponentConfig(site, component)
	componentConfig := p.getComponentConfig(component)

	vars := fmt.Sprintf("sentry_dsn = \"%s\"", siteComponentConfig.DSN)
	if p.globalConfig.AuthToken != "" {
		vars = fmt.Sprintf("sentry_dsn = sentry_key.%s.dsn_secret", component)
	}

	result := &schema.ComponentSchema{
		Variables: vars,
	}

	if p.IsEnabled() {
		resources, err := terraformRenderComponentResources(site, component, componentConfig.Version, p.environment, &p.globalConfig, siteComponentConfig)
		if err != nil {
			return nil, err
		}
		result.Resources = resources
	}

	return result, nil
}

func (p *SentryPlugin) getSiteConfig(site string) *SiteConfig {
	cfg, ok := p.siteConfigs[site]
	if !ok {
		cfg = &SiteConfig{}
	}
	return cfg.extendGlobalConfig(&p.globalConfig)
}

func (p *SentryPlugin) getSiteComponentConfig(site, name string) *SiteComponentConfig {
	siteCfg := p.getSiteConfig(site)
	if siteCfg == nil {
		return nil
	}
	cfg := siteCfg.getSiteComponentConfig(name)
	return cfg
}

func (p *SentryPlugin) getComponentConfig(component string) *ComponentConfig {
	cfg, ok := p.componentConfigs[component]
	if !ok {
		cfg = &ComponentConfig{}
	}
	return cfg
}

func terraformRenderComponentResources(site, component, componentVersion, environment string, globalCfg *GlobalConfig,
	cfg *SiteComponentConfig) (string, error) {
	if globalCfg.AuthToken == "" {
		return "", nil
	}

	log.Println(environment)

	templateContext := struct {
		SiteName         string
		ComponentName    string
		ComponentVersion string
		Environment      string
		Global           *GlobalConfig
		Config           *SiteComponentConfig
	}{
		SiteName:         site,
		ComponentName:    component,
		ComponentVersion: componentVersion,
		Environment:      environment,
		Global:           globalCfg,
		Config:           cfg,
	}

	tpl, err := templates.ReadFile("templates/resources.tmpl")
	if err != nil {
		return "", err
	}

	return helpers.RenderGoTemplate(string(tpl), templateContext)
}
