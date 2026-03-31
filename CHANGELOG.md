# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]
### Added
- Added `OriginalPurl` and `OriginalRequirement` fields to `Component` struct to preserve the original values before sanitisation or resolution
- Added `TestSanitiseComponentsOriginalFields` tests to verify original field preservation across all sanitisation paths

## [0.5.0] - 2026-03-19
### Changed
- Upgraded `scanoss/go-purl-helper` to v0.3.0
### Added 
- Added support for `conan` package URL

## [0.4.0] - 2026-03-18
### Fixed
- Fixed PURL version extraction to use proper PURL parsing instead of naive `strings.Split("@")`, correctly handling qualifiers, subpaths, and percent-encoded characters (e.g., `%40` in scoped npm packages)
### Changed
- PURL version now always overwrites the existing requirement when present in the PURL

## [0.3.0] - 2026-03-16
### Added
- Added PURL component properties (`PurlType`, `PurlName`, `PurlNamespace`, `PurlQualifiers`, `PurlSubpath`, `Name`, `URL`) to `Component` struct
- Added `ComponentNameFromString` function to extract the full component name (namespace + name) from a PURL string
### Changed
- Replaced `CheckPurl` + `GetComponent` flow with single `GetComponent` call using sentinel errors (`ErrComponentNotFound`, `ErrVersionNotFound`) for error handling
- Upgraded `go-models` to v0.7.0 

## [0.2.0] - 2026-03-09
### Changed
- Upgraded `go-models` to v0.6.0 to improve component search accuracy

## [0.1.0] - 2026-02-25
### Added
- Added `GetComponentsVersion` function to resolve concrete versions for a list of components using a fan-out/fan-in concurrency pattern
- Added `sanitiseComponents` to validate and normalise component PURLs (empty purl detection, invalid purl handling, semver operator extraction)
- Added `HasSemverOperator` utility to detect semver range operators (`>=`, `<=`, `~`, `>`, `<`) in version strings
- Added `ComponentDTO`, `Component`, and `ComponentVersionCfg` types
- Added unit tests for component sanitisation and semver utilities
- Added GitHub Actions CI workflow
- Added project scaffolding (LICENSE, CODE_OF_CONDUCT, CONTRIBUTING, Makefile, README)

[0.1.0]: https://github.com/scanoss/go-component-helper/tag/v0.1.0
[0.2.0]: https://github.com/scanoss/go-component-helper/compare/v0.1.0...v0.2.0
[0.3.0]: https://github.com/scanoss/go-component-helper/compare/v0.2.0...v0.3.0
[0.4.0]: https://github.com/scanoss/go-component-helper/compare/v0.3.0...v0.4.0
[0.5.0]: https://github.com/scanoss/go-component-helper/compare/v0.4.0...v0.5.0