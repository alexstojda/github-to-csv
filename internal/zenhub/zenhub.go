package zenhub

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Service interface {
	GetIssueData(repoId int64, issueNumber int) (*IssueData, error)
	doRequest(path string, response interface{}) error
}

const ZENHUB_API = "https://api.zenhub.com"
const ISSUES_PATH = "/p1/repositories/%d/issues/%d"

type IssueData struct {
	Estimate IntValue
	IsEpic   bool `json:"is_epic"`
}

type IntValue struct {
	Value uint
}

type Client struct {
	token string
}

func NewClient(token string) Service {
	return &Client{
		token: token,
	}
}

func (c *Client) GetIssueData(repoId int64, issueNumber int) (*IssueData, error) {
	data := &IssueData{}
	err := c.doRequest(fmt.Sprintf(ISSUES_PATH, repoId, issueNumber), data)
	if err != nil {
		return &IssueData{}, err
	} else {
		return data, nil
	}
}

func (c *Client) doRequest(path string, response interface{}) error {

	request, err := http.NewRequest("GET", ZENHUB_API+path, nil)
	if err != nil {
		return err
	}

	request.Header.Add("Content-Type", "application/json")
	request.Header.Add("X-Authentication-Token", c.token)

	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(response)
	if err != nil {
		return err
	}

	return nil
}
