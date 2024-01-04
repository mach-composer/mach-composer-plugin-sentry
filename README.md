# Sentry Plugin for MACH Composer

This repository contains the Sentry plugin for Mach Composer. It requires MACH
composer >= 2.5.x

This plugin uses the [Lab Digital Sentry Terraform Provider](https://registry.terraform.io/providers/labd/sentry/latest)

## Usage

```yaml
mach_composer:
  version: 1
  plugins:
    sentry:
      source: mach-composer/sentry
      version: 0.1.3

global:
  # ...
  sentry:
    auth_token: "token"
    organization: "org"
    project: "default project"
    rate_limit_window: 21600
    rate_limit_count: 100

sites:
  - identifier: my-site
    # ...
    components:
      - name: my-component
        # ...
        sentry:
          project: "component project" # override default
```
