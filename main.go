package main

import (
	"github.com/mach-composer/mach-composer-plugin-sdk/plugin"

	"github.com/mach-composer/mach-composer-plugin-sentry/internal"
)

func main() {
	p := internal.NewSentryPlugin()
	plugin.ServePlugin(p)
}
