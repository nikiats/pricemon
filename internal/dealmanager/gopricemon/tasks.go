package gopricemon

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

const (
	TaskStatusCompleted = "completed"
	TaskStatusFailed    = "failed"
)

type Task struct {
	ID          int        `json:"id"`
	Status      string     `json:"status"`
	Error       *string    `json:"error"`
	CompletedAt *time.Time `json:"completedAt"`
}

type TradeSettings struct {
	MaximumConcurrentTrades int `json:"maximumConcurrentTrades"`
}

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *Client) CreateTask(itemID, platformID int, actionType string, price decimal.Decimal, taskKey string) (int, error) {
	platformName, err := c.platformName(platformID)
	if err != nil {
		return 0, err
	}

	body, err := json.Marshal(struct {
		ItemID       int    `json:"itemID"`
		PlatformName string `json:"platformName"`
		ActionType   string `json:"actionType"`
		Price        string `json:"price"`
		TaskKey      string `json:"taskKey"`
	}{
		ItemID:       itemID,
		PlatformName: platformName,
		ActionType:   actionType,
		Price:        price.String(),
		TaskKey:      taskKey,
	})
	if err != nil {
		return 0, err
	}

	response, err := c.httpClient.Post(c.baseURL+"/tasks", "application/json", bytes.NewReader(body))
	if err != nil {
		return 0, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusCreated {
		return 0, responseError(response)
	}

	var task Task
	if err = json.NewDecoder(response.Body).Decode(&task); err != nil {
		return 0, err
	}

	return task.ID, nil
}

func (c *Client) GetTask(taskID int) (Task, error) {
	response, err := c.httpClient.Get(fmt.Sprintf("%s/tasks/%d", c.baseURL, taskID))
	if err != nil {
		return Task{}, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return Task{}, responseError(response)
	}

	var task Task
	if err = json.NewDecoder(response.Body).Decode(&task); err != nil {
		return Task{}, err
	}

	return task, nil
}

func (c *Client) GetTradeSettings() (TradeSettings, error) {
	response, err := c.httpClient.Get(c.baseURL + "/trade-settings")
	if err != nil {
		return TradeSettings{}, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return TradeSettings{}, responseError(response)
	}

	var settings TradeSettings
	if err = json.NewDecoder(response.Body).Decode(&settings); err != nil {
		return TradeSettings{}, err
	}
	if settings.MaximumConcurrentTrades < 1 {
		return TradeSettings{}, fmt.Errorf("gopricemon: invalid maximum concurrent trades")
	}

	return settings, nil
}

func (c *Client) platformName(platformID int) (string, error) {
	response, err := c.httpClient.Get(c.baseURL + "/platforms")
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return "", responseError(response)
	}

	var platforms []struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
	if err = json.NewDecoder(response.Body).Decode(&platforms); err != nil {
		return "", err
	}
	for _, platform := range platforms {
		if platform.ID == platformID {
			return platform.Name, nil
		}
	}

	return "", fmt.Errorf("platform %d not found", platformID)
}

func responseError(response *http.Response) error {
	var body struct {
		Error string `json:"error"`
	}
	if err := json.NewDecoder(response.Body).Decode(&body); err == nil && body.Error != "" {
		return fmt.Errorf("gopricemon: %s", body.Error)
	}

	return fmt.Errorf("gopricemon: %s", response.Status)
}
