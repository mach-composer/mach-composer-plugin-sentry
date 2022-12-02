# Sentry Plugin for Mach Composer 

This repository contains the Sentry plugin for Mach Composer. It requires Mach Composer 3.x


## Usage

```yaml
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
    components:
      - name: my-component
        # ...
        sentry:
          project: "component project" # override default
```
