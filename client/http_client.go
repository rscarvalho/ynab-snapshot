package client

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"reflect"
	"time"
)

const baseURL = "https://api.youneedabudget.com/v1"

const CurrentMonth = "current"

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

type categoryChanResult struct {
	Category *Category
	Error    error
}

type categoryRef struct {
	Category *Category
	Receiver chan *categoryChanResult
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

func (client *ynabClientImpl) GetCategories(budgetID string, month string) ([]CategoryGroup, error) {
	apiResponse := struct {
		Data struct {
			CategoryGroups []CategoryGroup `json:"category_groups"`
		} `json:"data"`
	}{}

	if err := client.doGet(fmt.Sprintf("/budgets/%s/categories", budgetID), &apiResponse); err != nil {
		return nil, err
	}

	if month == CurrentMonth {
		return apiResponse.Data.CategoryGroups, nil
	} else {
		rateLimit := time.Tick(50 * time.Millisecond)

		var categoryRefs []categoryRef

		for i := range apiResponse.Data.CategoryGroups {
			group := &apiResponse.Data.CategoryGroups[i]
			for i := range group.Categories {
				category := &group.Categories[i]
				channel := make(chan *categoryChanResult)
				categoryID := category.ID
				go func() {
					<-rateLimit
					category, err := client.GetCategoryForMonth(budgetID, categoryID, month)
					channel <- &categoryChanResult{category, err}

					close(channel)
				}()
				categoryRefs = append(categoryRefs, categoryRef{category, channel})
			}
		}

		for len(categoryRefs) > 0 {
			received := false
			for i := range categoryRefs {
				ref := &categoryRefs[i]
				time.Sleep(50 * time.Millisecond)

				select {
				case cat := <-ref.Receiver:
					if cat.Error != nil {
						_, _ = fmt.Fprintf(os.Stderr, "error loading category %s: %v\n", ref.Category.ID, cat.Error)
						continue
					}

					if !reflect.DeepEqual(ref.Category, cat) {
						copyCategories(ref.Category, cat.Category)
					}

					received = true
					categoryRefs = append(categoryRefs[:i], categoryRefs[i+1:]...)
				default:
				}

				if received {
					break
				}
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
