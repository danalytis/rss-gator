package config

import (
	"encoding/json"
	"fmt"
	"os"
)

const configFileName = ".gatorconfig.json"

type Config struct {
	Db_url            string `json:"db_url"`
	Current_user_name string `json:"current_user_name"`
}

func Read() (Config, error) {
	path, err := getConfigFilePath()
	if err != nil {
		return Config{}, fmt.Errorf("error: %s", err)
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		defaultConfig := Config{
			Db_url:            "postgres://postgres:@localhost:5432/gator?sslmode=disable",
			Current_user_name: "",
		}
		if err := write(defaultConfig); err != nil {
			fmt.Printf("Warning: Could not create default config file: %v\n", err)
		} else {
			fmt.Printf("Created default config file at %s\n", path)
		}
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("error: could not read config file - %v\n", err)
	}

	usrConfig := Config{}
	err = json.Unmarshal(data, &usrConfig)
	if err != nil {
		return Config{}, fmt.Errorf("error: could not unmarshal JSON: %w\nRaw data: %s\n", string(data), err)
	}

	return usrConfig, nil
}

func SetUser(cfg Config, userName string) error {
	cfg.Current_user_name = userName
	err := write(cfg)
	if err != nil {
		return fmt.Errorf("error setting user: %w", err)
	}
	return nil
}

func getConfigFilePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Errorf("error: could not find home directory - %v\n", err)
	}

	return home + "/" + configFileName, nil
}

func write(cfg Config) error {
	data, err := json.Marshal(&cfg)
	if err != nil {
		fmt.Errorf("error: could not marshal config: %v\n", err)
	}

	path, err := getConfigFilePath()
	if err != nil {
		fmt.Errorf("error: failed to get config file path: %v\n", err)
	}
	err = os.WriteFile(path, data, 0600)
	return err
}
