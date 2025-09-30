package config

import (
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

type Token string
type OaiProviderName string
type OaiBaseUrl string
type OaiApiKey string
type OaiProviderConfig struct {
	BaseUrl OaiBaseUrl `yaml:"base_url"`
	ApiKey  OaiApiKey  `yaml:"api_key"`
}
type OaiModelName string
type OaiModelIdentifier string
type OaiExtraHeaders map[string]string
type OaiExtraBody map[string]any
type OaiModelConfig struct {
	Identifier   OaiModelIdentifier `yaml:"identifier"`
	ExtraHeaders OaiExtraHeaders    `yaml:"extra_headers,omitempty"`
	ExtraBody    OaiExtraBody       `yaml:"extra_body,omitempty"`
}
type Config struct {
	Tokens       []Token                                             `yaml:"tokens"`
	OaiProviders map[OaiProviderName]OaiProviderConfig               `yaml:"oai_providers"`
	OaiModels    map[OaiProviderName]map[OaiModelName]OaiModelConfig `yaml:"oai_models"`
}

var (
	config Config
)

func Load(configFilePath string) error {
	yamlFile, err := os.ReadFile(configFilePath)
	if err != nil {
		// If the file can't be read
		return fmt.Errorf("failed to read file '%s': %w", configFilePath, err)
	}

	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		// If the YAML structure is incorrect
		return fmt.Errorf("failed to parse YAML content from '%s': %w", configFilePath, err)
	}

	log.Println("Loaded config")

	return nil
}

func GetHost() string {
	host := os.Getenv("HOST")
	if host == "" {
		return "127.0.0.1"
	}

	return host
}

func GetPort() int {
	portStr := os.Getenv("PORT")
	if portStr == "" {
		return 4283
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return 4283
	}

	return port
}

func GetTokens() ([]string, error) {
	tokens := make([]string, len(config.Tokens))
	for i, token := range config.Tokens {
		tokens[i] = string(token)
	}

	return tokens, nil
}

func GetOaiProviders() ([]string, error) {
	providers := make([]string, 0, len(config.OaiProviders))
	for providerName := range config.OaiProviders {
		providers = append(providers, string(providerName))
	}
	sort.Strings(providers)

	return providers, nil
}

func GetOaiModels() ([]string, error) {
	models := make([]string, 0)
	for providerName, providerModels := range config.OaiModels {
		for modelName := range providerModels {
			models = append(models, fmt.Sprintf("%s:%s", providerName, modelName))
		}
	}
	sort.Strings(models)

	return models, nil
}

func splitOaiIdentifier(oaiIdentifier string) (OaiProviderName, OaiModelName, error) {
	parts := strings.SplitN(oaiIdentifier, ":", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("malformed oai identifier %q", oaiIdentifier)
	}
	oaiProviderName, oaiModelName := parts[0], parts[1]
	if oaiProviderName == "" || oaiModelName == "" {
		return "", "", fmt.Errorf("malformed oai identifier %q", oaiIdentifier)
	}

	return OaiProviderName(oaiProviderName), OaiModelName(oaiModelName), nil
}

func GetOaiConfigByOaiIdentifier(oaiIdentifier string) (*OaiProviderConfig, *OaiModelConfig, error) {
	oaiProviderName, oaiModelName, err := splitOaiIdentifier(oaiIdentifier)
	if err != nil {
		return nil, nil, err
	}

	// Get provider config
	oaiProviderConfig, ok := config.OaiProviders[oaiProviderName]
	if !ok {
		return nil, nil, fmt.Errorf("provider %q not found", oaiProviderName)
	}
	// Get model config
	oaiModelConfig, ok := config.OaiModels[oaiProviderName][oaiModelName]
	if !ok {
		if _, found := config.OaiModels[oaiProviderName]; !found {
			return nil, nil, fmt.Errorf("provider %q not found in models config", oaiProviderName)
		}
		return nil, nil, fmt.Errorf("model %q not found for provider %q", oaiModelName, oaiProviderName)
	}

	return &oaiProviderConfig, &oaiModelConfig, nil
}
