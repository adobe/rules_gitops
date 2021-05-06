package gitlab

import (
	"errors"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/xanzy/go-gitlab"
)

var (
	gitlabHost  = flag.String("gitlab_host", "https://gitlab.com", "The host name of the gitlab instance")
	repo        = flag.String("gitlab_repo", "", "the repo to use for gitlab api requests")
	accessToken = flag.String("gitlab_access_token", os.Getenv("GITLAB_TOKEN"), "the access token to authenticate requests")
)

func CreatePR(from, to, title string) error {
	if *accessToken == "" {
		return errors.New("gitlab_access_token must be set")
	}

	opts := gitlab.CreateMergeRequestOptions{
		Title:              &title,
		Description:        nil,
		SourceBranch:       &from,
		TargetBranch:       &to,
		Labels:             nil,
		AssigneeID:         nil,
		AssigneeIDs:        nil,
		ReviewerIDs:        nil,
		TargetProjectID:    nil,
		MilestoneID:        nil,
		RemoveSourceBranch: nil,
		Squash:             nil,
		AllowCollaboration: nil,
	}

	gl, err := gitlab.NewClient(*accessToken, gitlab.WithBaseURL(*gitlabHost))
	if err != nil {
		return err
	}

	createdPr, resp, err := gl.MergeRequests.CreateMergeRequest(*repo, &opts)
	if err == nil {
		log.Println("Created MR: ", createdPr.WebURL)
		return nil
	}

	if resp.StatusCode == http.StatusConflict {
		// Handle the case: "Create MR" request fails because it already exists for this source branch
		log.Println("Reusing existing MR")
		return nil
	}

	// All other gitlab responses
	defer resp.Body.Close()
	body, readingErr := ioutil.ReadAll(resp.Body)
	if readingErr != nil {
		log.Println("cannot read response body")
	} else {
		log.Println("gitlab response: ", string(body))
	}

	return err
}
