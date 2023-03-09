package ops

import (
	"crypto/tls"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/ongoingio/urljoin"
)

const (
	adminAuthHeader = "admin-auth"
	ssoEndpoint     = "/admin/sso"
)

type loginResponse struct {
	Meta string `json:"Meta"`
}

type DashboardAdmin struct {
	Server
	Client *resty.Client
}

// SSO allows you to generate a temporary authentication URL, valid for 60 seconds.
// section can be either "dashboard" or "portal"
func (s *DashboardAdmin) SSO(section string, orgId string, email string, groupId string) (string, error) {
	var loginResp loginResponse
	// Make sure section is either "dashboard" or "portal"
	if section != "dashboard" && section != "portal" {
		return "", fmt.Errorf("sso section must be 'dashboard' or 'portal' but got '%s'", section)
	}
	// Initialize the client if it hasn't been initialized yet
	if s.Client == nil {
		s.Client = resty.New()
	}
	if s.Server.AllowInsecure {
		s.Client.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	}

	// Build the request. Resty will automatically encode the body as JSON when
	// the Content-Type header is set to "application/json"
	resp, err := s.Client.R().
		SetHeader(adminAuthHeader, s.Secret).
		SetHeader("Content-Type", "application/json").
		SetBody(map[string]string{
			"ForSection":   section,
			"OrgID":        orgId,
			"EmailAddress": email,
			"GroupID":      groupId,
		}).
		SetResult(&loginResp).
		Post(urljoin.Join(s.Url, ssoEndpoint))
	if err != nil {
		return "", fmt.Errorf("failed to send HTTP request: %v", err)
	}
	if resp.StatusCode() != 200 {
		return "", fmt.Errorf("HTTP request failed with status code %v: %s", resp.StatusCode(), resp.String())
	}
	if err := resp.Error(); err != nil {
		return "", fmt.Errorf("failed to decode HTTP response: %v", err)
	}

	return fmt.Sprintf("%s/tap?nonce=%s", s.Url, loginResp.Meta), nil
}

// GetOrganizations will get a list of organizations from the Tyk instance.
func (s *DashboardAdmin) GetOrganizations() (*[]Organization, error) {
	endpoint := "admin/organisations"
	response := new(struct {
		Organisations []Organization `json:"organisations"`
		Pages         int            `json:"pages"`
	})

	// Initialize the client if it hasn't been initialized yet
	if s.Client == nil {
		s.Client = resty.New()
	}
	if s.Server.AllowInsecure {
		s.Client.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	}

	resp, err := s.Client.R().
		SetHeader("admin-auth", s.Secret).
		SetHeader("Content-Type", "application/json").
		SetResult(response).
		Get(urljoin.Join(s.Url, endpoint))
	if err != nil {
		return nil, fmt.Errorf("failed to send HTTP request: %v", err)
	}
	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("HTTP request failed with status code %v: %s", resp.StatusCode(), resp.String())
	}
	if err := resp.Error(); err != nil {
		return nil, fmt.Errorf("failed to decode HTTP response: %v", err)
	}

	return &response.Organisations, nil
}

type Organization struct {
	Id             string        `json:"id"`
	OwnerName      string        `json:"owner_name"`
	OwnerSlug      string        `json:"owner_slug"`
	CnameEnabled   bool          `json:"cname_enabled"`
	Cname          string        `json:"cname"`
	Apis           []interface{} `json:"apis"`
	SsoEnabled     bool          `json:"sso_enabled"`
	DeveloperQuota int           `json:"developer_quota"`
	DeveloperCount int           `json:"developer_count"`
	EventOptions   struct {
	} `json:"event_options"`
	HybridEnabled bool `json:"hybrid_enabled"`
	Ui            struct {
		Languages struct {
		} `json:"languages"`
		HideHelp    bool   `json:"hide_help"`
		DefaultLang string `json:"default_lang"`
		LoginPage   struct {
		} `json:"login_page"`
		Nav struct {
		} `json:"nav"`
		Uptime struct {
		} `json:"uptime"`
		PortalSection struct {
		} `json:"portal_section"`
		Designer struct {
		} `json:"designer"`
		DontShowAdminSockets           bool `json:"dont_show_admin_sockets"`
		DontAllowLicenseManagement     bool `json:"dont_allow_license_management"`
		DontAllowLicenseManagementView bool `json:"dont_allow_license_management_view"`
		Cloud                          bool `json:"cloud"`
		Dev                            bool `json:"dev"`
	} `json:"ui"`
	OrgOptionsMeta struct {
	} `json:"org_options_meta"`
	OpenPolicy struct {
		Rules   string `json:"rules"`
		Enabled bool   `json:"enabled"`
	} `json:"open_policy"`
	AdditionalPermissions struct {
	} `json:"additional_permissions"`
}
