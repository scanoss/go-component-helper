# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]
- Upcoming changes...
### Added
- Added `GetComponentsVersion` function to resolve concrete versions for a list of components using a fan-out/fan-in concurrency pattern
- Added `sanitiseComponents` to validate and normalise component PURLs (empty purl detection, invalid purl handling, semver operator extraction)
- Added `HasSemverOperator` utility to detect semver range operators (`>=`, `<=`, `~`, `>`, `<`) in version strings
- Added `ComponentDTO`, `Component`, and `ComponentVersionCfg` types
- Added unit tests for component sanitisation and semver utilities
- Added GitHub Actions CI workflow
- Added project scaffolding (LICENSE, CODE_OF_CONDUCT, CONTRIBUTING, Makefile, README)

[0.1.0]: https://github.com/scanoss/go-component-helper/tag/v0.1.0
