package main

import (
	"context"
	"github-to-csv/internal/zenhub"
	"github.com/google/go-github/v32/github"
	"github.com/jszwec/csvutil"
	"golang.org/x/oauth2"
	"log"
	"os"
	"strconv"
	"time"
)

type Issue struct {
	Number      *int
	Title       *string
	State       *string
	Milestone   *string
	Body        *string
	EpicTitle   *string
	EpicNumber  *int
	IsEpic      bool
	StoryPoints *uint
}

func main() {
	githubToken := os.Getenv("GITHUB_TOKEN")
	zenhubToken := os.Getenv("ZENHUB_TOKEN")
	zenhubRepoId, err := strconv.Atoi(os.Getenv("ZENHUB_REPO_ID"))
	if err != nil {
		log.Fatal(err)
	}

	zenhubClient := zenhub.NewClient(zenhubToken)

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: githubToken},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	var githubIssues []*github.Issue
	var page int = 0

	// list all repositories for the authenticated user
	for {
		issues, resp, err := client.Issues.ListByRepo(ctx, "CSU-Booking-Platform", "application", &github.IssueListByRepoOptions{
			State: "all",
			Since: time.Time{},
			ListOptions: github.ListOptions{
				Page:    page,
				PerPage: 1000,
			},
		})
		if err != nil {
			log.Fatal(err)
		}

		githubIssues = append(githubIssues, issues...)

		if resp.NextPage == 0 {
			break
		} else {
			page = resp.NextPage
		}

	}

	zenhubEpics, err := zenhubClient.GetEpics(zenhubRepoId)
	if err != nil {
		log.Fatal(err)
	}

	issueToEpicMap := makeIssueToEpicMap(githubIssues, zenhubEpics)

	var parsedIssues []Issue

	for _, issue := range githubIssues {
		if issue.IsPullRequest() {
			continue
		}

		zIssue, err := zenhubClient.GetIssueData(zenhubRepoId, *issue.Number)
		if err != nil {
			log.Fatal(err)
		}

		var epicTitle *string = nil
		var epicNumber *int = nil

		if issueToEpicMap[*issue.Number] != nil {
			epicTitle = issueToEpicMap[*issue.Number].Title
			epicNumber = issueToEpicMap[*issue.Number].Number
		}

		var milestoneTitle *string = nil
		if issue.Milestone != nil {
			milestoneTitle = issue.Milestone.Title
		}

		parsedIssues = append(parsedIssues, Issue{
			Number:    issue.Number,
			Title:     issue.Title,
			State:     issue.State,
			Milestone: milestoneTitle,
			//
			EpicTitle:  epicTitle,
			EpicNumber: epicNumber,
			IsEpic:     zenhubEpics[*issue.Number] != nil,
			//
			StoryPoints: &zIssue.Estimate.Value,
			Body:        issue.Body,
		})
	}

	output, err := csvutil.Marshal(parsedIssues)
	if err != nil {
		log.Fatal("error: ", err)
	}

	f, err := os.Create("./out.csv")
	if err != nil {
		log.Fatal("error: ", err)
	}
	defer f.Close()

	_, err = f.Write(output)
	if err != nil {
		log.Fatal("error: ", err)
	}

}

func makeIssueMap(githubIssues []*github.Issue) map[int]*github.Issue {
	issueMap := map[int]*github.Issue{}
	for _, issue := range githubIssues {
		issueMap[*issue.Number] = issue
	}
	return issueMap
}

func makeIssueToEpicMap(githubIssues []*github.Issue, zenhubEpics map[int]*zenhub.EpicData) map[int]*github.Issue {
	issueToEpicMap := map[int]*github.Issue{}

	githubIssueMap := makeIssueMap(githubIssues)

	for _, epic := range zenhubEpics {
		for _, issue := range epic.Issues {
			issueToEpicMap[issue.IssueNumber] = githubIssueMap[epic.IssueNumber]
		}
	}

	return issueToEpicMap
}
