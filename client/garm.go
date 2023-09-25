// SPDX-License-Identifier: MIT

package client

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"path"
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

func GetGithubToken(client HTTPClient, bearerToken, metadataURLString string) (string, error) {
	log.Printf("TOKEN: %s", bearerToken)

	metadataURL, err := url.Parse(metadataURLString)
	if err != nil {
		return "", err
	}
	metadataURL.Path = path.Join(metadataURL.Path, "runner-registration-token")

	req, err := http.NewRequest(http.MethodGet, metadataURL.String(), nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+bearerToken)

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	return string(bodyBytes), nil
}
