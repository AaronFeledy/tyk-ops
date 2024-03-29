package cli_publisher

import (
	"fmt"
	"github.com/AaronFeledy/tyk-ops/pkg/clients/dashboard"
	"github.com/AaronFeledy/tyk-ops/pkg/clients/objects"
)

type DashboardPublisher struct {
	Secret      string
	Hostname    string
	OrgOverride string
	// Additional options to pass to the dashboard client
	ClientOptions struct {
		// InsecureSkipVerify is a flag that specifies if we should validate the
		// server's TLS certificate.
		InsecureSkipVerify bool
		// Skip creating APIs if they already exist
		SkipExisting bool
	}
}

func (p *DashboardPublisher) enforceOrgID(apiDefs *[]objects.DBApiDefinition) {
	if p.OrgOverride != "" {
		fmt.Println("org override detected, setting.")

		for i := range *apiDefs {
			(*apiDefs)[i].OrgID = p.OrgOverride
		}
	}
}

func (p *DashboardPublisher) enforceOrgIDForPolicies(pols *[]objects.Policy) {
	if p.OrgOverride != "" {
		fmt.Println("org override detected, setting.")

		for i := range *pols {
			(*pols)[i].OrgID = p.OrgOverride
		}
	}
}

func (p *DashboardPublisher) CreateAPIs(apiDefs *[]objects.DBApiDefinition) error {
	c, err := dashboard.NewDashboardClient(p.Hostname, p.Secret, p.OrgOverride)
	c.InsecureSkipVerify = p.ClientOptions.InsecureSkipVerify
	c.SkipExisting = p.ClientOptions.SkipExisting
	if err != nil {
		return err
	}
	if p.OrgOverride == "" {
		p.OrgOverride = c.OrgID
	}

	p.enforceOrgID(apiDefs)
	return c.CreateAPIs(apiDefs)
}

func (p *DashboardPublisher) UpdateAPIs(apiDefs *[]objects.DBApiDefinition) error {
	c, err := dashboard.NewDashboardClient(p.Hostname, p.Secret, p.OrgOverride)
	c.InsecureSkipVerify = p.ClientOptions.InsecureSkipVerify
	if err != nil {
		return err
	}
	if p.OrgOverride == "" {
		p.OrgOverride = c.OrgID
	}

	p.enforceOrgID(apiDefs)
	return c.UpdateAPIs(apiDefs)
}

func (p *DashboardPublisher) SyncAPIs(apiDefs []objects.DBApiDefinition) error {
	c, err := dashboard.NewDashboardClient(p.Hostname, p.Secret, p.OrgOverride)
	c.InsecureSkipVerify = p.ClientOptions.InsecureSkipVerify
	if err != nil {
		return err
	}

	if p.OrgOverride == "" {
		p.OrgOverride = c.OrgID
	}

	if p.OrgOverride != "" {
		fixedDefs := make([]objects.DBApiDefinition, len(apiDefs))
		for i, a := range apiDefs {
			newDef := a
			newDef.OrgID = p.OrgOverride
			fixedDefs[i] = newDef
		}

		return c.SyncAPIs(fixedDefs)
	}

	return c.SyncAPIs(apiDefs)
}

func (p *DashboardPublisher) Reload() error {
	fmt.Println("Dashboard does not require explicit reload. Skipping Reload.")
	return nil
}

func (p *DashboardPublisher) Name() string {
	return "Dashboard Publisher"
}

func (p *DashboardPublisher) CreatePolicies(pols *[]objects.Policy) error {
	c, err := dashboard.NewDashboardClient(p.Hostname, p.Secret, p.OrgOverride)
	c.InsecureSkipVerify = p.ClientOptions.InsecureSkipVerify
	c.SkipExisting = p.ClientOptions.SkipExisting
	if err != nil {
		return err
	}
	if p.OrgOverride == "" {
		p.OrgOverride = c.OrgID
	}

	p.enforceOrgIDForPolicies(pols)
	return c.CreatePolicies(pols)
}

func (p *DashboardPublisher) UpdatePolicies(pols *[]objects.Policy) error {
	c, err := dashboard.NewDashboardClient(p.Hostname, p.Secret, p.OrgOverride)
	c.InsecureSkipVerify = p.ClientOptions.InsecureSkipVerify
	if err != nil {
		return err
	}
	if p.OrgOverride == "" {
		p.OrgOverride = c.OrgID
	}

	p.enforceOrgIDForPolicies(pols)
	return c.UpdatePolicies(pols)
}

func (p *DashboardPublisher) SyncPolicies(pols []objects.Policy) error {
	c, err := dashboard.NewDashboardClient(p.Hostname, p.Secret, p.OrgOverride)
	c.InsecureSkipVerify = p.ClientOptions.InsecureSkipVerify
	if err != nil {
		return err
	}
	if p.OrgOverride == "" {
		p.OrgOverride = c.OrgID
	}

	if p.OrgOverride != "" {
		fixedPols := make([]objects.Policy, len(pols))
		for i, pol := range pols {
			newPol := pol
			newPol.OrgID = p.OrgOverride
			fixedPols[i] = newPol
		}

		return c.SyncPolicies(fixedPols)
	}

	return c.SyncPolicies(pols)
}
