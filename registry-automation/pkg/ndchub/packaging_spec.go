package ndchub

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/hasura/ndc-hub/registry-automation/pkg"
	"gopkg.in/yaml.v3"
)

type ConnectorMetadataDefinition struct {
	Version                       *Version                        `yaml:"version,omitempty" json:"version,omitempty"`
	NDCSpecGeneration             *NDCSpecGeneration              `json:"ndcSpecGeneration,omitempty" yaml:"ndcSpecGeneration,omitempty"`
	PackagingDefinition           PackagingDefinition             `json:"packagingDefinition" yaml:"packagingDefinition"`
	SupportedEnvironmentVariables []EnvironmentVariableDefinition `json:"supportedEnvironmentVariables" yaml:"supportedEnvironmentVariables"`
	Commands                      Commands                        `json:"commands" yaml:"commands"`
	CliPlugin                     *CliPluginDefinition            `json:"cliPlugin,omitempty" yaml:"cliPlugin,omitempty"`
	// Docker compose watch yaml definition types
	DockerComposeWatch        []interface{}              `json:"dockerComposeWatch" yaml:"dockerComposeWatch"`
	NativeToolchainDefinition *NativeToolchainDefinition `json:"nativeToolchainDefinition,omitempty" yaml:"nativeToolchainDefinition,omitempty"`

	// https://github.com/hasura/ndc-hub/blob/main/rfcs/0007-packaging-documentation-page.md
	DocumentationPage *string `json:"documentationPage,omitempty" yaml:"documentationPage,omitempty"`

	// metadata for the connector, not a part of the packaging spec
	Namespace string `json:"-" yaml:"-"`
	Name 	string `json:"-" yaml:"-"`
	VersionStr string `json:"-" yaml:"-"`
}

type PackagingType string

const (
	PrebuiltDockerImage PackagingType = "PrebuiltDockerImage"
	ManagedDockerBuild  PackagingType = "ManagedDockerBuild"
)

type Version string

const (
	V1 Version = "v1"
	V2 Version = "v2"
)

type NDCSpecGeneration string

const (
	V01 NDCSpecGeneration = "v0.1"
	V02 NDCSpecGeneration = "v0.2"
)

type NativeToolchainDefinition struct {
	Commands NativeToolchainCommands `yaml:"commands" json:"commands"`
}

type NativeToolchainCommands struct {
	Start                Command  `json:"start" yaml:"start"`
	Watch                *Command `json:"watch,omitempty" yaml:"watch,omitempty"`
	Update               *Command `json:"update,omitempty" yaml:"update,omitempty"`
	UpgradeConfiguration *Command `json:"upgradeConfiguration,omitempty" yaml:"upgradeConfiguration,omitempty"`
	CLIPluginEntrypoint  *Command `json:"cliPluginEntrypoint,omitempty" yaml:"cliPluginEntrypoint,omitempty"`
}

type PackagingDefinition struct {
	Type        PackagingType `json:"type" yaml:"type"`
	DockerImage *string       `json:"dockerImage,omitempty" yaml:"dockerImage,omitempty"`
}

func (p PackagingDefinition) GetDockerImage() string {
	if p.Type == PrebuiltDockerImage {
		return *p.DockerImage
	}
	return ""
}

type EnvironmentVariableDefinition struct {
	Name         string  `json:"name" yaml:"name"`
	Description  string  `json:"description" yaml:"description"`
	DefaultValue *string `json:"defaultValue,omitempty" yaml:"defaultValue,omitempty"`

	// https://github.com/hasura/ndc-hub/blob/main/rfcs/0009-mandatory-env-vars.md
	Required *bool `json:"required,omitempty" yaml:"required,omitempty"`
}

type Commands struct {
	Update                     *Command `json:"update,omitempty" yaml:"update,omitempty"`
	Watch                      *Command `json:"watch,omitempty" yaml:"watch,omitempty"`
	PrintSchemaAndCapabilities *Command `json:"printSchemaAndCapabilities,omitempty" yaml:"printSchemaAndCapabilities,omitempty"`
	UpgradeConfiguration       *Command `json:"upgradeConfiguration,omitempty" yaml:"upgradeConfiguration,omitempty"`
}

