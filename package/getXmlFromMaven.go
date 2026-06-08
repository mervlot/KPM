package install

import (
	"encoding/xml"
	"fmt"
	"io"
	"kpm/types"
	"net/http"
	"time"
)

var httpClient = &http.Client{
	Timeout: 15 * time.Second,
}

// GetMavenMetadata fetches and parses Maven metadata XML into types.Metadata
func GetMavenMetadata(url string) (*types.Metadata, error) {
	resp, err := httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status: %s", resp.Status)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read failed: %w", err)
	}

	var meta types.Metadata
	if err := xml.Unmarshal(data, &meta); err != nil {
		return nil, fmt.Errorf("xml unmarshal failed: %w", err)
	}

	return &meta, nil
}


func GetPomMetadata(url string) (*types.PomDependency, error) {
	resp, err := httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status: %s", resp.Status)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read failed: %w", err)
	}

	var meta types.PomDependency
	if err := xml.Unmarshal(data, &meta); err != nil {
		return nil, fmt.Errorf("xml unmarshal failed: %w", err)
	}

	return &meta, nil
}



func UnmarshalMavenXml(MetaXmlByte []byte) (*types.Metadata, error) {
	var meta types.Metadata
	if err := xml.Unmarshal(MetaXmlByte, &meta); err != nil {
		return nil, fmt.Errorf("xml unmarshal failed: %w", err)
	}

	return &meta, nil
}


func UnmarshalMavenPom(pomBytes []byte) (*types.PomDependency, error) {
	var meta types.PomDependency
	if err := xml.Unmarshal(pomBytes, &meta); err != nil {
		return nil, fmt.Errorf("xml unmarshal failed: %w", err)
	}

	return &meta, nil
}