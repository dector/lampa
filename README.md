> [!WARNING]
> Pre-1.0 software: Lampa is currently in the early stages of development.
>
> :construction: :construction: :construction:
>
> Expect frequent breaking changes (especially in CLI arguments), bugs, suboptimal code, and limited functionality.
>
> But if you're feeling adventurous - feel free to try it, your feedback is highly appreciated!
> Please report any issues you encounter, and feel free to share your ideas in [Discussions](https://github.com/dector/lampa/discussions) tab — though I can't guarantee immediate prioritization.

<p align="center">
  <img width="500" height="500" src="docs/lampa-logo.webp" alt="Lampa logo">
</p>

<p align="center">
    <a href="https://mastodon.online/search?q=from%3A%40dector+%23lampa&type=statuses">Updates on Mastodon</a>
</p>

# Lampa

## Contents

- [What is this](#what-is-this)
- [Getting Started](#getting-started)
  - [Install](#install)
  - [Runtime dependencies](#runtime-dependencies)
- [How To Use](#how-to-use)
  - [Generate JSON report for current version](#generate-json-report-for-current-version)
  - [Generate only HTML report for current version](#generate-only-html-report-for-current-version)
  - [Generate comparative HTML report for two releases](#generate-comparative-html-report-for-two-releases)
  - [GitHub Action](#github-action)
- [Contributing](#contributing)
- [License](#license)

## What is this

Lampa is a small tool that is useful for comparing two releases: it generates
overview reports where you can detect changes to third-party dependencies that are
added to the build.

## Getting Started

### Install

Download latest version from [Releases page](https://github.com/dector/lampa/releases/latest).

**or**

for Linux/MacOS use Homebrew:

``` shell
brew tap dector/lampa https://github.com/dector/lampa
brew install dector/lampa
```

### Runtime dependencies

DX will be improved in the future but currently you need to:

  - Have [Java](https://adoptium.net) installed.
  - Have [Android SDK](https://developer.android.com/studio) installed - for now we need `aapt2` but I have plans to change it in the future.
  - Have [Bundletool](https://github.com/google/bundletool/releases/latest) installed.

## How To Use

All commands are executed inside the root folder of Android project
(unless you explicitly specify path to project).

Remember that you can always use `lampa help` if you forget something.

### Generate JSON report for current version

You will need to use this report for comparative HTML report.

``` shell
export BUNDLETOOL_JAR="~/Apps/bundletool-all-1.18.1.jar"
export ANDROID_SDK_ROOT="~/Apps/AndroidSDK"

lampa collect
```

If program finished successfully - you can find report file
`report.lampa.json` in the project folder.

Be aware that by-default program is not rewriting report if it exists.
But you can opt-in for such behavior explicitly by adding `--overwrite` flag:

``` shell
lampa collect --overwrite
```

Other useful flags are:

  - `--project <project-dir>` - specify path to project root explicitly.
  - `--to-dir <out-dir>` - change the location of the report(s).
  - `--variant <gradle-variant>` - specify custom build variant that you use in Gradle. Might be useful if you have flavors etc.
  - `--format html`/`--format json,html` - if you need only HTML report or both.
  - `--file-name <report-file-name>` - if you need to customize generated report filename (without extension).

[Sample report](http://dector.space/lampa/github/libre-tube/LibreTube/v0.28.1.json).

### Generate only HTML report for current version

``` shell
export BUNDLETOOL_JAR="~/Apps/bundletool-all-1.18.1.jar"
export ANDROID_SDK_ROOT="~/Apps/AndroidSDK"

lampa collect --format html
```

[Sample report](http://dector.space/lampa/github/libre-tube/LibreTube/v0.28.1.html).

### Generate comparative HTML report for two releases

First, you need to generate JSON report for release 1 (e.g. `1.json`).
Then, you need to generate JSON report for release 2 (e.g. `2.json`).

After, you need to generate comparative report with `lampa compare`.

For example:

``` shell
git checkout v0.28.0
lampa collect --to-dir build --file-name v0.28.0

git checkout v0.28.1
lampa collect --to-dir build --file-name v0.28.1

lampa compare build/v0.28.0.json build/v0.28.1.json build/diff.html
```

[Sample report](https://dector.space/lampa/github/libre-tube/LibreTube/v0.28.0..v0.28.1.html).

### GitHub Action

GitHub Action:
```
dector/run-lampa@v1
```

You can use this GitHub Action to integrate Lampa into your CI/CD pipeline.

See detailed instructions on [GitHub Marketplace](https://github.com/marketplace/actions/run-lampa).

[Production-ready example workflow](https://github.com/marketplace/actions/run-lampa#example-workflow)

## Contributing

I will add this section latest. For now feel free to contact me directly or
open new [discussion](https://github.com/dector/lampa/discussions).

## License

Project is distributed under [MIT License](https://opensource.org/license/mit).