func (c *Commands) GetDockerImages() []string {
	dockerImageSet := make(map[string]bool)
	if c.Update != nil {
		if img := c.Update.GetDockerImage(); img != "" {
			dockerImageSet[img] = true
		}
	}
	if c.Watch != nil {
		if img := c.Watch.GetDockerImage(); img != "" {
			dockerImageSet[img] = true
		}
	}
	if c.PrintSchemaAndCapabilities != nil {
		if img := c.PrintSchemaAndCapabilities.GetDockerImage(); img != "" {
			dockerImageSet[img] = true
		}
	}
	if c.UpgradeConfiguration != nil {
		if img := c.UpgradeConfiguration.GetDockerImage(); img != "" {
			dockerImageSet[img] = true
		}
	}
	dockerImages := make([]string, 0, len(dockerImageSet))
	for img := range dockerImageSet {
		dockerImages = append(dockerImages, img)
	}
	return dockerImages
}

type CliPluginType string

const (
	BinaryPluginType       CliPluginType = "Binary"
	DockerPluginType       CliPluginType = "Docker"
	BinaryInlinePluginType CliPluginType = "BinaryInline"
)

type CliPluginDefinition struct {
	Binary *BinaryCliPluginDefinition
	Docker *DockerCliPluginDefinition
}

func (c *CliPluginDefinition) GetDockerImage() string {
	if c.Docker != nil {
		return c.Docker.GetDockerImage()
	}
	return "";
}

type BinaryCliPluginDefinition struct {
	External *BinaryExternalCliPluginDefinition
	Inline   *BinaryInlineCliPluginDefinition
}

type BinaryInlineCliPluginDefinition struct {
	Type      CliPluginType             `json:"type" yaml:"type"`
	Platforms []BinaryCliPluginPlatform `json:"platforms" yaml:"platforms"`
}

type PlatformSelector string

const (
	PlatformDarwinArm64  PlatformSelector = "darwin-arm64"
	PlatformLinuxArm64   PlatformSelector = "linux-arm64"
	PlatformDarwinAmd64  PlatformSelector = "darwin-amd64"
	PlatformWindowsAmd64 PlatformSelector = "windows-amd64"
	PlatformLinuxAmd64   PlatformSelector = "linux-amd64"
)

type BinaryCliPluginPlatform struct {
	Selector PlatformSelector `json:"selector" yaml:"selector"`
	URI      string           `json:"uri" yaml:"uri"`
	SHA256   string           `json:"sha256" yaml:"sha256"`
	Bin      string           `json:"bin" yaml:"bin"`
}

type BinaryExternalCliPluginDefinition struct {
	Type    *CliPluginType `json:"type" yaml:"type"`
	Name    string         `json:"name" yaml:"name"`
	Version string         `json:"version" yaml:"version"`
}

type DockerCliPluginDefinition struct {
	Type        CliPluginType `json:"type" yaml:"type"`
	DockerImage string        `json:"dockerImage" yaml:"dockerImage"`
}

func (d *DockerCliPluginDefinition) GetDockerImage() string {
	return d.DockerImage
}

type CommandType string

const (
	DockerizedCommandType  CommandType = "Dockerized"
	ShellScriptCommandType CommandType = "ShellScript"
)

type DockerizedCommand struct {
	Type        CommandType `json:"type" yaml:"type"`
	DockerImage string      `json:"dockerImage" yaml:"dockerImage"`
	CommandArgs []string    `json:"commandArgs" yaml:"commandArgs"`
}

func (d *DockerizedCommand) GetDockerImage() string {
	if d.Type == DockerizedCommandType {
		return d.DockerImage
	}
	return ""
}

type ShellScriptCommand struct {
	Type       CommandType `json:"type" yaml:"type"`
	Bash       string      `json:"bash" yaml:"bash"`
	Powershell string      `json:"powershell" yaml:"powershell"`
}

type Command struct {
	String             *string
	DockerizedCommand  *DockerizedCommand
	ShellScriptCommand *ShellScriptCommand
}

func (c *Command) GetDockerImage() string {
	if c.DockerizedCommand != nil {
		return c.DockerizedCommand.GetDockerImage()
	}
	return ""
}

