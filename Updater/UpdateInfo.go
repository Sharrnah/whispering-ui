package Updater

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"io"
	"log"
	"net/http"
	"os"
)

/* example yaml
packages:
  app:
    version: 1.0.0.16
    locationUrls:
      DEFAULT:
        - https://eu2.someurl.com/app1.0.0.16_win.zip
    SHA256: 0

  data:
    version: 1.0.0.3
    locationUrls:
      EU:
        - https://eu2.someurl.com/data1.0.0.3_win.zip
      US:
        - https://usc1.someurl.com/data1.0.0.3_win.zip
    SHA256: 0
*/

type UpdateInfo struct {
	Version      string              `yaml:"version"`
	LocationUrls map[string][]string `yaml:"locationUrls"`
	SHA256       string              `yaml:"SHA256"`
}

func (i *UpdateInfo) WriteYaml(fileName string) {
	// marshal the struct to yaml and save as file
	yamlFile, err := yaml.Marshal(i)
	if err != nil {
		log.Printf("error: %v", err)
	}
	err = os.WriteFile(fileName, yamlFile, 0644)
	if err != nil {
		log.Printf("error: %v", err)
	}
}

func (i *UpdateInfo) ReadYaml(data []byte) error {
	// read the yaml file and unmarshal it to the struct
	var err error
	err = yaml.Unmarshal(data, i)
	if err != nil {
		log.Printf("Unmarshal: %v", err)
	}
	return err
}

type UpdatePackages struct {
	Packages map[string]UpdateInfo `yaml:"packages"`
	//DoNotAskAgain bool                  `yaml:"doNotAskAgain,omitempty"`
}

func (u *UpdatePackages) parsePackagesFromYaml(data []byte) error {
	var err error
	err = yaml.Unmarshal(data, u)
	if err != nil {
		log.Printf("Unmarshal: %v", err)
	}
	return err
}

func (u *UpdatePackages) getYaml(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return []byte{}, fmt.Errorf("GET error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return []byte{}, fmt.Errorf("status error: %v", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, fmt.Errorf("read body: %v", err)
	}

	return data, nil
}

func (u *UpdatePackages) WriteYaml(fileName string) {
	// marshal the struct to yaml and save as file
	yamlFile, err := yaml.Marshal(u)
	if err != nil {
		log.Printf("error: %v", err)
	}
	err = os.WriteFile(fileName, yamlFile, 0644)
	if err != nil {
		log.Printf("error: %v", err)
	}
}

func (u *UpdatePackages) GetUpdateInfo(url string) error {
	data, err := u.getYaml(url)
	if err != nil {
		return err
	}
	return u.parsePackagesFromYaml(data)
}
