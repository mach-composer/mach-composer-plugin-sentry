package internal

import (
	"fmt"

	"github.com/mach-composer/mach-composer-plugin-helpers/helpers"
	"github.com/mach-composer/mach-composer-plugin-sdk/plugin"
	"github.com/mach-composer/mach-composer-plugin-sdk/schema"
	"github.com/mitchellh/mapstructure"
)

type SentryPlugin struct {
	environment  string
	provider     string
	globalConfig GlobalConfig
	siteConfigs  map[string]*SiteConfig
}

func NewSentryPlugin() schema.MachComposerPlugin {
	state := &SentryPlugin{
		provider:    "0.11.1",
		siteConfigs: map[string]*SiteConfig{},
	}

	return plugin.NewPlugin(&schema.PluginSchema{
		Identifier: "sentry",

		Configure: state.Configure,
		IsEnabled: state.IsEnabled,

		GetValidationSchema: state.GetValidationSchema,

		// Config
		SetGlobalConfig:        state.SetGlobalConfig,
		SetSiteConfig:          state.SetSiteConfig,
		SetSiteComponentConfig: state.SetSiteComponentConfig,

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
		Components: map[string]ComponentConfig{},
	}
	if err := mapstructure.Decode(data, &cfg); err != nil {
		return err
	}
	p.siteConfigs[site] = &cfg
	return nil
}

func (p *SentryPlugin) SetSiteComponentConfig(site string, component string, data map[string]any) error {
	cfg := ComponentConfig{}
	if err := mapstructure.Decode(data, &cfg); err != nil {
		return err
	}

	siteCfg, ok := p.siteConfigs[site]
	if !ok {
		siteCfg = &SiteConfig{
			Components: map[string]ComponentConfig{},
		}
		p.siteConfigs[site] = siteCfg
	}
	siteCfg.Components[component] = cfg

	return nil
}

func (p *SentryPlugin) TerraformRenderProviders(site string) (string, error) {
	if p.globalConfig.AuthToken == "" {
		return "", nil
	}
	result := fmt.Sprintf(`
		sentry = {
			source = "jianyuan/sentry"
			version = "%s"
		}`, helpers.VersionConstraint(p.provider))
	return result, nil
}

func (p *SentryPlugin) TerraformRenderResources(site string) (string, error) {
	cfg := p.getSiteConfig(site)
	if cfg == nil {
		return "", nil
	}

	if p.globalConfig.AuthToken == "" {
		return "", nil
	}

	templateContext := struct {
		Token string
		URL   string
	}{
		Token: helpers.SerializeToHCL("token", p.globalConfig.AuthToken),
		URL:   p.globalConfig.BaseURL,
	}

	template := `
		provider "sentry" {
			{{ .Token }}
			base_url = {{ if .URL }}{{ .URL|printf "%q" }}{{ else }}"https://sentry.io/api/"{{ end }}
		}
	`
	return helpers.RenderGoTemplate(template, templateContext)
}

func (p *SentryPlugin) RenderTerraformComponent(site string, component string) (*schema.ComponentSchema, error) {
	cfg := p.getComponentSiteConfig(site, component)

	vars := fmt.Sprintf("sentry_dsn = \"%s\"", cfg.DSN)
	if p.globalConfig.AuthToken != "" {
		vars = fmt.Sprintf("sentry_dsn = sentry_key.%s.dsn_secret", component)
	}

	resources, err := terraformRenderComponentResources(site, component, cfg, &p.globalConfig)
	if err != nil {
		return nil, err
	}

	result := &schema.ComponentSchema{
		Variables: vars,
		Resources: resources,
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

func (p *SentryPlugin) getComponentSiteConfig(site, name string) *ComponentConfig {
	siteCfg := p.getSiteConfig(site)
	if siteCfg == nil {
		return nil
	}
	cfg := siteCfg.getComponentSiteConfig(name)
	cfg.Environment = p.environment
	return cfg
}

func terraformRenderComponentResources(site, component string, cfg *ComponentConfig, globalCfg *GlobalConfig) (string, error) {
	if globalCfg.AuthToken == "" {
		return "", nil
	}

	templateContext := struct {
		ComponentName string
		SiteName      string
		Global        *GlobalConfig
		Config        *ComponentConfig
	}{
		ComponentName: component,
		SiteName:      site,
		Global:        globalCfg,
		Config:        cfg,
	}

	template := `
	resource "sentry_key" "{{ .ComponentName }}" {
		organization      = {{ .Global.Organization|printf "%q" }}
		project           = {{ .Config.Project|printf "%q" }}
		name              = "{{ .SiteName }}-{{ .Config.Environment }}-{{ .ComponentName }}"
		{{ if .Config.RateLimitWindow }}
		rate_limit_window = {{ .Config.RateLimitWindow }}
		{{ end }}
		{{ if .Config.RateLimitCount }}
		rate_limit_count  = {{ .Config.RateLimitCount }}
		{{ end }}
	}
	`
	return helpers.RenderGoTemplate(template, templateContext)
}
