package consul

import (
	"fmt"
	"log"

	"github.com/hashicorp/consul/api"
)

type Register struct {
	Host    string
	Name    string
	Address string
	Port    int
	Check   *api.AgentServiceCheck
	client  *api.Client
}

func NewRegister(host string, name string, address string, port int, check *api.AgentServiceCheck) *Register {
	return &Register{
		Host:    host,
		Name:    name,
		Address: address,
		Port:    port,
		Check:   check,
	}
}

func NewRegisterGRPC(host string, name string, address string, port int) *Register {
	return NewRegister(host, name, address, port, &api.AgentServiceCheck{
		Interval:                       "3s",
		DeregisterCriticalServiceAfter: "1m",
		GRPC:                           fmt.Sprintf("%s:%d/%s", address, port, name),
	})
}

func NewRegisterHTTP(host string, name string, address string, port int) *Register {
	return NewRegister(host, name, address, port, &api.AgentServiceCheck{
		Interval:                       "10s",
		DeregisterCriticalServiceAfter: "1m",
		HTTP:                           fmt.Sprintf("http://%s:%d/health", address, port),
	})
}

func (r *Register) serviceID() string {
	return fmt.Sprintf("%s-%s-%d", r.Name, r.Address, r.Port)
}

func (r *Register) Register() error {
	if r.client == nil {
		cfg := api.DefaultConfig()
		cfg.Address = r.Host

		client, err := api.NewClient(cfg)
		if err != nil {
			return err
		}
		r.client = client
	}

	log.Printf("[Consul] Register: %+v\n", *r)
	return r.client.Agent().ServiceRegister(&api.AgentServiceRegistration{
		ID:      r.serviceID(),
		Name:    r.Name,
		Address: r.Address,
		Port:    r.Port,
		Check:   r.Check,
	})
}

func (r *Register) Deregister() error {
	if r.client == nil {
		return nil
	}

	log.Printf("[Consul] Deregister: %+v\n", *r)
	return r.client.Agent().ServiceDeregister(r.serviceID())
}
