package convertor

import (
	"encoding/json"
	"fmt"
	"strings"

	"google.golang.org/grpc/xds/internal/xdsclient/xdsmetadataregistry"
)

func init() {
	xdsmetadataregistry.Register("type.googleapis.com/envoy.extensions.load_balancing_policies.client_side_weighted_round_robin.v3.ClientSideWeightedRoundRobin", convertWeightedRoundRobinProtoToServiceConfig)

}