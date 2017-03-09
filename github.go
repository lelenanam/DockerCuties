package main

import (
	"context"
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

// Github provides credentials for accessing github
type Github struct {
	githubPersonalAccessToken string
}

// PullsSinceFunc applies function f to all pull requests starting from number since
func (g Github) PullsSinceFunc(since int, f func(*github.Issue) error) error {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: g.githubPersonalAccessToken},
	)
	tc := oauth2.NewClient(oauth2.NoContext, ts)
	client := github.NewClient(tc)

	last := since
	if since < 0 {
		since = 1
	}

	oldest, _, err := client.Issues.Get(context.Background(), Owner, Repo, since)
	if err != nil {
		return err
	}
	sinceDate := oldest.CreatedAt
	//Search pull requests created after since
	q := fmt.Sprintf("is:pr repo:%s/%s created:>=%s", Owner, Repo, sinceDate.Format("2006-01-02"))
	opt := &github.SearchOptions{
		Sort:        "created",
		Order:       "asc",
		ListOptions: github.ListOptions{Page: 1, PerPage: PerPage},
	}

	for {
		if opt.Page > SearchIssuesLimit/PerPage {
			//Start new search
			since = last
			oldest, _, err := client.Issues.Get(context.Background(), Owner, Repo, last)
			log.WithFields(log.Fields{"oldest": &oldest.Number}).Info("Search with new oldest")
			if err != nil {
				return err
			}
			sinceDate = oldest.CreatedAt
			log.WithFields(log.Fields{"date": sinceDate}).Info("New since")
			q = fmt.Sprintf("is:pr repo:%s/%s created:>=%s", Owner, Repo, sinceDate.Format("2006-01-02"))
			opt.Page = 1
		}
		pullreqs, _, err := client.Search.Issues(context.Background(), q, opt)
		if err != nil {
			return err
		}
		if len(pullreqs.Issues) == 0 {
			return nil
		}
		log.WithFields(log.Fields{"Page": opt.Page}).Info("Github pull requests")
		for _, pr := range pullreqs.Issues {
			//Skip numbers<since
			if *pr.Number < since {
				continue
			}
			log.WithFields(log.Fields{"number": *pr.Number, " title": *pr.Title}).Info("Pull request")
			if err := f(&pr); err != nil {
				return err
			}
		}
		last = *pullreqs.Issues[len(pullreqs.Issues)-1].Number + 1
		opt.Page++
	}
}
