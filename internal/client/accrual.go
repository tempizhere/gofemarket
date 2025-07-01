package client

import (
	"context"
	"encoding/json"
	"net/http"
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
func (c *AccrualClient) GetAccrual(ctx context.Context, orderNumber string) (model.Order, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/api/orders/"+orderNumber, nil)
	if err != nil {
		return model.Order{}, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return model.Order{}, err
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
			return model.Order{}, err
		}
		return model.Order{
			Number:  accrual.Order,
			Status:  accrual.Status,
			Accrual: accrual.Accrual,
		}, nil
	case http.StatusNoContent:
		return model.Order{}, model.ErrOrderNotFound
	case http.StatusTooManyRequests:
		return model.Order{}, model.ErrTooManyRequests
	default:
		return model.Order{}, model.ErrOrderNotFound
	}
}
