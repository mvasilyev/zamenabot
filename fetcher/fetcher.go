package fetcher

import (
	"encoding/csv"
	"fmt"
	"net/http"
)

type HTTPClient interface {
	Get(url string) (*http.Response, error)
}

type Fetcher struct {
	SheetID    string
	HTTPClient HTTPClient
}

func (fetcher *Fetcher) FetchSchedule() ([][]string, error) {
	url := fmt.Sprintf("https://docs.google.com/spreadsheets/d/%s/gviz/tq?tqx=out:csv", fetcher.SheetID)

	resp, err := fetcher.HTTPClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	reader := csv.NewReader(resp.Body)
	data, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	return data, nil
}
