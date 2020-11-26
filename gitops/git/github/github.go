package github

import (
	"context"
	"errors"
	"flag"
	"github.com/google/go-github/v32/github"
	"golang.org/x/oauth2"
	"io/ioutil"
	"log"
	"os"
)

var (
	repoOwner = flag.String("github_repo_owner", "", "the owner user/organization to use for github api requests")
	repo = flag.String("github_repo", "", "the repo to use for github api requests")
	pat = flag.String("github_access_token", os.Getenv("GITHUB_TOKEN"), "the access token to authenticate requests")
)

func CreatePR(from, to, title string) error {
	if *repoOwner == "" {
		return errors.New("github_repo_owner must be set")
	}
	if *repo == "" {
		return errors.New("github_repo must be set")
	}
	if *pat == "" {
		return errors.New("github_access_token must be set")
	}

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: *pat},
	)
	tc := oauth2.NewClient(ctx, ts)
	gh := github.NewClient(tc)

	pr := &github.NewPullRequest{
		Title:               &title,
		Head:                &from,
		Base:                &to,
		Body:                &title,
		Issue:               nil,
		MaintainerCanModify: new(bool),
		Draft:               new(bool),
	}
	_, resp, err := gh.PullRequests.Create(ctx, *repoOwner, *repo, pr)
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	log.Print("github response: ", string(body))
	if 422 == resp.StatusCode {
		log.Print("Reusing existing PR")
		return nil
	}
	return err
}

