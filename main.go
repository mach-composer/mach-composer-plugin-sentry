package main

import (
	"github.com/mach-composer/mach-composer-plugin-sdk/v2/plugin"
	"github.com/mach-composer/mach-composer-plugin-sdk/v2/schema"

	"github.com/mach-composer/mach-composer-plugin-sentry/internal"
)

func main() {
	p := internal.NewSentryPlugin()
	plugin.ServePlugin(plugin.NewPlugin(&schema.PluginSchema{
		Identifier: "sentry",

		Configure: p.Configure,

		GetValidationSchema: p.GetValidationSchema,

		// Config
		SetGlobalConfig:        p.SetGlobalConfig,
		SetSiteConfig:          p.SetSiteConfig,
		SetSiteComponentConfig: p.SetSiteComponentConfig,
		SetComponentConfig:     p.SetComponentConfig,

		// Renders
		RenderTerraformProviders: p.RenderTerraformProviders,
		RenderTerraformResources: p.RenderTerraformResources,
		RenderTerraformComponent: p.RenderTerraformComponent,
	}))
}
