package bundle

import (
	"archive/zip"
	"fmt"
	"io"
	"lampa/internal/proto"
)

type Manifest struct {
	Package string

	VersionCode string
	VersionName string

	MinSdkVersion     string
	TargetSdkVersion  string
	CompileSdkVersion string
}

func LoadManifest(aabFile string) (Manifest, error) {
	manifestContent, err := LoadRawManifest(aabFile)
	if err != nil {
		return Manifest{}, err
	}
	manifestXml, err := proto.ParseXml(manifestContent)
	if err != nil {
		return Manifest{}, err
	}
	manifest, err := decodeManifest(manifestXml)
	if err != nil {
		return Manifest{}, err
	}

	return manifest, nil
}

func LoadRawManifest(aabFile string) ([]byte, error) {
	zipReader, err := zip.OpenReader(aabFile)
	if err != nil {
		panic(err)
	}
	defer zipReader.Close()

	var manifestData []byte
	for _, file := range zipReader.File {
		if file.Name == "base/manifest/AndroidManifest.xml" {
			rc, err := file.Open()
			if err != nil {
				return nil, err
			}
			defer rc.Close()

			manifestData, err = io.ReadAll(rc)
			if err != nil {
				return nil, err
			}
			break
		}
	}
	if manifestData == nil {
		return nil, fmt.Errorf("AndroidManifest.xml not found")
	}

	return manifestData, nil
}

func decodeManifest(xml *proto.XmlNode) (Manifest, error) {
	m := Manifest{}

	// Print p as XML using a simple loop

	root := xml.GetElement()
	rootName := root.GetName()
	if rootName != "manifest" {
		return m, fmt.Errorf("invalid root element: %s, expecting manifest", rootName)
	}

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
			// case "application":
			// 	for _, attr := range e.GetAttribute() {
			// 		switch attr.GetName() {
			// 		// case "label":
			// 		// 	m.Label = attr.GetValue()
			// 		// case "icon":
			// 		// 	m.Icon = attr.GetValue()
			// 		// case "theme":
			// 		// 	m.Theme = attr.GetValue()
			// 		}
			// 	}
		}
	}

	return m, nil
}
