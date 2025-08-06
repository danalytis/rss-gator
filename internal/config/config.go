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

func Read() Config {
	path, err := getConfigFilePath()
	if err != nil {
		fmt.Errorf("error: %s", err)
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
		fmt.Errorf("error: could not read config file")
	}

	usrConfig := Config{}
	err = json.Unmarshal(data, &usrConfig)
	if err != nil {
		fmt.Errorf("error: could not unmarshal JSON: ", err)
		fmt.Errorf("Raw data:", string(data))
	}

	return usrConfig
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
		fmt.Errorf("error: could not find home directory")
	}

	return home + "/" + configFileName, nil
}

func write(cfg Config) error {

	data, err := json.Marshal(&cfg)
	if err != nil {
		fmt.Errorf("error: could not marshal config: ", err)
	}

	path, err := getConfigFilePath()
	if err != nil {
		fmt.Errorf("error: failed to get config file path", err)
	}

	err = os.WriteFile(path, data, 0644)
	return err
}
