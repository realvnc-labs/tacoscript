package utils

import (
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

type Location struct {
	IsURL       bool
	URL         *url.URL
	LocalPath   string
	RawLocation string
}

func ParseLocation(rawLocation string) Location {
	locationStr := ParseLocationOS(rawLocation)

	if !strings.HasPrefix(locationStr, "//") &&
		(strings.HasPrefix(locationStr, string(os.PathSeparator)) || filepath.IsAbs(locationStr)) {
		return Location{
			IsURL:       false,
			URL:         nil,
			LocalPath:   locationStr,
			RawLocation: rawLocation,
		}
	}

	u, err := url.Parse(locationStr)
	if err != nil {
		return Location{
			IsURL:       false,
			LocalPath:   locationStr,
			RawLocation: rawLocation,
		}
	}

	if u.Host != "" || u.Scheme != "" {
		return Location{
			IsURL:       true,
			URL:         u,
			RawLocation: rawLocation,
		}
	}

	return Location{
		IsURL:       false,
		URL:         nil,
		LocalPath:   locationStr,
		RawLocation: rawLocation,
	}
}
