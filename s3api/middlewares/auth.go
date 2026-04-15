package middlewares

import (
	"net/http"
	"strings"
)

const (
	AuthorizationHeader = "Authorization"
	AWSV4Prefix        = "AWS4-HMAC-SHA256"
	AWSV2Prefix        = "AWS "
)

// AuthResult holds the parsed authentication information from a request.
type AuthResult struct {
	AccessKey     string
	Signature     string
	SignedHeaders []string
	Region        string
	Service       string
}

// ParseAuthHeader extracts authentication details from the Authorization header.
// Note: Only AWS Signature Version 4 is supported; V2 is intentionally not implemented
// as it is deprecated and considered insecure.
func ParseAuthHeader(r *http.Request) (*AuthResult, error) {
	authHeader := r.Header.Get(AuthorizationHeader)
	if authHeader == "" {
		return nil, ErrMissingAuthHeader
	}

	if strings.HasPrefix(authHeader, AWSV4Prefix) {
		return parseV4Auth(authHeader)
	}

	return nil, ErrUnsupportedAuthMethod
}

func parseV4Auth(header string) (*AuthResult, error) {
	// AWS4-HMAC-SHA256 Credential=ACCESS/DATE/REGION/SERVICE/aws4_request, SignedHeaders=..., Signature=...
	parts := strings.SplitN(strings.TrimPrefix(header, AWSV4Prefix+" "), ", ", 3)
	if len(parts) != 3 {
		return nil, ErrMalformedAuthHeader
	}

	result := &AuthResult{}

	for _, part := range parts {
		part = strings.TrimSpace(part)
		switch {
		case strings.HasPrefix(part, "Credential="):
			cred := strings.TrimPrefix(part, "Credential=")
			fields := strings.Split(cred, "/")
			if len(fields) < 5 {
				return nil, ErrMalformedAuthHeader
			}
			result.AccessKey = fields[0]
			result.Region = fields[2]
			result.Service = fields[3]
		case strings.HasPrefix(part, "SignedHeaders="):
			headers := strings.TrimPrefix(part, "SignedHeaders=")
			result.SignedHeaders = strings.Split(headers, ";")
		case strings.HasPrefix(part, "Signature="):
			result.Signature = strings.TrimPrefix(part, "Signature=")
		}
	}

	// Both AccessKey and Signature are required; SignedHeaders absence is also suspicious
	// but we only enforce the two most critical fields here.
	if result.AccessKey == "" || result.Signature == "" {
		return nil, ErrMalformedAuthHeader
	}

	return result, nil
}
