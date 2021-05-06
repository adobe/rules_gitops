package github

import (
	"context"
	"errors"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/google/go-github/v32/github"
	"golang.org/x/oauth2"
)

var (
	repoOwner            = flag.String("github_repo_owner", "", "the owner user/organization to use for github api requests")
	repo                 = flag.String("github_repo", "", "the repo to use for github api requests")
	pat                  = flag.String("github_access_token", os.Getenv("GITHUB_TOKEN"), "the access token to authenticate requests")
	githubEnterpriseHost = flag.String("github_enterprise_host", "", "The host name of the private enterprise github, e.g. git.corp.adobe.com")
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

	var gh *github.Client
	if *githubEnterpriseHost != "" {
		baseUrl := "https://" + *githubEnterpriseHost + "/api/v3/"
		uploadUrl := "https://" + *githubEnterpriseHost + "/api/uploads/"
		var err error
		gh, err = github.NewEnterpriseClient(baseUrl, uploadUrl, tc)
		if err != nil {
			log.Println("Error in creating github client", err)
			return nil
		}
	} else {
		gh = github.NewClient(tc)
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
		log.Println("Created PR: ", *createdPr.URL)
		return err
	}

	if resp.StatusCode == http.StatusUnprocessableEntity {
		// Handle the case: "Create PR" request fails because it already exists
		log.Println("Reusing existing PR")
		return nil
	}

	// All other github responses
	defer resp.Body.Close()
	body, readingErr := ioutil.ReadAll(resp.Body)
	if readingErr != nil {
		log.Println("cannot read response body")
	} else {
		log.Println("github response: ", string(body))
	}

	return err
}
