package auth

import (
	"fmt"
	"net/http"
	"strings"
)

func GetBearerToken(headers http.Header) (string, error) {
	auth := headers.Get("Authorization")
	if auth == "" {
		return "", fmt.Errorf("No Authorization Token")
	}

	token := strings.TrimPrefix(auth, "Bearer ")

	return token, nil
}

func GetAPIKey(headers http.Header) (string, error) {
	auth := headers.Get("Authorization")
	if auth == "" {
		return "", fmt.Errorf("No Authorization Token")
	}

	token := strings.TrimPrefix(auth, "ApiKey ")

	return token, nil
}
