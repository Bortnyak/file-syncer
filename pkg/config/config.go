package config

import (
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	BucketName      string `yaml:"bucket_name"`
	WorkDir         string `yaml:"work_dir"`
	CloudConfigPath string `yaml:"cloud_config_path"`
}

const GOOGLE_APPLICATION_CREDENTIALS_ENV = "GOOGLE_APPLICATION_CREDENTIALS"
const SYNCER_BUCKET_NAME_ENV = "SYNCER_BUCKET_NAME"
const SYNCER_WORK_DIR_ENV = "SYNCER_WORK_DIR"
const CONFIG_FILE_NAME = "config.yml"

func LoadConfig() (*Config, error) {
	yamlData, err := os.ReadFile(CONFIG_FILE_NAME)
	if err != nil {
		log.Fatal("Error while reading app config file", err)
		return nil, err
	}

	config := Config{}
	yaml.Unmarshal(yamlData, &config)

	// Google cloud reads the creds config from a json file
	config.setConfigToEnv()
	os.Environ()

	return &config, nil
}

func (c *Config) GetCloudConfigPath() string {
	return c.CloudConfigPath
}

func (c *Config) setConfigToEnv() {
	os.Setenv(GOOGLE_APPLICATION_CREDENTIALS_ENV, c.CloudConfigPath)
	os.Setenv(SYNCER_BUCKET_NAME_ENV, c.BucketName)
	os.Setenv(SYNCER_WORK_DIR_ENV, c.WorkDir)
}
