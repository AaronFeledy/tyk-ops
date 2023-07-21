package rest

// UserAgent is the user agent request header used for all requests.
// This can be useful in identifying our requests on the server side.
var UserAgent = "TykOps/1.x-Dev"

// InitUserAgent sets the UserAgent variable to the given version.
func InitUserAgent(version string) {
	UserAgent = "TykOps/" + version
}
