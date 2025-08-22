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

// Package xdslbregistry provides a registry of converters that convert proto
// from load balancing configuration, defined by the xDS API spec, to JSON load
// balancing configuration.
package xdsmetadataregistry

import (
	"fmt"
	"net"
	"encoding/json"
	"google.golang.org/protobuf/proto"

	// https://pkg.go.dev/github.com/envoyproxy/go-control-plane/envoy@v1.32.4/config/core/v3
	v3proto "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
)

type Metadata struct {
	m map[string]MetadataValue
}

type MetadataValue interface {
	mdVal()
}

// can convert this into an interface
type Converter func([]byte) (MetadataValue, error)

type a86Converter struct{}

var (
	// m is a map from proto type to Converter.
	m = make(map[string]Converter)
)

// Register registers the converter to the map keyed on a proto type. Must be
// called at init time. Not thread safe.
func Register(protoType string, c Converter) {
	m[protoType] = c
}

// SetRegistry sets the xDS LB registry. Must be called at init time. Not thread
// safe.
func SetRegistry(registry map[string]Converter) {
	m = registry
}

// implements MetadataValue to be returned in convert
type a86MetadataValue struct {
	MetadataValue
	// this is the value that contains an address string
	address string
	// concrete type
	
}

type JSONMetadata struct {
	MetadataValue 
	Data json.RawMessage
}


func (a86Converter) convert(anyBytes []byte) (MetadataValue, error) {
	// Deserialize anyBytes into a86MetadataValue.
	// getting a Cluster.Metadata.TypedFilterMetadata
	// checkout unmarshal_cds.go and unmarshal_eds.go parseEndpoints() just want to return address
	// string and put that string in a struct
	// unmarshal into this envoy/api/envoy/config/core/v3/address.proto
	addressProto := &v3proto.Address{}

	if err := proto.Unmarshal(anyBytes, addressProto); err != nil {
		return nil, fmt.Errorf("failed to unmarshal resource: %v", err)
	}
	socket_address := addressProto.GetSocketAddress()
	if socket_address == nil {
		return nil, fmt.Errorf("no socket_address field in metadata")
	}
	port_value := socket_address.GetPortValue()
	if port_value == 0 {
		return nil, fmt.Errorf("port value not set in socket_address")
	}
	if net.ParseIP(socket_address.GetAddress()) == nil {
		return nil, fmt.Errorf("address field is not a valid IPv4 or IPv6 address: %q", socket_address.GetAddress())
	}
	// verify socket address, IPv4, and IPv6, and verify that port_value
	// search for those
	metadata := a86MetadataValue{
		address: socket_address.Address,
	}

	return metadata, nil
}

// RegisterA86Converter registers the converter for A86 metadata.
func RegisterA86Converter() {
	Register(
		"envoy.http11_proxy_transport_socket.proxy_address",
		a86Converter{}.convert)
}

// writes the code that grabs the converter from the map lookup the registry and call the proper convert
// internal cluster will have a map from string to the MetadataValue interface
func init() {
	RegisterA86Converter()
}

func GetConverter(typeURL string) Converter {
	return m[typeURL]
}
