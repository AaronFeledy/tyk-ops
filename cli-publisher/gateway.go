package cli_publisher

import (
	"errors"
	"github.com/AaronFeledy/tyk-ops/pkg/clients/gateway"
	"github.com/AaronFeledy/tyk-ops/pkg/clients/objects"
)

type GatewayPublisher struct {
	Secret   string
	Hostname string
	// Additional options to pass to the gateway client
	ClientOptions struct {
		// InsecureSkipVerify is a flag that specifies if we should validate the
		// server's TLS certificate.
		InsecureSkipVerify bool
		// Skip creating APIs if they already exist
		SkipExisting bool
	}
}

func (p *GatewayPublisher) CreateAPIs(apiDefs *[]objects.DBApiDefinition) error {
	c, err := gateway.NewGatewayClient(p.Hostname, p.Secret)
	c.SkipExisting = p.ClientOptions.SkipExisting
	c.InsecureSkipVerify = p.ClientOptions.InsecureSkipVerify
	if err != nil {
		return err
	}

	return c.CreateAPIs(apiDefs)
}

func (p *GatewayPublisher) UpdateAPIs(apiDefs *[]objects.DBApiDefinition) error {
	c, err := gateway.NewGatewayClient(p.Hostname, p.Secret)
	c.InsecureSkipVerify = p.ClientOptions.InsecureSkipVerify
	if err != nil {
		return err
	}

	return c.UpdateAPIs(apiDefs)
}

func (p *GatewayPublisher) Name() string {
	return "Gateway Publisher"
}

func (p *GatewayPublisher) Reload() error {
	c, err := gateway.NewGatewayClient(p.Hostname, p.Secret)
	c.InsecureSkipVerify = p.ClientOptions.InsecureSkipVerify
	if err != nil {
		return err
	}

	return c.Reload()
}

func (p *GatewayPublisher) SyncAPIs(apiDefs []objects.DBApiDefinition) error {
	c, err := gateway.NewGatewayClient(p.Hostname, p.Secret)
	c.InsecureSkipVerify = p.ClientOptions.InsecureSkipVerify
	if err != nil {
		return err
	}

	return c.SyncAPIs(apiDefs)
}

func (p *GatewayPublisher) CreatePolicies(pols *[]objects.Policy) error {
	return errors.New("Policy handling not supported by Gateway publisher")
}

func (p *GatewayPublisher) UpdatePolicies(pols *[]objects.Policy) error {
	return errors.New("Policy handling not supported by Gateway publisher")
}

func (p *GatewayPublisher) SyncPolicies(pols []objects.Policy) error {
	return errors.New("Policy handling not supported by Gateway publisher")
}
