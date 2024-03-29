package gateway

import (
	"errors"
	"fmt"
	"github.com/AaronFeledy/tyk-ops/pkg/clients/objects"
	"github.com/AaronFeledy/tyk-ops/pkg/output"

	"encoding/json"

	"github.com/gofrs/uuid"
	"github.com/levigross/grequests"
	"github.com/ongoingio/urljoin"
)

type Client struct {
	url    string
	secret string
	// InsecureSkipVerify is a flag that specifies if we should validate the
	// server's TLS certificate.
	InsecureSkipVerify bool
	// Skip creating APIs if they already exist
	SkipExisting bool
}

const (
	endpointAPIs     string = "/tyk/apis/"
	endpointCerts    string = "/tyk/certs"
	reloadAPIs       string = "/tyk/reload/group"
	endpointPolicies string = "/tyk/policies"
)

var (
	UseUpdateError error = errors.New("Object seems to exist (same API ID, Listen Path or Slug), use update()")
	UseCreateError error = errors.New("Object does not exist, use create()")
)

type APIMessage struct {
	Key     string `json:"key"`
	Status  string `json:"status"`
	Action  string `json:"action"`
	Message string `json:"message"`
}

type APISList []objects.APIDefinition

func NewGatewayClient(url, secret string) (*Client, error) {
	return &Client{
		url:    url,
		secret: secret,
	}, nil
}

func (c *Client) SetInsecureTLS(val bool) {
	c.InsecureSkipVerify = val
}

func (c *Client) GetActiveID(def *objects.DBApiDefinition) string {
	return def.APIID
}

func (c *Client) FetchAPIs() ([]objects.DBApiDefinition, error) {
	fullPath := urljoin.Join(c.url, endpointAPIs)

	ro := &grequests.RequestOptions{
		Headers: map[string]string{
			"x-tyk-authorization": c.secret,
			"content-type":        "application/json",
		},
		InsecureSkipVerify: c.InsecureSkipVerify,
	}

	resp, err := grequests.Get(fullPath, ro)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("API Returned error: %v", resp.String())
	}

	apis := APISList{}
	if err := resp.JSON(&apis); err != nil {
		return nil, err
	}

	retList := make([]objects.DBApiDefinition, len(apis))
	for i := range apis {
		retList[i] = objects.DBApiDefinition{APIDefinition: &apis[i]}
	}

	return retList, nil
}

func getAPIsIdentifiers(apiDefs *[]objects.DBApiDefinition) (map[string]*objects.DBApiDefinition, map[string]*objects.DBApiDefinition, map[string]*objects.DBApiDefinition, map[string]*objects.DBApiDefinition) {
	apiids := make(map[string]*objects.DBApiDefinition)
	ids := make(map[string]*objects.DBApiDefinition)
	slugs := make(map[string]*objects.DBApiDefinition)
	paths := make(map[string]*objects.DBApiDefinition)

	for i := range *apiDefs {
		apiDef := (*apiDefs)[i]
		apiids[apiDef.APIID] = &apiDef
		ids[apiDef.Id.Hex()] = &apiDef
		slugs[apiDef.Slug] = &apiDef
		paths[apiDef.Proxy.ListenPath+"-"+apiDef.Domain] = &apiDef
	}

	return apiids, ids, slugs, paths
}

