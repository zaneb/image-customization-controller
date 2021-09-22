package ignition

import (
	"reflect"
	"testing"

	ignition_config_types_32 "github.com/coreos/ignition/v2/config/v3_2/types"
	"github.com/google/go-cmp/cmp"
	"k8s.io/utils/pointer"
)

func TestNMStateOutputToFiles(t *testing.T) {
	tests := []struct {
		name            string
		generatedConfig []byte
		want            []ignition_config_types_32.File
		wantErr         bool
	}{
		{
			name:            "empty",
			generatedConfig: []byte(`---`),
			want:            []ignition_config_types_32.File{},
		},
		{
			name: "basic working",
			generatedConfig: []byte(`---
NetworkManager:
- - eth1.nmconnection
  - '[connection]

	id=eth1

	uuid=b6a3fd9a-4c76-4213-a698-8c0f0749b193

	type=ethernet

	interface-name=eth1

	permissions=


	[ethernet]

	mac-address-blacklist=


	[ipv4]

	dhcp-client-id=mac

	dns-search=

	method=disabled


	[ipv6]

	addr-gen-mode=eui64

	dhcp-duid=ll

	dhcp-iaid=mac

	dns-search=

	method=disabled


	[proxy]

	'
- - linux-br0.nmconnection
  - '[connection]

	id=linux-br0

	uuid=f942ffa5-668d-41f3-86bd-ef53e35565f4

	type=bridge

	autoconnect-slaves=1

	interface-name=linux-br0

	permissions=


	[bridge]


	[ipv4]

	dhcp-client-id=mac

	dns-search=

	method=disabled


	[ipv6]

	addr-gen-mode=eui64

	dhcp-duid=ll

	dhcp-iaid=mac

	dns-search=

	method=disabled


	[proxy]

	'
`),
			want: []ignition_config_types_32.File{
				{
					Node: ignition_config_types_32.Node{Path: "/etc/NetworkManager/system-connections/eth1.nmconnection"},
					FileEmbedded1: ignition_config_types_32.FileEmbedded1{
						Contents: ignition_config_types_32.Resource{
							Source: pointer.StringPtr(`data:,[connection]
id=eth1
uuid=b6a3fd9a-4c76-4213-a698-8c0f0749b193
type=ethernet
interface-name=eth1
permissions=

[ethernet]
mac-address-blacklist=

[ipv4]
dhcp-client-id=mac
dns-search=
method=disabled

[ipv6]
addr-gen-mode=eui64
dhcp-duid=ll
dhcp-iaid=mac
dns-search=
method=disabled

[proxy]
`)}},
				},
				{
					Node: ignition_config_types_32.Node{Path: "/etc/NetworkManager/system-connections/linux-br0.nmconnection"},
					FileEmbedded1: ignition_config_types_32.FileEmbedded1{
						Contents: ignition_config_types_32.Resource{
							Source: pointer.StringPtr(`data:,[connection]
id=linux-br0
uuid=f942ffa5-668d-41f3-86bd-ef53e35565f4
type=bridge
autoconnect-slaves=1
interface-name=linux-br0
permissions=

[bridge]

[ipv4]
dhcp-client-id=mac
dns-search=
method=disabled

[ipv6]
addr-gen-mode=eui64
dhcp-duid=ll
dhcp-iaid=mac
dns-search=
method=disabled

[proxy]
`)}},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &ignitionBuilder{}
			got, err := b.nmstateOutputToFiles(tt.generatedConfig)
			if (err != nil) != tt.wantErr {
				t.Errorf("ignitionBuilder.nmstateOutputToFiles() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf(cmp.Diff(tt.want, got))
			}
		})
	}
}
