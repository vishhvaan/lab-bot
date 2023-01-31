package config

import (
	"io/ioutil"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

var Secrets map[string]string

var Members map[string]Member

type Member struct {
	FirstName string   `yaml:"first_name"`
	LastName  string   `yaml:"last_name"`
	Subgroup  string   `yaml:"subgroup"`
	UserID    string   `yaml:"userID"`
	Birthday  string   `yaml:"birthday"`
	Roles     []string `yaml:"roles"`
}

func ParseMembers(membersFile string) {
	yamlMembers, err := ioutil.ReadFile(membersFile)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Fatal("Cannot read members file.")
	}

	err = yaml.Unmarshal(yamlMembers, &Members)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Fatal("Cannot parse members file.")
	}
}

func ParseSecrets(secretsFile string) {
	yamlSecrets, err := ioutil.ReadFile(secretsFile)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Fatal("Cannot read secrets file.")
	}

	err = yaml.Unmarshal(yamlSecrets, &Secrets)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Fatal("Cannot parse secrets file.")
	}
}
