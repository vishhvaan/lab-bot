package config

import (
	"io/ioutil"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

type Member struct {
	FirstName string   `yaml:"first_name"`
	LastName  string   `yaml:"last_name"`
	Subgroup  string   `yaml:"subgroup"`
	UserID    string   `yaml:"userID"`
	Birthday  string   `yaml:"birthday"`
	Roles     []string `yaml:"roles"`
}

func ParseMembers(membersFile string) (members map[string]Member) {
	yamlMembers, err := ioutil.ReadFile(membersFile)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Fatal("Cannot read members file.")
	}

	err = yaml.Unmarshal(yamlMembers, &members)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Fatal("Cannot parse members file.")
	}
	return members
}

func ParseSecrets(secretsFile string) (secrets map[string]string) {
	yamlSecrets, err := ioutil.ReadFile(secretsFile)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Fatal("Cannot read secrets file.")
	}

	err = yaml.Unmarshal(yamlSecrets, &secrets)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Fatal("Cannot parse secrets file.")
	}
	return secrets
}
