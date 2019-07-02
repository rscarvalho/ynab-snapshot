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

const baseURL = "https://api.youneedabudget.com/v1"

const CURRENT_MONTH = "Current Month"

// YNABClient is an HTTP client for the You Need a Budget REST api
type YNABClient interface {
	GetBudgets() ([]Budget, error)
	GetCategories(budgetID string, month string) ([]CategoryGroup, error)
}

// NewClient creates a new YNABClient with apiKey
func NewClient(apiKey string) YNABClient {
	client := &ynabClientImpl{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: time.Second * 30,
		},
	}

	return client
}

type ynabClientImpl struct {
	httpClient *http.Client
	apiKey     string
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

type categoryRef = struct {
	Category *Category
	Receiver chan *Category
}

func (client *ynabClientImpl) GetCategories(budgetID string, month string) ([]CategoryGroup, error) {
	apiResponse := struct {
		Data struct {
			CategoryGroups []CategoryGroup `json:"category_groups"`
		} `json:"data"`
	}{}

	if err := client.doGet(fmt.Sprintf("/budgets/%s/categories", budgetID), &apiResponse); err != nil {
		return nil, err
	}

	if month != CURRENT_MONTH {
		rateLimit := time.Tick(500 * time.Millisecond)
		<-rateLimit

		var categoryRefs []categoryRef

		for _, group := range apiResponse.Data.CategoryGroups {
			for _, category := range group.Categories {
				channel := make(chan *Category)
				categoryID := category.ID
				go func() {
					<-rateLimit
					category, err := client.GetCategoryForMonth(budgetID, categoryID, month)
					fmt.Printf("DONE: %s: %v\n", categoryID, err)
					if err != nil {
						channel <- nil
					} else {
						channel <- category
					}
				}()
				categoryRefs = append(categoryRefs, categoryRef{&category, channel})
			}
		}

		for _, ref := range categoryRefs {
			select {
			case cat := <-ref.Receiver:
				if cat == nil {
					_, _ = fmt.Fprintf(os.Stderr, "Error loading category %s", ref.Category.ID)
					continue
				}
				ref.Category.Activity = cat.Activity
				ref.Category.Balance = cat.Balance
				ref.Category.Budgeted = cat.Budgeted
				ref.Category.CategoryGroupID = cat.CategoryGroupID
				ref.Category.Deleted = cat.Deleted
				ref.Category.GoalCreationMonth = cat.GoalCreationMonth
				ref.Category.GoalPercentageComplete = cat.GoalPercentageComplete
				ref.Category.GoalTarget = cat.GoalTarget
				ref.Category.GoalTargetMonth = cat.GoalTargetMonth
				ref.Category.GoalType = cat.GoalType
				ref.Category.Hidden = cat.Hidden
				ref.Category.Name = cat.Name
				ref.Category.Note = cat.Note
				ref.Category.OriginalCategoryGroupID = cat.OriginalCategoryGroupID
			}
		}
	}

	return apiResponse.Data.CategoryGroups, nil
}

func (client *ynabClientImpl) GetCategoryForMonth(budgetID string, categoryID string, month string) (*Category, error) {
	apiResponse := struct {
		Data struct {
			Category *Category `json:"category"`
		} `json:"data"`
	}{}

	url := fmt.Sprintf("/budgets/%s/months/%s-01/categories/%s", budgetID, month, categoryID)
	if err := client.doGet(url, &apiResponse); err != nil {
		return nil, err
	}

	return apiResponse.Data.Category, nil
}

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

		return fmt.Errorf("invalid response status %d %s: Body=%s", response.StatusCode, response.Status, message)
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
