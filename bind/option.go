package bind

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
)

type option map[string]string

func (opt option) apply(o option) {
	for k, v := range o {
		opt[k] = v
	}
}

type Options map[string]option

func fetchFromUrl(path string) (Options, error) {
	resp, err := http.Get(path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	options := make(Options)
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&options); err != nil {
		return nil, err
	}
	return options, nil
}

func fetchFromFile(path string) (Options, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	options := make(Options)
	if err := json.Unmarshal(data, &options); err != nil {
		return nil, err
	}
	return options, nil
}

func GetOption(path string) (Options, error) {
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return fetchFromUrl(path)
	}
	return fetchFromFile(path)
}
