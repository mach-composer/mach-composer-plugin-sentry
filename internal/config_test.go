package internal

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestExtendGlobalConfig(t *testing.T) {
	globalConfig := GlobalConfig{
		BaseConfig: BaseConfig{
			TrackDeployments: boolPtr(true),
		},
	}

	siteConfig := SiteConfig{
		BaseConfig: BaseConfig{
			TrackDeployments: boolPtr(false),
		},
	}

	extendedCfg := siteConfig.extendGlobalConfig(globalConfig)
	assert.Equal(t, false, *extendedCfg.TrackDeployments)
}

func TestExtendSiteConfig(t *testing.T) {
	siteCfg := SiteConfig{
		BaseConfig: BaseConfig{
			TrackDeployments: boolPtr(true),
		},
	}

	siteComponentConfig := SiteComponentConfig{
		BaseConfig: BaseConfig{
			TrackDeployments: boolPtr(false),
		},
	}

	extendedCfg := siteComponentConfig.extendSiteConfig(siteCfg)
	assert.Equal(t, false, *extendedCfg.TrackDeployments)
}
