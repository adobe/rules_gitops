package gitlab

import (
	"errors"
	"flag"
	"log"
	"os"

	"github.com/xanzy/go-gitlab"
)

var (
	gitlabHost = flag.String("gitlab_host", "gitlab.com", "The host name of the gitlab instance")
	repo       = flag.String("gitlab_repo", "", "the repo to use for gitlab api requests")
	pat        = flag.String("gitlab_access_token", os.Getenv("GITLAB_TOKEN"), "the access token to authenticate requests")
)

func CreatePR(from, to, title string) error {
	if *pat == "" {
		return errors.New("github_access_token must be set")
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

	gl, err := gitlab.NewClient("token", gitlab.WithBaseURL(*gitlabHost))
	if err != nil {
		return err
	}

	createdPr, _, err := gl.MergeRequests.CreateMergeRequest(*repo, &opts)
	if err != nil {
		return err
	}

	log.Println("Created MR: ", createdPr.WebURL)
	return nil
}
