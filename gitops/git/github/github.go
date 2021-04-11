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
	githubHost = flag.String("github_host", "", "The host name of the private enterprise github, e.g. git.corp.adobe.com")
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
	if *githubHost != "" {
		baseUrl := "https://" + *githubHost + "/api/v3/"
		uploadUrl := "https://" + *githubHost + "/api/uploads/"
		var err error
		gh, err = github.NewEnterpriseClient(baseUrl, uploadUrl, tc)
		if err != nil {
			log.Println("Error in creating github client", err)
			return nil
		}
	}

	pr := &github.NewPullRequest{
		Title:               &title,
		Head:                &from,
		Base:                &to,
		Body:                &title,
		Issue:               nil,
		MaintainerCanModify: new(bool),
		Draft:               new(bool),
	}
	createdPr, resp, err := gh.PullRequests.Create(ctx, *repoOwner, *repo, pr)
	if err == nil {
		// PR created
		log.Println("Created PR: ", *createdPr.URL)
	} else if 422 == resp.StatusCode {
		// Handle the case: "Create PR" request fails because it already exists
		log.Println("Reusing existing PR")
		err = nil
	} else {
		// All other github responses
		defer resp.Body.Close()
		body, readingErr := ioutil.ReadAll(resp.Body)
		if readingErr != nil {
			log.Println("cannot read response body")
		}
		log.Println("github response: ", string(body))
	}

	return err
}