func (c *Client) CreateAPIs(apiDefs *[]objects.DBApiDefinition) error {
	existingAPIs, err := c.FetchAPIs()
	if err != nil {
		return err
	}

	apiids, ids, slugs, paths := getAPIsIdentifiers(&existingAPIs)

	var existsCount int64
	for i := range *apiDefs {
		apiDef := (*apiDefs)[i]
		fmt.Printf("Creating API %v: %v\n", i, apiDef.Name)
		var existsError error
		if thisAPI, ok := apiids[apiDef.APIID]; ok && thisAPI != nil {
			fmt.Println("Warning: API ID Exists")
			existsError = UseUpdateError
		} else if thisAPI, ok := ids[apiDef.Id.Hex()]; ok && thisAPI != nil {
			fmt.Println("Warning: Object ID Exists")
			existsError = UseUpdateError
		} else if thisAPI, ok := slugs[apiDef.Slug]; ok && thisAPI != nil {
			fmt.Println("Warning: Slug Exists")
			existsError = UseUpdateError
		} else if thisAPI, ok := paths[apiDef.Proxy.ListenPath+"-"+apiDef.Domain]; ok && thisAPI != nil {
			fmt.Println("Warning: Listen Path Exists")
			existsError = UseUpdateError
		}

		if existsError != nil {
			if c.SkipExisting {
				existsCount++
				continue
			}
			return existsError
		}

		data, err := json.Marshal(apiDef.APIDefinition)
		if err != nil {
			return err
		}

		// Create
		fullPath := urljoin.Join(c.url, endpointAPIs)
		createResp, err := grequests.Post(fullPath, &grequests.RequestOptions{
			JSON: data,
			Headers: map[string]string{
				"x-tyk-authorization": c.secret,
				"content-type":        "application/json",
			},
			InsecureSkipVerify: c.InsecureSkipVerify,
		})

		if err != nil {
			return err
		}

		if createResp.StatusCode != 200 {
			return fmt.Errorf("API Returned error: %v (code: %v)", createResp.String(), createResp.StatusCode)
		}

		var status APIMessage
		if err := createResp.JSON(&status); err != nil {
			return err
		}

		if status.Status != "ok" {
			return fmt.Errorf("API request completed, but with error: %v", status.Message)
		}

		// initiate a reload
		go c.Reload()

		// Add updated API to existing API list.
		apiids[apiDef.APIID] = &apiDef
		slugs[apiDef.Slug] = &apiDef
		paths[apiDef.Proxy.ListenPath+"-"+apiDef.Domain] = &apiDef

		fmt.Printf("--> Status: OK, ID:%v\n", apiDef.APIID)
	}

	if existsCount > 0 {
		output.User.Println("%v APIs already exist and were skipped", existsCount)
	}

	return nil
}

func (c *Client) Reload() error {
	// Reload
	fmt.Println("Reloading...")
	fullPath := urljoin.Join(c.url, reloadAPIs)
	reloadREsp, err := grequests.Get(fullPath, &grequests.RequestOptions{
		Headers: map[string]string{
			"x-tyk-authorization": c.secret,
		},
		InsecureSkipVerify: c.InsecureSkipVerify,
	})

	if err != nil {
		return err
	}

	if reloadREsp.StatusCode != 200 {
		return fmt.Errorf("API Returned error: %v (code: %v)", reloadREsp.String(), reloadREsp.StatusCode)
	}

	var status APIMessage
	if err := reloadREsp.JSON(&status); err != nil {
		return err
	}

	if status.Status != "ok" {
		fmt.Errorf("API request completed, but with error: %v", status.Message)
	}

	return nil
}

func (c *Client) UpdateAPIs(apiDefs *[]objects.DBApiDefinition) error {
	existingAPIs, err := c.FetchAPIs()
	if err != nil {
		return err
	}

	apiids, ids, slugs, paths := getAPIsIdentifiers(&existingAPIs)

	for i := range *apiDefs {
		apiDef := (*apiDefs)[i]
		fmt.Printf("Updating API %v: %v\n", i, apiDef.Name)
		if thisAPI, ok := apiids[apiDef.APIID]; ok && thisAPI != nil {
			apiDef.Id = thisAPI.Id
		} else if thisAPI, ok := ids[apiDef.Id.Hex()]; ok && thisAPI != nil {
			if apiDef.APIID == "" {
				apiDef.APIID = thisAPI.APIID
			}
		} else if thisAPI, ok := slugs[apiDef.Slug]; ok && thisAPI != nil {
			if apiDef.APIID == "" {
				apiDef.APIID = thisAPI.APIID
			}
			if apiDef.Id == "" {
				apiDef.Id = thisAPI.Id
			}
		} else if thisAPI, ok := paths[apiDef.Proxy.ListenPath+"-"+apiDef.Domain]; ok && thisAPI != nil {
			if apiDef.APIID == "" {
				apiDef.APIID = thisAPI.APIID
			}
			if apiDef.Id == "" {
				apiDef.Id = thisAPI.Id
			}
		} else {
			return UseCreateError
		}

		// Update
		if apiDef.APIID == "" {
			return errors.New("API ID must be set")
		}

		data, err := json.Marshal(apiDef.APIDefinition)
		if err != nil {
			return err
		}

		updatePath := urljoin.Join(c.url, endpointAPIs, apiDef.APIID)
		uResp, err := grequests.Put(updatePath, &grequests.RequestOptions{
			JSON: data,
			Headers: map[string]string{
				"x-tyk-authorization": c.secret,
				"content-type":        "application/json",
			},
			Params: map[string]string{
				"accept_additional_properties": "true",
			},
			InsecureSkipVerify: c.InsecureSkipVerify,
		})

		if err != nil {
			return err
		}

		if uResp.StatusCode != 200 {
			return fmt.Errorf("API updating returned error: %v (code: %v)", uResp.String(), uResp.StatusCode)
		}

		// initiate a reload
		go c.Reload()

		// Add updated API to existing API list.
		apiids[apiDef.APIID] = &apiDef
		ids[apiDef.Id.Hex()] = &apiDef
		slugs[apiDef.Slug] = &apiDef
		paths[apiDef.Proxy.ListenPath+"-"+apiDef.Domain] = &apiDef

		fmt.Printf("--> Status: OK, ID:%v\n", apiDef.APIID)
	}

	return nil
}

