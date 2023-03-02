package owntone

import (
	"fmt"
	"os"
	"strings"
)

const EnvUrl = "OWNTONE_URL"

type Client struct {
	url string
}

func CheckConfig() error {
	url := os.Getenv(EnvUrl)
	if len(url) == 0 {
		return fmt.Errorf("Not found \"OWNTONE_URL\". Please set OWNTONE_URL via Environment variable")
	}
	return nil
}

func (c Client) buildUrl(path string) string {
	return c.url + path
}

func NewOwntoneClient() Client {
	url := os.Getenv(EnvUrl)
	if strings.HasSuffix(url, "/") != true {
		url = url + "/"
	}
	return Client{
		url,
	}
}
