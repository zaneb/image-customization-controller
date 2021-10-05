package ignition

import (
	"encoding/json"
	"errors"
	"os/exec"
	"strings"

	ignition_config_types_32 "github.com/coreos/ignition/v2/config/v3_2/types"
	vpath "github.com/coreos/vcontext/path"
)

const (
	// https://github.com/openshift/ironic-image/blob/master/scripts/configure-coreos-ipa#L14
	ironicAgentPodmanFlags = "--tls-verify=false"

	// https://github.com/openshift/ironic-image/blob/master/scripts/configure-coreos-ipa#L11
	ironicInspectorVlanInterfaces = "all"
)

type ignitionBuilder struct {
	nmStateData           []byte
	ironicBaseURL         string
	ironicAgentImage      string
	ironicAgentPullSecret string
	ironicRAMDiskSSHKey   string
}

func New(nmStateData []byte, ironicBaseURL, ironicAgentImage, ironicAgentPullSecret, ironicRAMDiskSSHKey string) *ignitionBuilder {
	if ironicAgentImage == "" {
		// https://github.com/openshift/ironic-image/blob/master/scripts/configure-coreos-ipa#L13
		ironicAgentImage = "quay.io/dtantsur/ironic-agent" // TODO check
	}

	return &ignitionBuilder{
		nmStateData:           nmStateData,
		ironicBaseURL:         ironicBaseURL,
		ironicAgentImage:      ironicAgentImage,
		ironicAgentPullSecret: ironicAgentPullSecret,
		ironicRAMDiskSSHKey:   ironicRAMDiskSSHKey,
	}
}

func (b *ignitionBuilder) Generate() ([]byte, error) {
	if b.ironicBaseURL == "" {
		return nil, errors.New("ironicBaseURL is required")
	}
	config := ignition_config_types_32.Config{
		Ignition: ignition_config_types_32.Ignition{
			Version: "3.2.0",
		},
		Storage: ignition_config_types_32.Storage{
			Files: []ignition_config_types_32.File{b.ironicPythonAgentConf()},
		},
		Systemd: ignition_config_types_32.Systemd{
			Units: []ignition_config_types_32.Unit{b.ironicAgentService()},
		},
	}
	if b.ironicAgentPullSecret != "" {
		config.Storage.Files = append(config.Storage.Files, b.authFile())
	}

	if b.ironicRAMDiskSSHKey != "" {
		config.Passwd.Users = append(config.Passwd.Users, ignition_config_types_32.PasswdUser{
			Name: "core",
			SSHAuthorizedKeys: []ignition_config_types_32.SSHAuthorizedKey{
				ignition_config_types_32.SSHAuthorizedKey(strings.TrimSpace(b.ironicRAMDiskSSHKey)),
			},
		})
	}

	if len(b.nmStateData) > 0 {
		nmstatectl := exec.Command("nmstatectl", "gc", "-")
		nmstatectl.Stdin = strings.NewReader(string(b.nmStateData))
		out, err := nmstatectl.Output()
		if err != nil {
			return nil, err
		}

		files, err := nmstateOutputToFiles(out)
		if err != nil {
			return nil, err
		}
		config.Storage.Files = append(config.Storage.Files, files...)
	}

	report := config.Storage.Validate(vpath.ContextPath{})
	if report.IsFatal() {
		return nil, errors.New(report.String())
	}

	return json.Marshal(config)
}