func (x *Command) UnmarshalYAML(n *yaml.Node) error {
	if n.Kind == yaml.ScalarNode {
		if n.Tag == "!!str" {
			x.String = &n.Value
			return nil
		}
		return &yaml.TypeError{
			Errors: []string{fmt.Sprintf("unmarshaling Command: %v is not a string command", n.Value)},
		}
	}
	var cmd map[string]interface{}
	if err := n.Decode(&cmd); err != nil {
		return &yaml.TypeError{
			Errors: []string{fmt.Sprintf("unmarshaling Command: %v is not a docker or shell command: %s", n.Value,
				err.Error())},
		}
	}
	cmdType, ok := cmd["type"]
	if !ok {
		return &yaml.TypeError{
			Errors: []string{fmt.Sprintf("unmarshaling Command: %v does not have a type", n.Value)},
		}
	}
	switch CommandType(cmdType.(string)) {
	case DockerizedCommandType:
		return n.Decode(&x.DockerizedCommand)
	case ShellScriptCommandType:
		return n.Decode(&x.ShellScriptCommand)
	}
	return &yaml.TypeError{
		Errors: []string{fmt.Sprintf("unmarshaling Command: %v is not a docker or shell command", cmdType)},
	}
}

func (x Command) MarshalYAML() (interface{}, error) {
	if x.String != nil {
		return *x.String, nil
	}
	if x.DockerizedCommand != nil {
		return *x.DockerizedCommand, nil
	}
	if x.ShellScriptCommand != nil {
		return *x.ShellScriptCommand, nil
	}
	return nil, fmt.Errorf("marshaling Command: no field found to marshal")
}

func (x *CliPluginDefinition) UnmarshalYAML(n *yaml.Node) error {
	type union struct {
		Type *string `yaml:"type"`
	}
	var u union
	if err := n.Decode(&u); err != nil {
		return err
	}
	if u.Type == nil {
		var binary BinaryCliPluginDefinition
		if err := n.Decode(&binary); err != nil {
			return err
		}
		x.Binary = &binary
	} else if *u.Type == string(DockerPluginType) {
		var docker DockerCliPluginDefinition
		if err := n.Decode(&docker); err != nil {
			return err
		}
		x.Docker = &docker
	} else {
		var binary BinaryCliPluginDefinition
		if err := n.Decode(&binary); err != nil {
			return err
		}
		x.Binary = &binary
	}
	return nil
}

func (x CliPluginDefinition) MarshalYAML() (interface{}, error) {
	if x.Docker != nil {
		return *x.Docker, nil
	}
	if x.Binary != nil {
		return *x.Binary, nil
	}
	return nil, fmt.Errorf("marshaling CliPluginDefinition: no field found to marshal")
}

func (x *BinaryCliPluginDefinition) UnmarshalYAML(n *yaml.Node) error {
	type union struct {
		Type      *string                    `yaml:"type"`
		Name      *string                    `yaml:"name"`
		Version   *string                    `yaml:"version"`
		Platforms *[]BinaryCliPluginPlatform `yaml:"platforms"`
	}
	var u union
	if err := n.Decode(&u); err != nil {
		return err
	}
	if u.Type == nil {
		x.External = &BinaryExternalCliPluginDefinition{
			Type:    nil,
			Name:    *u.Name,
			Version: *u.Version,
		}
	} else if *u.Type == string(BinaryPluginType) {
		x.External = &BinaryExternalCliPluginDefinition{
			Type:    (*CliPluginType)(u.Type),
			Name:    *u.Name,
			Version: *u.Version,
		}
	} else if *u.Type == string(BinaryInlinePluginType) {
		x.Inline = &BinaryInlineCliPluginDefinition{
			Type:      (CliPluginType)(*u.Type),
			Platforms: *u.Platforms,
		}
	} else {
		return fmt.Errorf("type %q is invalid", *u.Type)
	}
	return nil
}

func (x BinaryCliPluginDefinition) MarshalYAML() (interface{}, error) {
	if x.External != nil {
		return *x.External, nil
	}
	if x.Inline != nil {
		return *x.Inline, nil
	}
	return nil, fmt.Errorf("marshaling BinaryCliPluginDefinition: no field found to marshal")
}

