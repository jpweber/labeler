package configReader

import (
	"io/ioutil"
	"log"
	"os"

	yaml "gopkg.in/yaml.v2"
)

type Config struct {
	Excludes  map[string]bool `yaml:"excludes"`
	Namespace string
}

// Read reads info from config file
func Read(configFile string) Config {
	_, err := os.Stat(configFile)
	if err != nil {
		log.Fatal("Config file is missing: ", configFile)
	}
	file, _ := os.Open(configFile)
	fileBytes, err := ioutil.ReadAll(file)
	var config Config
	err = yaml.Unmarshal(fileBytes, &config)
	if err != nil {
		log.Fatalln("config unmarshaling error:", err)
	}

	return config
}
