# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---


## [0.3.3-snapshot] - Unreleased

[DIFF v0.3.2 <-> v0.3.3](https://github.com/dector/lampa/compare/v0.3.2...v0.3.3)


## [0.3.2] - 2025-07-10


### Added

- [CHANGELOG.md](CHANGELOG.md) file (#32)

### Changed

- Changed Homebrew tap name to `dector/lampa`.
Now it's a part of another repo: [dector/homebrew-lampa](https://github.com/dector/homebrew-lampa).
- Using `git-go` library instead of system `git`.

### Refactored

- Reports generation.

[DIFF v0.3.0 <-> v0.3.2](https://github.com/dector/lampa/compare/v0.3.0...v0.3.2)


> [!NOTE]
> `0.3.1` was skipped accidentally.


## [0.3.0] - 2025-07-06

### Changed
- **BREAKING**: Use stdout for errors on CI environment
- **BREAKING**: Using internal mechanism to parse aab data (#24)
- Dropped bundletool and aapt2 as dependencies (#24)

[DIFF v0.2.0-01 <-> v0.3.0](https://github.com/dector/lampa/compare/v0.2.0-01...v0.3.0)


## [0.2.0-01] - 2025-07-06

### Added
- Enable INFO level with verbosity flag `-v` (#13)

### Changed
- INFO verbosity is enabled on CI by default

[DIFF v0.1.0-5.dev <-> v0.2.0-01](https://github.com/dector/lampa/compare/v0.1.0-5.dev...v0.2.0-01)


## [0.1.0-5.dev] - 2025-07-06

### Added
- Add `version --short` to display only version tag (#34)
- Publish checksums file on release (#33)

### Changed
- **BREAKING**: Release artefacts now contains version in file name (#34)

[DIFF v0.1.0-4.dev <-> v0.1.0-5.dev](https://github.com/dector/lampa/compare/v0.1.0-4.dev...v0.1.0-5.dev)


## [0.1.0-4.dev] - 2025-07-06

Introduces first public version of json report.

### Added
- **BREAKING**: `--format` flag to specify report type
- Download Bundletool automatically (#19)

### Removed
- **BREAKING** `--with-html` is gone. Use `--format` instead.

[DIFF v0.1.0-3.dev <-> v0.1.0-4.dev](https://github.com/dector/lampa/compare/v0.1.0-3.dev...v0.1.0-4.dev)


## [0.1.0-3.dev] - 2025-07-06

### Added
- Build and analyze AAB release instead of APK.

### Fixed
- Create report dir if it doesn't exist (#30)
- Display correct release version

[DIFF v0.1.0-2.dev <-> v0.1.0-3.dev](https://github.com/dector/lampa/compare/v0.1.0-2.dev...v0.1.0-3.dev)


## [0.1.0-2.dev] - 2025-07-06

### Added
- Check if Java is present in the system (#23)
- Detect CI mode and don't display spinner (#29)

[DIFF v0.1.0-1.dev <-> v0.1.0-2.dev](https://github.com/dector/lampa/compare/v0.1.0-1.dev...v0.1.0-2.dev)


## [0.1.0-1.dev] - 2025-07-05

### Added
- Formula for brew

### Changed
- Misc minor usability changes

[DIFF v0.1.0-0.dev <-> v0.1.0-1.dev](https://github.com/dector/lampa/compare/v0.1.0-0.dev...v0.1.0-1.dev)


## [0.1.0-0.dev] - 2025-07-04

First version: runs analysis on APK file.

[0.1.0-0.dev]: https://github.com/dector/lampa/releases/tag/v0.1.0-0.dev
