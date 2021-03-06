// Copyright 2017 Istio Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package external

import (
	networking "istio.io/api/networking/v1alpha3"
	"istio.io/istio/pilot/pkg/model"
)

// externalDiscovery communicates with ExternalService CRDs and monitors for changes
type externalDiscovery struct {
	store model.IstioConfigStore
}

// NewServiceDiscovery creates a new ExternalServices discovery service
func NewServiceDiscovery(store model.IstioConfigStore) model.ServiceDiscovery {
	return &externalDiscovery{
		store: store,
	}
}

// Services list declarations of all services in the system
func (d *externalDiscovery) Services() ([]*model.Service, error) {
	services := make([]*model.Service, 0)
	for _, config := range d.store.ExternalServices() {
		externalService := config.Spec.(*networking.ExternalService)
		services = append(services, convertServices(externalService)...)
	}

	return services, nil
}

// GetService retrieves a service by host name if it exists
func (d *externalDiscovery) GetService(hostname string) (*model.Service, error) {
	for _, service := range d.getServices() {
		if service.Hostname == hostname {
			return service, nil
		}
	}

	return nil, nil
}

func (d *externalDiscovery) getServices() []*model.Service {
	services := make([]*model.Service, 0)
	for _, config := range d.store.ExternalServices() {
		externalService := config.Spec.(*networking.ExternalService)
		services = append(services, convertServices(externalService)...)
	}
	return services
}

// ManagementPorts retries set of health check ports by instance IP.
// This does not apply to External Service registry, as External Services do not
// manage the service instances.
func (d *externalDiscovery) ManagementPorts(addr string) model.PortList {
	return nil
}

// Instances retrieves instances for a service and its ports that match
// any of the supplied labels. All instances match an empty tag list.
func (d *externalDiscovery) Instances(hostname string, ports []string,
	labels model.LabelsCollection) ([]*model.ServiceInstance, error) {
	portMap := make(map[string]bool)
	for _, port := range ports {
		portMap[port] = true
	}

	out := []*model.ServiceInstance{}
	for _, config := range d.store.ExternalServices() {
		externalService := config.Spec.(*networking.ExternalService)
		for _, instance := range convertInstances(externalService) {
			if instance.Service.Hostname == hostname &&
				labels.HasSubsetOf(instance.Labels) &&
				portMatch(instance, portMap) {
				out = append(out, instance)
			}
		}
	}

	return out, nil
}

// returns true if an instance's port matches with any in the provided list
func portMatch(instance *model.ServiceInstance, portMap map[string]bool) bool {
	return len(portMap) == 0 || portMap[instance.Endpoint.ServicePort.Name]
}

// GetProxyServiceInstances lists service instances co-located with a given proxy
func (d *externalDiscovery) GetProxyServiceInstances(node *model.Proxy) ([]*model.ServiceInstance, error) {
	// There is no proxy sitting next to google.com.  If supplied, istio will end up generating a full envoy
	// configuration with routes to internal services, (listeners, etc.) for the external service
	// (which does not exist in the cluster).
	return []*model.ServiceInstance{}, nil
}
