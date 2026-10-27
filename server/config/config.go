package config

import (
	"errors"
	"io/fs"
	"log"
	"os"
	"sync"

	"gopkg.in/yaml.v3"
)

var (
	instance Config
	once     sync.Once
)

func GetInstance() *Config {
	once.Do(initialize)

	return &instance
}

func initialize() {
	root, err := readByFile()
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			create()
		}

		if root, err = defaultRoot(); err != nil {
			log.Fatalf("Failed to read default configuration!")
		}

		log.Println("using default configuration")
	}

	if root, err = validate(root); err != nil {
		log.Fatalf("Failed to validate configuration!")
	}

	if err := decodeConfig(root, &instance); err != nil {
		log.Fatalf("Failed to parse configuration: %s", err)
	}

	checkDefaultValue()

	if instance.Authentication == "disable" {
		log.Println("NOTICE: Authentication is disabled! Please ensure your service is secure!")
	}

	log.Println("config loaded successfully")
}

func readByFile() (*yaml.Node, error) {
	data, err := os.ReadFile(ConfigurationFile)
	if err != nil {
		return nil, err
	}

	return parseConfig(data)
}

// Create configuration file.
func create() {
	var (
		file *os.File
		data []byte
		err  error
	)

	// A directory needs its execute bit, otherwise nothing inside it can be
	// opened by path.
	_ = os.MkdirAll("/etc/kvm", 0o755)

	file, err = os.OpenFile("/etc/kvm/server.yaml", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		log.Printf("open config failed: %s", err)
		return
	}
	defer func() {
		_ = file.Close()
	}()

	if data, err = yaml.Marshal(defaultConfig); err != nil {
		log.Printf("failed to marshal default config: %s", err)
		return
	}

	if _, err = file.Write(data); err != nil {
		log.Printf("failed to save config: %s", err)
		return
	}

	if err = file.Sync(); err != nil {
		log.Printf("failed to sync config: %s", err)
		return
	}

	log.Println("create file /etc/kvm/server.yaml with default configuration")
}

// Validate the configuration. This is to ensure compatibility with earlier versions.
func validate(root *yaml.Node) (*yaml.Node, error) {
	if hasPorts(root) {
		return root, nil
	}

	_ = os.Remove(ConfigurationFile)
	log.Println("delete empty configuration file")

	create()

	return defaultRoot()
}
