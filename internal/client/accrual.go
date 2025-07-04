package client

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/tempizhere/gofemarket/internal/model"
)

// AccrualClient взаимодействует с системой начислений.
type AccrualClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewAccrualClient создает новый AccrualClient с пулом соединений.
func NewAccrualClient(baseURL string) *AccrualClient {
	return &AccrualClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:      10,
				IdleConnTimeout:   30 * time.Second,
				DisableKeepAlives: false,
			},
		},
	}
}

// GetAccrual получает информацию о начислении.
func (c *AccrualClient) GetAccrual(ctx context.Context, orderNumber string) (model.Order, time.Duration, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/api/orders/"+orderNumber, nil)
	if err != nil {
		return model.Order{}, 0, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return model.Order{}, 0, err
	}
	defer func() { _ = resp.Body.Close() }()

	switch resp.StatusCode {
	case http.StatusOK:
		var accrual struct {
			Order   string  `json:"order"`
			Status  string  `json:"status"`
			Accrual float64 `json:"accrual"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&accrual); err != nil {
			return model.Order{}, 0, err
		}
		return model.Order{
			Number:  accrual.Order,
			Status:  accrual.Status,
			Accrual: accrual.Accrual,
		}, 0, nil
	case http.StatusNoContent:
		return model.Order{}, 0, model.ErrOrderNotFound
	case http.StatusTooManyRequests:
		retryAfter := 60 * time.Second // Значение по умолчанию
		if retryHeader := resp.Header.Get("Retry-After"); retryHeader != "" {
			if seconds, err := strconv.Atoi(retryHeader); err == nil {
				retryAfter = time.Duration(seconds) * time.Second
			}
		}
		return model.Order{}, retryAfter, model.ErrTooManyRequests
	default:
		return model.Order{}, 0, model.ErrOrderNotFound
	}
}
