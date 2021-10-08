package env

import "github.com/kelseyhightower/envconfig"

type EnvInputs struct {
	DeployISO             string `envconfig:"DEPLOY_ISO" required:"true"`
	IronicBaseURL         string `envconfig:"IRONIC_BASE_URL" required:"true"`
	IronicAgentImage      string `envconfig:"IRONIC_AGENT_IMAGE"`
	IronicAgentPullSecret string `envconfig:"IRONIC_AGENT_PULL_SECRET"`
	IronicRAMDiskSSHKey   string `envconfig:"IRONIC_RAMDISK_SSH_KEY"`
}

func New() (*EnvInputs, error) {
	env := &EnvInputs{}
	err := envconfig.Process("", env)
	return env, err
}
