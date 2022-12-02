package plugin

import (
	"github.com/mach-composer/mach-composer-plugin-sdk/plugin"

	"github.com/mach-composer/mach-composer-plugin-sentry/internal"
)

func Serve() {
	p := internal.NewSentryPlugin()
	plugin.ServePlugin(p)
}
