package client

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

type YNABClient interface {
	GetBudgets() ([]Budget, error)
	GetCategories(budgetId string) ([]CategoryGroup, error)
}

type ynabClientImpl struct {
	httpClient *http.Client
	apiKey     string
}

const baseURL = "https://api.youneedabudget.com/v1"

func makeRequest(client *ynabClientImpl, method string, path string, body io.Reader) (*http.Request, error) {
	url := fmt.Sprintf("%s%s", baseURL, path)
	request, e := http.NewRequest(method, url, body)

	if e != nil {
		return nil, e
	}

	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", client.apiKey))
	return request, nil
}

func (client *ynabClientImpl) doRequest(path string, method string, body io.Reader, responseStruct interface{}) error {
	request, err := makeRequest(client, method, path, body)

	if err != nil {
		return err
	}

	response, err := client.httpClient.Do(request)
	if err != nil {
		return err
	}

	if response.StatusCode != 200 {
		bytes, err := ioutil.ReadAll(response.Body)

		var message string
		if err != nil {
			message = err.Error()
		} else {
			message = string(bytes)
		}

		return fmt.Errorf("Invalid response status %d %s: Body=%s", response.StatusCode, response.Status, message)
	}

	err = json.NewDecoder(response.Body).Decode(responseStruct)
	//b, _ := ioutil.ReadAll(response.Body)
	//fmt.Printf("body=%v", string(b))
	//err = json.Unmarshal(b, responseStruct)
	if err != nil {
		return err
	}
	err = response.Body.Close()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error closing response %s", err.Error())
	}
	return nil
}

func (client *ynabClientImpl) doGet(path string, responseStruct interface{}) error {
	return client.doRequest(path, "GET", nil, responseStruct)
}

func (client *ynabClientImpl) GetBudgets() ([]Budget, error) {
	apiResponse := struct {
		Data struct {
			Budgets []Budget `json:"budgets"`
		} `json:"data"`
	}{}

	if err := client.doGet("/budgets", &apiResponse); err != nil {
		return nil, err
	}

	return apiResponse.Data.Budgets, nil
}

func (client *ynabClientImpl) GetCategories(budgetId string) ([]CategoryGroup, error) {
	apiResponse := struct {
		Data struct {
			CategoryGroups []CategoryGroup `json:"category_groups"`
		} `json:"data"`
	}{}

	if err := client.doGet(fmt.Sprintf("/budgets/%s/categories", budgetId), &apiResponse); err != nil {
		return nil, err
	}

	return apiResponse.Data.CategoryGroups, nil
}

func NewClient(apiKey string) YNABClient {
	client := &ynabClientImpl{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: time.Second * 30,
		},
	}

	return client
}