func (c *Client) SyncAPIs(apiDefs []objects.DBApiDefinition) error {
	deleteAPIs := []string{}
	updateAPIs := []objects.DBApiDefinition{}
	createAPIs := []objects.DBApiDefinition{}

	apis, err := c.FetchAPIs()
	if err != nil {
		return err
	}

	GWIDMap := map[string]int{}
	GitIDMap := map[string]int{}

	// Build the gw ID map
	for i, api := range apis {
		// Lets get a full list of existing IDs
		GWIDMap[api.APIID] = i
	}

	// Build the Git ID Map
	for i, def := range apiDefs {
		if def.APIID != "" {
			GitIDMap[def.APIID] = i
		} else {
			uid, err := uuid.NewV4()
			if err != nil {
				fmt.Println("error generating UUID", err)
				return err
			}
			created := fmt.Sprintf("temp-%v", uid.String())
			GitIDMap[created] = i
		}
	}

	// Updates are when we find items in git that are also in dash
	for key, index := range GitIDMap {
		_, ok := GWIDMap[key]
		if ok {
			updateAPIs = append(updateAPIs, apiDefs[index])
		}
	}

	// Deletes are when we find items in the dash that are not in git
	for key, _ := range GWIDMap {
		_, ok := GitIDMap[key]
		if !ok {
			deleteAPIs = append(deleteAPIs, key)
		}
	}

	// Create operations are when we find things in Git that are not in the dashboard
	for key, index := range GitIDMap {
		_, ok := GWIDMap[key]
		if !ok {
			createAPIs = append(createAPIs, apiDefs[index])
		}
	}

	fmt.Printf("Deleting: %v\n", len(deleteAPIs))
	fmt.Printf("Updating: %v\n", len(updateAPIs))
	fmt.Printf("Creating: %v\n", len(createAPIs))

	// Do the deletes
	for _, dbId := range deleteAPIs {
		fmt.Printf("SYNC Deleting: %v\n", dbId)
		if err := c.deleteAPI(dbId); err != nil {
			return err
		}
	}

	// Do the updates
	if err := c.UpdateAPIs(&updateAPIs); err != nil {
		return err
	}
	for _, apiDef := range updateAPIs {
		fmt.Printf("SYNC Updated: %v\n", apiDef.APIID)
	}

	// Do the creates
	if err := c.CreateAPIs(&createAPIs); err != nil {
		return err
	}
	for _, apiDef := range updateAPIs {
		fmt.Printf("SYNC Created: %v\n", apiDef.Name)
	}

	return nil
}

func (c *Client) DeleteAPI(id string) error {
	return c.deleteAPI(id)
}

func (c *Client) deleteAPI(id string) error {
	delPath := urljoin.Join(c.url, endpointAPIs)
	delPath += id

	delResp, err := grequests.Delete(delPath, &grequests.RequestOptions{
		Headers: map[string]string{
			"x-tyk-authorization": c.secret,
			"content-type":        "application/json",
		},
		InsecureSkipVerify: c.InsecureSkipVerify,
	})

	if err != nil {
		return err
	}

	if delResp.StatusCode != 200 {
		return fmt.Errorf("API Returned error: %v", delResp.String())
	}

	// initiate a reload
	go c.Reload()

	return nil
}
