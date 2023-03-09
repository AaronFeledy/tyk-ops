package cli

import (
	"crypto/tls"
	"fmt"
	"github.com/AaronFeledy/tyk-ops/pkg/output"
	"github.com/go-resty/resty/v2"
	"github.com/json-iterator/go"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"strings"
)

// UserAgent is the user agent request header used for all requests.
// This can be useful in identifying our requests on the server side.
const UserAgent = "TykOps/" + VERSION

// restCmd defines the `tykops rest` CLI command
var restCmd = &cobra.Command{
	Use:     "rest [method] url",
	Short:   "Make a REST request",
	Long:    "Make a REST request to the specified URL. The method defaults to GET when not specified.",
	Example: RootCmd.Use + " rest GET http://example.com/api/endpoint -H 'Accept:application/json,Accept-Encoding:gzip'",
	Args:    cobra.MaximumNArgs(2),
	Run:     cmdRest,
}

// restOpt defines the flags for the `tykops login` CLI command
func restOpt() {
	// Flags that apply to this command
	restCmd.Flags().BoolP("insecure", "k", false, "override TLS certificate validation")
	restCmd.Flags().IntP("truncate", "t", 0, "truncate output to specified length")
	restCmd.Flags().Lookup("truncate").NoOptDefVal = "500"
	restCmd.Flags().StringSliceP("headers", "H", make([]string, 0), "add headers to the request")
	restCmd.Flags().Lookup("headers").NoOptDefVal = "key:value"
	restCmd.Flags().BoolP("server-response", "S", false, "show server response")
	restCmd.Flags().StringSliceP("values", "V", make([]string, 0), "get specific values from the response")
	restCmd.Flags().StringP("body", "b", "", "request body")

	// Binding flags to viper allows us to set values via config file or environment variables
	_ = viper.BindPFlag("insecure", restCmd.Flags().Lookup("insecure"))
}

func cmdRest(cmd *cobra.Command, args []string) {
	var method string
	var url string

	// The user can specify the method as the first argument, or it will default to GET
	switch len(args) {
	case 1:
		method = "GET"
		url = args[0]
	case 2:
		method = args[0]
		url = args[1]
	}

	// We need to parse the headers supplied by the user.
	// The Flags package doesn't support maps, so we use a slice of strings instead.
	headerSlice, err := cmd.Flags().GetStringSlice("headers")
	if err != nil {
		output.User.Error("error parsing supplied headers ", err)
		os.Exit(1)
	}
	// Headers are supplied as a slice of strings in the format "key:value"
	// We need to parse them into a map[string]string
	reqHeaders := make(map[string]string)
	for _, header := range headerSlice {
		parts := strings.Split(header, ":")
		if len(parts) != 2 {
			output.User.Error("error parsing supplied headers: use 'key:value' format separated by commas")
			os.Exit(1)
		}
		reqHeaders[parts[0]] = parts[1]
	}

	client := resty.New()

	allowInsecure, _ := cmd.Flags().GetBool("insecure")
	if allowInsecure {
		client.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	}

	req := client.R().
		SetHeader("User-Agent", UserAgent).
		SetHeaders(reqHeaders)
	body, _ := cmd.Flags().GetString("body")
	if body != "" {
		req.SetBody(body)
	}

	resp, err := req.Execute(method, url)
	if err != nil {
		output.User.Error("error making request: %s", err)
		os.Exit(1)
	}

	if !resp.IsSuccess() {
		output.User.Error(fmt.Sprintf("request failed: %v %s", resp.StatusCode(), resp.String()))
		os.Exit(resp.StatusCode())
	}

	outputObj := make(map[string]interface{})
	outputObj["output_key"] = "body"

	prepareResponse(cmd, resp, &outputObj)

	// If the user specified the --server-response flag, we print the response details
	if enabled, _ := cmd.Flags().GetBool("server-response"); enabled {
		// No key outputs the response details
		outputObj["output_key"] = ""
	}

	var sourceBytes []byte
	if outputObj["output_key"] == "" {
		sourceBytes, _ = jsoniter.Marshal(outputObj)
	} else {
		sourceBytes, _ = jsoniter.Marshal(outputObj[outputObj["output_key"].(string)])
	}

	// We need to parse the values supplied by the user.
	valueSlice, err := cmd.Flags().GetStringSlice("values")
	if err != nil {
		output.User.Error("error parsing supplied values ", err)
		os.Exit(1)
	}
	if len(valueSlice) > 0 {
		valueMap := make(map[string]interface{})
		for _, value := range valueSlice {
			parts := strings.Split(value, ":")
			if len(parts) == 2 {
				// Two parts means the user specified field labels for the output
				valueMap[parts[0]] = jsoniter.Get(sourceBytes, parts[1]).ToString()
			} else {
				// One part means the user didn't specify field labels, so we use the field path
				valueMap[parts[0]] = jsoniter.Get(sourceBytes, parts[0]).ToString()
			}
		}
		outputObj["fields"] = valueMap
		outputObj["output_key"] = "fields"
	}

	var jsonOut []byte
	if outputObj["output_key"] == "" {
		jsonOut, _ = jsoniter.Marshal(outputObj)
	} else {
		jsonOut, _ = jsoniter.Marshal(outputObj[outputObj["output_key"].(string)])
	}

	outString := string(jsonOut)
	truncate, _ := cmd.Flags().GetInt("truncate")
	if truncate > 0 && len(outString) > truncate {
		outString = outString[:truncate] + "..."
	}

	// Pretty printed output is easier to read
	output.PrettyString(outString)
}

// prepareResponse takes the response from the request and parses it into a map[string]interface{}
func prepareResponse(cmd *cobra.Command, resp *resty.Response, outputObj *map[string]interface{}) *map[string]interface{} {
	response := *outputObj
	response["status"] = resp.StatusCode()
	response["headers"] = resp.Header()

	// Try to parse the response body as JSON
	bodyObj := make(map[string]interface{})
	if err := jsoniter.Unmarshal(resp.Body(), &bodyObj); err == nil {
		response["body"] = bodyObj
	} else {
		// If it's not JSON, just get the raw string
		response["body"] = string(resp.Body())
	}

	return &response
}

// init allows us to set up our command before any other functions are called
func init() {
	restOpt()
	RootCmd.AddCommand(restCmd)
}