func (def *ConnectorMetadataDefinition) Validate() error {
	if def.Version != nil && *def.Version == V2 {
		// Must contain ndc spec generation
		if def.NDCSpecGeneration == nil || *def.NDCSpecGeneration == "" {
			return fmt.Errorf("packaging spec v2 must contain ndc spec generation")
		}
		// If CLI plugin is specified, it must not be binary external
		if def.CliPlugin != nil {
			if def.CliPlugin.Binary != nil && def.CliPlugin.Binary.External != nil {
				return fmt.Errorf("packaging spec v2 must not contain binary external cli plugin. " +
					"Can only be binary inline or docker")
			}
		}
	}
	if def.Version == nil || *def.Version == V1 {
		// If CLI plugin is specified, it must not be binary inline
		if def.CliPlugin != nil {
			if def.CliPlugin.Binary != nil && def.CliPlugin.Binary.Inline != nil {
				return fmt.Errorf("packaging spec v1 must not contain binary inline cli plugin. " +
					"Can only be binary external or docker")
			}
		}
	}
	return nil
}



// GetArtifacts will return the artifacts for the connector
// It will have a list of all Docker images that the connector uses (if any). There can be multiple because there might be a connector image and a plugin image.
//
// arg artifactsPath is the path where the artifacts will be downloaded. If it is not provided, a random path will be generated.
func (def *ConnectorMetadataDefinition) GetArtifacts(artifactsPath string) (*ConnectorArtifacts, error) {
	// Get the docker images
	dockerImages := def.GetDockerImages()

	// Get the plugin dir path
	artifactsDirPath := ""

	if artifactsPath != "" {
		artifactsDirPath = artifactsPath
	} else {
		artifactPath, err := getTempArtifactsPath(def.Namespace, def.Name, string(*def.Version))
		if err != nil {
			return nil, fmt.Errorf("failed to get temp artifacts path: %v", err)
		}
		artifactsDirPath = artifactPath
	}

	// download CLI plugins
	err := DownloadPluginBinaries(artifactsDirPath, WithConnectorMetadata(def))

	return &ConnectorArtifacts{
		DockerImages:  dockerImages,
		ArtifactsDirPath: artifactsDirPath,
	}, err
}

func getTempArtifactsPath(namespace, connectorName, version string) (string, error) {
	// Get the present working directory
	pwd := os.Getenv("PWD")
	if pwd == "" {
		return "", fmt.Errorf("failed to get the present working directory")
	}
	// Create a temp path for the artifacts
	tempPath := filepath.Join(pwd, "artifacts", namespace, connectorName, version)
	if _, err := os.Stat(tempPath); os.IsNotExist(err) {
		err := os.MkdirAll(tempPath, 0755)
		if err != nil {
			return "", fmt.Errorf("failed to create temp path for artifacts: %v",	 err)
		}
	}
	return tempPath, nil

}

// GetDockerImages will return the docker images for the connector (if any)
// There can be multiple because there might be a connector image and a plugin image.
func (def *ConnectorMetadataDefinition) GetDockerImages() []string {
	// fmt.Printf("GetDockerImages: %v\n", def)
	dockerImages := make(map[string]bool) // map to avoid duplicates
	if def.PackagingDefinition.GetDockerImage() != "" {
		dockerImages[def.PackagingDefinition.GetDockerImage()] = true
	}

	if def.CliPlugin !=  nil && def.CliPlugin.GetDockerImage() != "" {
		dockerImages[def.CliPlugin.GetDockerImage()] = true
	}

	for _, img := range def.Commands.GetDockerImages() {
		dockerImages[img] = true
	}

	dockerImgArr := make([]string, 0, len(dockerImages))
	for k := range dockerImages {
    	dockerImgArr = append(dockerImgArr, k)
	}
	return dockerImgArr
}

func GetPackagingSpec(uri, namespace, name, version string) (connectorMetadataDefinition *ConnectorMetadataDefinition, tgzPath string, extractedTgzPath string, err error) {
	def, tgzPath, extractedTgzPath, err := pkg.GetConnectorVersionMetadata(uri, namespace, name, version)
	if err != nil {
		return nil, "", "", err
	}
	defBytes, err := yaml.Marshal(def)
	if err != nil {
		return nil, "", "", err
	}
	var spec ConnectorMetadataDefinition
	err = yaml.Unmarshal(defBytes, &spec)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to unmarshal connector-metadata.yaml: %v", err)
	}

	spec.Namespace = namespace
	spec.Name = name
	spec.VersionStr = version

	return &spec, tgzPath, extractedTgzPath, nil
}

type ConnectorArtifacts struct {
	DockerImages []string
	ArtifactsDirPath string
}
