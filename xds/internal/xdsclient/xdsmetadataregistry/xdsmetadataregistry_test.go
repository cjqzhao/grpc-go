/*
 *
 * Copyright 2023 gRPC authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package xdsmetadataregistry

import (
	"testing"

	v3proto "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/grpc/internal/grpctest"
	"google.golang.org/protobuf/proto"
)

const a86FilterName = "envoy.http11_proxy_transport_socket.proxy_address"

//several different success cases and failure cases
// ipv4 with a port, ipv6 with a port, unmarhsal_cdstest can just have one success and one failure case.\

type s struct {
	grpctest.Tester
}

func Test(t *testing.T) {
	grpctest.RunSubTests(t, s{})
}

func (s) TestA86ConverterSuccess(t *testing.T) {
	converter, ok := m[a86FilterName]
	if !ok {
		t.Fatalf("Converter for %q not found in registry", a86FilterName)
	}
	tests := []struct {
		name string
		addr *v3proto.Address
		want a86MetadataValue
	}{
		{
			name: "valid IPv4 address and port",
			addr: &v3proto.Address{
				Address: &v3proto.Address_SocketAddress{
					SocketAddress: &v3proto.SocketAddress{
						Address: "192.168.1.1",
						PortSpecifier: &v3proto.SocketAddress_PortValue{
							PortValue: 8080,
						},
					},
				},
			},
			want: a86MetadataValue{
				address: "192.168.1.1",
			},
		},
		{
			name: "valid IPv6 address and port",
			addr: &v3proto.Address{
				Address: &v3proto.Address_SocketAddress{
					SocketAddress: &v3proto.SocketAddress{
						Address: "2001:0db8:85a3:0000:0000:8a2e:0370:7334",
						PortSpecifier: &v3proto.SocketAddress_PortValue{
							PortValue: 9090,
						},
					},
				},
			},
			want: a86MetadataValue{
				address: "2001:0db8:85a3:0000:0000:8a2e:0370:7334",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			anyBytes, err := proto.Marshal(tt.addr)
			if err != nil {
				t.Fatalf("Failed to marshal address proto: %v", err)
			}

			got, err := converter(anyBytes)
			if err != nil {
				t.Fatalf("convert() failed with error: %v", err)
			}

			// Compares result versus want
			if diff := cmp.Diff(tt.want, got, cmp.AllowUnexported(a86MetadataValue{})); diff != "" {
				t.Errorf("convert() returned unexpected diff (-want +got):\n%s", diff)
			}
		})
	}
}

func (s) TestA86ConverterFailure(t *testing.T) {
	converter, ok := m[a86FilterName]
	if !ok {
		t.Fatalf("Converter for %q not found in registry", a86FilterName)
	}
	tests := []struct {
		name    string
		addr    *v3proto.Address
		wantErr string
	}{
		{
			name: "invalid address",
			addr: &v3proto.Address{
				Address: &v3proto.Address_SocketAddress{
					SocketAddress: &v3proto.SocketAddress{
						Address: "invalid-ip",
						PortSpecifier: &v3proto.SocketAddress_PortValue{
							PortValue: 8080,
						},
					},
				},
			},
			wantErr: "address field is not a valid IPv4 or IPv6 address: \"invalid-ip\"",
		},
		{
			name: "missing socket_address",
			addr: &v3proto.Address{
				// No SocketAddress field set
			},
			wantErr: "no socket_address field in metadata",
		},
		{
			name: "address is not a socket address",
			addr: &v3proto.Address{
				Address: &v3proto.Address_EnvoyInternalAddress{
					EnvoyInternalAddress: &v3proto.EnvoyInternalAddress{
						AddressNameSpecifier: &v3proto.EnvoyInternalAddress_ServerListenerName{
							ServerListenerName: "some-internal-listener",
						},
					},
				},
			},
			wantErr: "no socket_address field in metadata",
		},
		{
			name: "port value not set",
			addr: &v3proto.Address{
				Address: &v3proto.Address_SocketAddress{
					SocketAddress: &v3proto.SocketAddress{
						Address:       "127.0.0.1",
						PortSpecifier: nil, // Port value is not set
					},
				},
			},
			wantErr: "port value not set in socket_address",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			anyBytes, err := proto.Marshal(tt.addr)
			if err != nil {
				t.Fatalf("Failed to marshal address proto: %v", err)
			}

			// Call the convert function and check the returned error.
			_, gotErr := converter(anyBytes)
			if gotErr == nil || gotErr.Error() != tt.wantErr {
				t.Errorf("convert() got error = %v, wantErr = %q", gotErr, tt.wantErr)
			}
		})
	}
}
