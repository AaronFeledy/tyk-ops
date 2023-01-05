package mserv

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/levigross/grequests"
	"github.com/ongoingio/urljoin"
	"io"
	"mime/multipart"
	"net/textproto"
)

// Client manages the connection to mserv
type Client struct {
	// url is the mserv endpoint URL
	url string
	// secret is the mserv auth token
	secret string
	// InsecureSkipVerify is a flag that specifies if we should validate the
	// server's TLS certificate.
	InsecureSkipVerify bool
}

// BundlePushParams is used to pass request parameters to the BundlePush() method
type BundlePushParams struct {
	// APIID is an optional parameter that specifies the API ID to associate the bundle with.
	APIID *string
	// StoreOnly is an optional parameter that specifies if the bundle should be stored without processing.
	// TODO: This does not appear to actually do anything. Verify and remove.
	StoreOnly *bool
	// UploadFile is the zip file to upload
	UploadFile io.ReadCloser
	// ProgressWriter is an optional parameter that specifies a writer to write progress to. This is used by the CLI to
	// display a progress bar during upload.
	ProgressWriter io.Writer
}

// BundleData is the data returned by the BundlePush() method
type BundleData struct {
	Id string `json:"bundle_id"`
}

// Payload is the API response payload from mserv
type Payload struct {
	Error   string      `json:"Error,omitempty"`
	Payload interface{} `json:"Payload,omitempty"`
	Status  string      `json:"Status,omitempty"`
}

// NewMservClient creates a new mserv client.
func NewMservClient(url, secret string) (*Client, error) {
	return &Client{
		url:    url,
		secret: secret,
	}, nil
}

// BundlePush uploads a bundle file to mserv
func (c *Client) BundlePush(params *BundlePushParams) (*BundleData, error) {
	endpoint := urljoin.Join(c.url, "/api/mw")

	reqBodyBuffer := &bytes.Buffer{}
	reqBodyReader := bufio.NewReader(reqBodyBuffer)

	// We can optionally write data as it is read from the file to a progress writer. The CLI uses this to count bytes
	// and display a progress bar.
	if params.ProgressWriter != nil {
		reqBodyReader = bufio.NewReader(io.TeeReader(reqBodyBuffer, params.ProgressWriter))
	}

	multipartWriter := multipart.NewWriter(reqBodyBuffer)
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="%s"; filename="%s"`, "uploadfile", "bundle.zip"))
	h.Set("Content-Type", "application/zip")
	reqBodyPartWriter, err := multipartWriter.CreatePart(h)

	// Copying the file data to the multipart writer adds appropriate headers and boundaries to the body of the request.
	if _, err = io.Copy(reqBodyPartWriter, params.UploadFile); err != nil && err != io.EOF {
		return nil, err
	}
	err = params.UploadFile.Close()
	err = multipartWriter.Close()
	if err != nil {
		return nil, err
	}

	reqQueryParams := map[string]string{}
	reqQueryParams["store_only"] = fmt.Sprintf("%v", *params.StoreOnly)
	if params.APIID != nil && *params.APIID != "" {
		reqQueryParams["api_id"] = *params.APIID
	}

	// Do the request
	pushResp, err := grequests.Post(endpoint, &grequests.RequestOptions{
		Headers: map[string]string{
			"X-Api-Key":    c.secret,
			"Content-Type": multipartWriter.FormDataContentType(),
		},
		InsecureSkipVerify: c.InsecureSkipVerify,
		Params:             reqQueryParams,
		RequestBody:        reqBodyReader,
	})
	if err != nil {
		return nil, err
	}

	if pushResp.StatusCode != 200 {
		return nil, fmt.Errorf("mserv api returned error: %v (code: %v)", pushResp.Error, pushResp.StatusCode)
	}

	var payload Payload
	if err := pushResp.JSON(&payload); err != nil {
		return nil, err
	}

	if payload.Status != "ok" {
		return nil, fmt.Errorf("mserv api request completed, but with error: %v", payload.Error)
	}

	// Payload must be a map[string]interface{}
	payloadData, ok := payload.Payload.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("mserv api returned unexpected payload: %v", payload.Payload)
	}

	// Payload must contain a "BundleID" key
	bundleID, ok := payloadData["BundleID"].(string)
	if !ok {
		return nil, fmt.Errorf("mserv api returned unexpected payload: %v", payloadData)
	}

	return &BundleData{Id: bundleID}, nil
}
