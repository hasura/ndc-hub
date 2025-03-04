package ndchub

import (
	"encoding/json"
	"os"
)

type TestConfig struct {
	Path string `json:"-"`

	HubID                string   `json:"hub_id"`
	Port                 *int     `json:"port,omitempty"`
	Envs                 []string `json:"envs,omitempty"`
	SetupComposeFilePath *string  `json:"setup_compose_file_path,omitempty"`
	RunCloudTests        *bool    `json:"run_cloud_tests,omitempty"`
	SnapshotsDir         string   `json:"snapshots_dir"`
}

func GetTestConfig(path string) (*TestConfig, error) {
	tcBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var testConfig TestConfig
	if err := json.Unmarshal(tcBytes, &testConfig); err != nil {
		return nil, err
	}
	testConfig.Path = path

	return &testConfig, nil
}
