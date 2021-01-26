package zenhub

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Service interface {
	GetIssueData(repoId int, issueNumber int) (*IssueData, error)
	doRequest(path string, response interface{}) error
	GetEpics(repoId int) (map[int]*EpicData, error)
}

const ZenHubApi = "https://api.zenhub.com"
const IssuesPath = "/p1/repositories/%d/issues/%d"
const EpicsListPath = "/p1/repositories/%d/epics"
const EpicPath = "/p1/repositories/%d/epics/%d"

type IssueData struct {
	IssueNumber int `json:"issue_number"`
	Estimate    IntValue
	IsEpic      bool `json:"is_epic"`
}

type EpicData struct {
	IssueNumber        int
	TotalEpicEstimates IntValue `json:"total_epic_estimates"`
	Estimate           IntValue
	Issues             []IssueData
}

type IntValue struct {
	Value uint
}

type epicListResponse struct {
	EpicIssues []struct {
		IssueNumber int `json:"issue_number"`
		RepoId      int `json:"repo_id"`
	} `json:"epic_issues"`
}

type Client struct {
	token string
}

func NewClient(token string) Service {
	return &Client{
		token: token,
	}
}

func (c *Client) GetEpics(repoId int) (map[int]*EpicData, error) {
	epicsList := &epicListResponse{}
	err := c.doRequest(fmt.Sprintf(EpicsListPath, repoId), epicsList)
	if err != nil {
		return nil, err
	}

	epics := map[int]*EpicData{}

	for _, epicIssue := range epicsList.EpicIssues {
		epic := &EpicData{}
		err := c.doRequest(fmt.Sprintf(EpicPath, repoId, epicIssue.IssueNumber), epic)
		if err != nil {
			return nil, err
		}
		epic.IssueNumber = epicIssue.IssueNumber
		epics[epicIssue.IssueNumber] = epic
	}

	return epics, nil
}

func (c *Client) GetIssueData(repoId int, issueNumber int) (*IssueData, error) {
	data := &IssueData{}
	err := c.doRequest(fmt.Sprintf(IssuesPath, repoId, issueNumber), data)
	if err != nil {
		return &IssueData{}, err
	} else {
		return data, nil
	}
}

func (c *Client) doRequest(path string, response interface{}) error {

	request, err := http.NewRequest("GET", ZenHubApi+path, nil)
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
