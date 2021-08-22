package config

import (
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v3"
)

type Member struct {
	FirstName string   `yaml:"first_name"`
	LastName  string   `yaml:"last_name"`
	Subgroup  string   `yaml:"subgroup"`
	UserID    string   `yaml:"userID"`
	Roles     []string `yaml:"roles"`
}

func ParseMembers(membersFile string) (members map[string]Member) {
	yamlMembers, err := ioutil.ReadFile(membersFile)
	if err != nil {
		log.Fatalf("Err: cannot read members file: %v", err)
	}

	err = yaml.Unmarshal(yamlMembers, &members)
	if err != nil {
		log.Fatalf("Err: cannot parse members file: %v", err)
	}
	return members
}

func ParseSecrets(secretsFile string) (secrets map[string]string) {
	yamlSecrets, err := ioutil.ReadFile(secretsFile)
	if err != nil {
		log.Fatalf("Err: cannot read secrets file: %v", err)
	}

	err = yaml.Unmarshal(yamlSecrets, &secrets)
	if err != nil {
		log.Fatalf("Err: cannot parse secrets file: %v", err)
	}
	return secrets
}
