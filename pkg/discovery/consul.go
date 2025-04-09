package discovery

import (
	"fmt"
	"strconv"

	"github.com/hashicorp/consul/api"
	"github.com/sdshorin/generia/pkg/logger"
	"go.uber.org/zap"
)

// ServiceDiscovery defines interface for service discovery
type ServiceDiscovery interface {
	Register(serviceID, name, host string, port int, tags []string) error
	Deregister(serviceID string) error
	ResolveService(serviceName string) (string, error)
}

// ConsulServiceDiscovery implements ServiceDiscovery using Consul
type ConsulServiceDiscovery struct {
	client *api.Client
}

// NewConsulClient creates a new Consul client
func NewConsulClient(address string) (ServiceDiscovery, error) {
	config := api.DefaultConfig()
	config.Address = address

	client, err := api.NewClient(config)
	if err != nil {
		return nil, err
	}

	return &ConsulServiceDiscovery{client: client}, nil
}

// Register registers a service with Consul
func (c *ConsulServiceDiscovery) Register(serviceID, name, host string, port int, tags []string) error {
	// In Docker environment, we need to use the service name as the address
	// because 0.0.0.0 is not valid for service registration in Consul
	
	// If the host is 0.0.0.0, use the service name instead for registration
	// This is crucial for proper service discovery in containerized environments
	registrationAddress := host
	if host == "0.0.0.0" || host == "" {
		// Use service name as the address which matches the container name in Docker
		// This enables proper DNS resolution within the Docker network
		registrationAddress = name 
		logger.Logger.Info("Using service name as address for Consul registration",
			zap.String("serviceName", name),
			zap.String("originalHost", host))
	}
	
	logger.Logger.Debug("Attempting to register service with Consul",
        zap.String("serviceID", serviceID),
        zap.String("name", name),
        zap.String("registrationAddress", registrationAddress),
        zap.Int("port", port),
        zap.Strings("tags", tags),
    )

	// Define service registration
	registration := &api.AgentServiceRegistration{
		ID:      serviceID,
		Name:    name,
		Address: registrationAddress,
		Port:    port,
		Tags:    tags,
		// Configure health check for the service
		Check: &api.AgentServiceCheck{
			// Use the registration address for health checks
			GRPC:                           fmt.Sprintf("%s:%d", registrationAddress, port),
			GRPCUseTLS:                     false,
			Interval:                       "10s", 
			Timeout:                        "5s",  // Increased timeout for reliability
			DeregisterCriticalServiceAfter: "30s",
		},
	}

	// Register service
	err := c.client.Agent().ServiceRegister(registration)
	if err != nil {
		logger.Logger.Error("Failed to register service with Consul", 
			zap.String("serviceID", serviceID),
			zap.Error(err))
		return err
	}

	logger.Logger.Info("Service registered with Consul",
		zap.String("id", serviceID),
		zap.String("name", name),
		zap.String("address", registrationAddress),
		zap.Int("port", port),
	)

	return nil
}

// Deregister deregisters a service with Consul
func (c *ConsulServiceDiscovery) Deregister(serviceID string) error {
	err := c.client.Agent().ServiceDeregister(serviceID)
	if err != nil {
		return err
	}

	logger.Logger.Info("Service deregistered from Consul", zap.String("id", serviceID))
	return nil
}

// ResolveService resolves a service address by name
func (c *ConsulServiceDiscovery) ResolveService(serviceName string) (string, error) {
	// Get healthy service instances
	services, _, err := c.client.Health().Service(serviceName, "", true, nil)
	if err != nil {
		return "", err
	}

	if len(services) == 0 {
		// Try to get unhealthy service instances as well (for debugging/fallback)
		allServices, _, err := c.client.Health().Service(serviceName, "", false, nil)
		if err != nil {
			return "", err
		}
		
		if len(allServices) == 0 {
			return "", fmt.Errorf("no instances of service %s found (healthy or unhealthy)", serviceName)
		}
		
		logger.Logger.Warn("No healthy instances found for service, using DNS resolution instead",
			zap.String("serviceName", serviceName))
			
		// In Docker/Kubernetes environment, service discovery should work using DNS
		// The container/service name should be resolvable within the network
		return serviceName + ":" + strconv.Itoa(allServices[0].Service.Port), nil
	}

	// Simple load balancing - pick the first healthy instance
	service := services[0].Service
	
	// Get the address from the service
	address := service.Address
	
	// This is important - in Docker Compose, we should have a valid address from registration
	// But we'll add a fallback just in case
	if address == "" {
		// Fall back to service name as DNS entry (Docker Compose sets this up automatically)
		address = serviceName
		logger.Logger.Warn("Empty service address detected, falling back to service name",
			zap.String("serviceName", serviceName))
	}

	logger.Logger.Info("Resolved service address",
		zap.String("serviceName", serviceName),
		zap.String("address", address),
		zap.Int("port", service.Port))

	return address + ":" + strconv.Itoa(service.Port), nil
}