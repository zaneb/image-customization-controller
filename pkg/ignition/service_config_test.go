package ignition

import (
	"net/url"
	"reflect"
	"testing"

	ignition_config_types_32 "github.com/coreos/ignition/v2/config/v3_2/types"
	"github.com/google/go-cmp/cmp"
	"k8s.io/utils/pointer"
)

func TestIronicPythonAgentConf(t *testing.T) {
	tests := []struct {
		name                          string
		ironicBaseURL                 string
		ironicInspectorVlanInterfaces string
		want                          ignition_config_types_32.File
	}{
		{
			name:          "basic",
			ironicBaseURL: "http://example.com/foo",
			want: ignition_config_types_32.File{
				Node: ignition_config_types_32.Node{Path: "/etc/ironic-python-agent.conf"},
				FileEmbedded1: ignition_config_types_32.FileEmbedded1{
					Contents: ignition_config_types_32.Resource{
						Source: pointer.StringPtr("data:," + url.QueryEscape(`
[DEFAULT]
api_url = http://example.com/foo:6385
inspection_callback_url = http://example.com/foo:5050/v1/continue
insecure = True

collect_lldp = True
enable_vlan_interfaces = all
inspection_collectors = default,extra-hardware,logs
inspection_dhcp_all_interfaces = True
`))}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &ignitionBuilder{
				ironicBaseURL: tt.ironicBaseURL,
			}
			if got := b.ironicPythonAgentConf(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf(cmp.Diff(tt.want, got))
			}
		})
	}
}

func TestIronicAgentService(t *testing.T) {
	tests := []struct {
		name                  string
		ironicAgentImage      string
		ironicAgentPullSecret string
		want                  ignition_config_types_32.Unit
	}{
		{
			name:                  "basic",
			ironicAgentImage:      "http://example.com/foo:latest",
			ironicAgentPullSecret: "foo",
			want: ignition_config_types_32.Unit{
				Name:     "ironic-agent.service",
				Enabled:  pointer.BoolPtr(true),
				Contents: pointer.StringPtr("data:;base64,[Unit]\\nDescription=Ironic Agent\\nAfter=network-online.target\\nWants=network-online.target\\n[Service]\\nTimeoutStartSec=0\\nExecStartPre=/bin/podman pull http://example.com/foo:latest --tls-verify=false --authfile=/etc/authfile.json\\nExecStart=/bin/podman run --privileged --network host --mount type=bind,src=/etc/ironic-python-agent.conf,dst=/etc/ironic-python-agent/ignition.conf --mount type=bind,src=/dev,dst=/dev --mount type=bind,src=/sys,dst=/sys --mount type=bind,src=/,dst=/mnt/coreos --name ironic-agent http://example.com/foo:latest\\n[Install]\\nWantedBy=multi-user.target"),
			},
		}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &ignitionBuilder{
				ironicAgentImage:      tt.ironicAgentImage,
				ironicAgentPullSecret: tt.ironicAgentPullSecret,
			}
			if got := b.ironicAgentService(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf(cmp.Diff(tt.want, got))
			}
		})
	}
}
