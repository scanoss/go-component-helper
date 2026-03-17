# SCANOSS Platform 2.0 Component Helper Package
Welcome to the SCANOSS Platform 2.0 component helper package.

This package contains helper functions to make development of Go services easier to configure for component version resolution.

## Usage
The main function in this package is `GetComponentsVersion`. It takes a list of components (with PURLs and optional requirements), resolves their concrete versions using the SCANOSS API, and returns the results.

```go
import (
    componenthelper "github.com/scanoss/go-component-helper/componenthelper"
)

results := componenthelper.GetComponentsVersion(componenthelper.ComponentVersionCfg{
    MaxWorkers: 5,
    Ctx:        ctx,
    S:          logger,
    DB:         db,
    Input: []componenthelper.ComponentDTO{
        {Purl: "pkg:npm/lodash", Requirement: ">=4.17.0"},
        {Purl: "pkg:github/scanoss/scanner.c@1.2.3"},
    },
})
```

### PURL Version Handling
When a PURL contains a version (e.g., `pkg:github/scanoss/scanner.c@1.2.3`), the version is automatically extracted and moved to the `Requirement` field. **This overwrites any existing requirement.** The PURL is then stored without the version (e.g., `pkg:github/scanoss/scanner.c`).

This means the following inputs are equivalent:
- `{Purl: "pkg:npm/lodash@4.17.0"}`
- `{Purl: "pkg:npm/lodash", Requirement: "4.17.0"}`

Qualifiers and subpaths in the PURL are preserved (e.g., `pkg:npm/%40scope/name@1.0.0?repository_url=https://example.com` becomes `pkg:npm/%40scope/name?repository_url=https://example.com` with Requirement `1.0.0`).

More details about each function can be found in the packaged documentation.

## Bugs/Features
To request features or alert about bugs, please do so [here](https://github.com/scanoss/go-component-helper/issues).

## Changelog
Details of major changes to the library can be found in [CHANGELOG.md](CHANGELOG.md).