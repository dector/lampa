# Bundle Parser

Package for parsing and extracting metadata from Android App Bundle (AAB) files.

## Overview

This package reads AAB files (ZIP archives) and extracts application metadata including manifest information and resource tables. All data is stored in protobuf format within the AAB.

## Usage

### Load Manifest

Extract app metadata from the manifest:

```go
import "github.com/dector/lampa/pkg/bundles"

manifest, err := bundles.LoadManifest("path/to/app.aab")
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Package: %s\n", manifest.Package)
fmt.Printf("Version: %s (%d)\n", manifest.VersionName, manifest.VersionCode)
fmt.Printf("SDK: min=%d, target=%d\n", manifest.MinSdkVersion, manifest.TargetSdkVersion)
```

### Load Resources

Extract the resource table:

```go
resources, err := bundles.LoadResources("path/to/app.aab")
if err != nil {
    log.Fatal(err)
}

// Iterate through packages, types, and entries to find resources
for _, pkg := range resources.Package {
    for _, typ := range pkg.Type {
        for _, entry := range typ.Entry {
            // Access resource entries
        }
    }
}
```

### Read Arbitrary Files

Extract any file from within the AAB:

```go
data, err := bundles.ReadFileFromAab("path/to/app.aab", "base/dex/classes.dex")
if err != nil {
    log.Fatal(err)
}
```

## Structure

- `manifest.go` - Parse Android manifest from `base/manifest/AndroidManifest.xml`
- `resources.go` - Parse resource table from `base/resources.pb`
- `zip.go` - Utility for reading files from AAB (ZIP archive)
- `xml.go` - Convert protobuf XML to readable text format
- `model/` - Protobuf definitions for Android resources
- `parsers/` - Protobuf unmarshaling utilities

## Data Format

AAB files use binary protobuf encoding for metadata:
- Manifest: protobuf-encoded XML at `base/manifest/AndroidManifest.xml`
- Resources: binary protobuf at `base/resources.pb`

## Contributing

This package is part of the [Lampa](https://github.com/dector/lampa) project. See the main repository for contribution guidelines.

## License

MIT License - see the main Lampa project for details.
