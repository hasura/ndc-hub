package ndchub

import (
	"fmt"
	"log"
	"net/http"
	"path/filepath"

	"github.com/hasura/ndc-hub/registry-automation/pkg"
	"gopkg.in/yaml.v3"
)

type PluginManifest struct {
	Name             string     `json:"name" yaml:"name"`
	Version          string     `json:"version" yaml:"version"`
	ShortDescription string     `json:"shortDescription" yaml:"shortDescription"`
	Homepage         string     `json:"homepage" yaml:"homepage"`
	Hidden           bool       `json:"hidden" yaml:"hidden"`
	Platforms        []Platform `json:"platforms" yaml:"platforms"`
}

type Platform struct {
	Selector string     `json:"selector" yaml:"selector"`
	URI      string     `json:"uri" yaml:"uri"`
	SHA256   string     `json:"sha256" yaml:"sha256"`
	Bin      string     `json:"bin" yaml:"bin"`
	Files    []FilePair `json:"files" yaml:"files"`
}

type FilePair struct {
	From string `json:"from" yaml:"from"`
	To   string `json:"to" yaml:"to"`
}

type ManifestDownloadOptions struct {
    Name           string
    Version             string
    ConnectorMetadata   *ConnectorMetadataDefinition
}

type ManifestOption func(*ManifestDownloadOptions)

func WithNameAndVersion(name, version string) ManifestOption {
    return func(opt *ManifestDownloadOptions) {
        opt.Name = name
        opt.Version = version
    }
}

func WithConnectorMetadata(md *ConnectorMetadataDefinition) ManifestOption {
    return func(opt *ManifestDownloadOptions) {
        opt.ConnectorMetadata = md
    }
}

func DownloadPluginBinaries(artifactsDirPath string, opts ...ManifestOption) {
	manifest, err := DownloadPluginsManifest(opts...)
	if err != nil {
		log.Fatalf("Failed to download plugins manifest: %v", err)
	}
	if manifest == nil {
		return
	}

	for _, platform := range manifest.Platforms {
		fileName := fmt.Sprintf("%s-%s", platform.Selector, platform.Bin)
		filePath := filepath.Join(artifactsDirPath, fileName)

		if err := pkg.DownloadFile(platform.URI, filePath, map[string]string{"sha256": platform.SHA256}); err != nil {
			log.Fatalf("Failed to download plugin binary for %s: %v", platform.Selector, err)
		}
	}
}

func DownloadPluginsManifest(opts ...ManifestOption) (*PluginManifest, error) {
    var options ManifestDownloadOptions

    // Apply options
    for _, opt := range opts {
        opt(&options)
    }

    switch {
    case options.ConnectorMetadata != nil:
        return downloadPluginsManifestWithConnectorMetadata(options.ConnectorMetadata)

    case options.Name != "" && options.Version != "":
        return downloadPluginsManifestWithNameAndVersion(options.Name, options.Version)

    default:
        return nil, fmt.Errorf("insufficient parameters provided to DownloadPluginsManifest")
    }
}

func downloadPluginsManifestWithConnectorMetadata(md *ConnectorMetadataDefinition) (*PluginManifest, error) {
	if md == nil {
		return nil, fmt.Errorf("error downloading plugins manifest: connector metadata cannot be nil")
	}
	if md.CliPlugin == nil {
		log.Printf("No CLI Plugin found for %s/%s:%s. Skipping download", md.Namespace, md.Name, md.VersionStr)
		return nil, nil
	}
	if md.CliPlugin.Docker != nil && md.CliPlugin.Docker.DockerImage != "" {
		log.Printf("CLI Plugin for %s/%s:%s is a Docker Image, skipping binary download", md.Namespace, md.Name, md.VersionStr)
		return nil, nil
	}
	return downloadPluginsManifestWithNameAndVersion(md.CliPlugin.Binary.External.Name, md.CliPlugin.Binary.External.Version)
}

func downloadPluginsManifestWithNameAndVersion(name, version string) (*PluginManifest, error) {
	if name == "" {
		return nil, fmt.Errorf("error downloading plugins manifest: name cannot be empty")
	}
	if version == "" {
		return nil, fmt.Errorf("error downloading plugins manifest: version cannot be empty")
	}

	manifestURL := getManifestUrl(name, version)
	log.Printf("Downloading manifest from %s", manifestURL)
	resp, err := http.Get(manifestURL)
	if err != nil {
		return nil, fmt.Errorf("failed to download manifest from %s: %v", manifestURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download manifest: %s", resp.Status)
	}

	var manifest PluginManifest
	if err := yaml.NewDecoder(resp.Body).Decode(&manifest); err != nil {
		return nil, fmt.Errorf("failed to decode manifest: %v", err)
	}

	return &manifest, nil
}

func getManifestUrl(name, version string) string {
	return fmt.Sprintf("https://raw.githubusercontent.com/hasura/cli-plugins-index/refs/heads/master/plugins/%s/%s/manifest.yaml", name, version)
}