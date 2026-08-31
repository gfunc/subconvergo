package handler

import (
	"crypto/sha256"
	"crypto/subtle"
	"net/http"

	"github.com/gfunc/subconvergo/config"
	"github.com/gin-gonic/gin"
)

// API token gate.
//
// The protected endpoints (/readconf, /getruleset, /render, /flushcache)
// funnel through requireAPIToken so comparison and failure behavior live in
// one place. /getprofile is the exception: it does not use requireAPIToken
// and instead compares the request token against the per-profile
// profile_token (falling back to the global api_access_token) via tokenEqual
// directly. Two hardening rules:
//
//   - Fail closed: with no configured token the protected endpoints are
//     DISABLED and refuse every request. (An empty configured token must
//     never mean "allow everyone".)
//   - Constant time: comparison runs over fixed-length SHA-256 digests via
//     crypto/subtle so neither token length nor matching prefix leaks
//     through timing.

// requireAPIToken reports whether the request may proceed, writing the 403
// response itself when it may not.
func requireAPIToken(c *gin.Context) bool {
	configured := config.Global.Common.APIAccessToken
	if configured == "" {
		c.String(http.StatusForbidden, "Forbidden\n")
		return false
	}
	if !tokenEqual(c.Query("token"), configured) {
		c.String(http.StatusForbidden, "Forbidden\n")
		return false
	}
	return true
}

// tokenEqual compares two tokens in constant time over fixed-length digests.
// An empty configured token never matches anything (fail closed).
func tokenEqual(given, configured string) bool {
	if given == "" || configured == "" {
		return false
	}
	givenSum := sha256.Sum256([]byte(given))
	configuredSum := sha256.Sum256([]byte(configured))
	return subtle.ConstantTimeCompare(givenSum[:], configuredSum[:]) == 1
}
