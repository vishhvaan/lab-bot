package config

import (
	"io/ioutil"

	"go.uber.org/zap"
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
		zap.L().Fatal("Cannot read members file.",
			zap.Error(err),
		)
	}

	err = yaml.Unmarshal(yamlMembers, &members)
	if err != nil {
		zap.L().Fatal("Cannot parse members file.",
			zap.Error(err),
		)
	}
	return members
}

func ParseSecrets(secretsFile string) (secrets map[string]string) {
	yamlSecrets, err := ioutil.ReadFile(secretsFile)
	if err != nil {
		zap.L().Fatal("Cannot read secrets file.",
			zap.Error(err),
		)
	}

	err = yaml.Unmarshal(yamlSecrets, &secrets)
	if err != nil {
		zap.L().Fatal("Cannot parse secrets file.",
			zap.Error(err),
		)
	}
	return secrets
}
