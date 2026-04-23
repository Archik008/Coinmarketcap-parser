package coinmarketcap

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
)

type fetchParams struct {
	path     string
	baseURL  string
	apiToken string
	client   *http.Client
}

func fetch[J any](ctx context.Context, p fetchParams, queueCh chan<- struct{}) (J, error) {
	var zero J

	endpoint, err := url.JoinPath(p.baseURL, p.path)
	if err != nil {
		return zero, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return zero, err
	}
	req.Header.Set("X-CMC_PRO_API_KEY", p.apiToken)
	req.Header.Set("Accept", "application/json")

	queueCh <- struct{}{}

	resp, err := p.client.Do(req)
	if err != nil {
		return zero, err
	}
	defer resp.Body.Close()

	var result J
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return zero, err
	}
	return result, nil
}

func fetchCryptoCurrency(ctx context.Context, params fetchParams, queueCh chan<- struct{}) (ListingsResponse, error) {

	targetUrl, err := url.JoinPath(params.baseURL, params.path)
	if err != nil {
		return ListingsResponse{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, targetUrl, nil)
	if err != nil {
		return ListingsResponse{}, err
	}

	q := url.Values{}

	q.Add("start", "1")
	q.Add("limit", "150")
	q.Add("sort", "market_cap")
	q.Add("sort_dir", "desc")
	q.Add("convert", "USD")
	q.Add("cryptocurrency_type", "coins")

	req.URL.RawQuery = q.Encode()
	req.Header.Set("X-CMC_PRO_API_KEY", params.apiToken)
	req.Header.Set("Accept", "application/json")

	queueCh <- struct{}{}

	resp, err := params.client.Do(req)
	if err != nil {
		return ListingsResponse{}, err
	}
	defer resp.Body.Close()

	var result ListingsResponse

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return ListingsResponse{}, err
	}

	return result, nil
}
