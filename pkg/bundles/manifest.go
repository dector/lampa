package bundles

import (
	"fmt"
	"lampa/pkg/bundles/model"
	"lampa/pkg/bundles/parsers"
)

type Manifest struct {
	Package string
	Label   string

	VersionCode string
	VersionName string

	MinSdkVersion     string
	TargetSdkVersion  string
	CompileSdkVersion string
}

func LoadManifest(aabFile string) (Manifest, error) {
	content, err := ReadFileFromAab(
		aabFile,
		"base/manifest/AndroidManifest.xml",
	)
	if err != nil {
		return Manifest{}, err
	}

	xml, err := parsers.ParseXml(content)
	if err != nil {
		return Manifest{}, err
	}

	manifest, err := decodeManifest(xml)
	if err != nil {
		return Manifest{}, err
	}

	return manifest, nil
}

func decodeManifest(xml *model.XmlNode) (Manifest, error) {
	m := Manifest{}

	root := xml.GetElement()

	// Verify that it's correct manifest
	rootName := root.GetName()
	if rootName != "manifest" {
		return m, fmt.Errorf("invalid root element: %s, expecting manifest", rootName)
	}

	// Parse top-level attributes
	for _, attr := range root.GetAttribute() {
		switch attr.GetName() {
		case "package":
			m.Package = attr.GetValue()
		case "versionCode":
			m.VersionCode = attr.GetValue()
		case "versionName":
			m.VersionName = attr.GetValue()
		case "compileSdkVersion":
			m.CompileSdkVersion = attr.GetValue()
		}
	}

	// Parse other tags
	for _, c := range root.GetChild() {
		e := c.GetElement()
		switch e.GetName() {
		case "uses-sdk":
			for _, attr := range e.GetAttribute() {
				switch attr.GetName() {
				case "minSdkVersion":
					m.MinSdkVersion = attr.GetValue()
				case "targetSdkVersion":
					m.TargetSdkVersion = attr.GetValue()
				}
			}
		case "application":
			for _, attr := range e.GetAttribute() {
				switch attr.GetName() {
				case "label":
					m.Label = attr.GetValue()
				}
			}
		}
	}

	return m, nil
}
