package main

import (
	"context"
	"github-to-csv/internal/zenhub"
	"github.com/google/go-github/v32/github"
	"github.com/jszwec/csvutil"
	"golang.org/x/oauth2"
	"log"
	"os"
)

type Issue struct {
	Number      *int
	Title       *string
	State       *string
	Milestone   *string
	StoryPoints uint
}

func main() {
	githubToken := os.Getenv("GITHUB_TOKEN")
	zenhubToken := os.Getenv("ZENHUB_TOKEN")

	zenhubClient := zenhub.NewClient(zenhubToken)
	zenhubRepoId := int64(292935927)

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: githubToken},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	// list all repositories for the authenticated user
	sprint1Issues, _, _ := client.Issues.ListByRepo(ctx, "CSU-Booking-Platform", "application", &github.IssueListByRepoOptions{
		Milestone: "1",
		State:     "all",
	})

	var parsedIssues []Issue

	for _, issue := range sprint1Issues {
		zIssue, err := zenhubClient.GetIssueData(zenhubRepoId, *issue.Number)
		if err != nil {
			log.Fatal(err)
		}

		parsedIssues = append(parsedIssues, Issue{
			Number:      issue.Number,
			Title:       issue.Title,
			State:       issue.State,
			Milestone:   issue.Milestone.Title,
			StoryPoints: zIssue.Estimate.Value,
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
