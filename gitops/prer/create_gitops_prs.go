/*
Copyright 2020 Adobe. All rights reserved.
This file is licensed to you under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License. You may obtain a copy
of the License at http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR REPRESENTATIONS
OF ANY KIND, either express or implied. See the License for the specific language
governing permissions and limitations under the License.
*/
package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	oe "os/exec"
	"strings"
	"sync"

	"github.com/adobe/rules_gitops/gitops/analysis"
	"github.com/adobe/rules_gitops/gitops/bazel"
	"github.com/adobe/rules_gitops/gitops/commitmsg"
	"github.com/adobe/rules_gitops/gitops/exec"
	"github.com/adobe/rules_gitops/gitops/git"
	"github.com/adobe/rules_gitops/gitops/git/bitbucket"
	"github.com/adobe/rules_gitops/gitops/git/github"
	"github.com/adobe/rules_gitops/gitops/git/gitlab"

	proto "github.com/golang/protobuf/proto"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

// SliceFlags should be used with flags.Var to define a command line flag with multiple values
type SliceFlags []string

func (i *SliceFlags) String() string {
	return "[" + strings.Join(*i, ",") + "]"
}

func (i *SliceFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

var (
	releaseBranch          = flag.String("release_branch", "master", "filter gitops targets by release branch")
	bazelCmd               = flag.String("bazel_cmd", "tools/bazel", "bazel binary to use")
	workspace              = flag.String("workspace", "", "path to workspace root")
	repo                   = flag.String("git_repo", "https://bitbucket.tubemogul.info/scm/tm/repo.git", "git repo location")
	gitMirror              = flag.String("git_mirror", "", "git mirror location, like /mnt/mirror/bitbucket.tubemogul.info/tm/repo.git for jenkins")
	gitopsPath             = flag.String("gitops_path", "cloud", "location to store files in repo.")
	gitopsTmpDir           = flag.String("gitops_tmpdir", os.TempDir(), "location to check out git tree with /cloud.")
	target                 = flag.String("target", "//... except //experimental/...", "target to scan. Useful for debugging only")
	pushParallelism        = flag.Int("push_parallelism", 5, "Number of image pushes to perform concurrently")
	prInto                 = flag.String("gitops_pr_into", "master", "use this branch as the source branch and target for deployment PR")
	prBody                 = flag.String("gitops_pr_body", "GitOps deployment <branch>", "a body message for deployment PR")
	prTitle                = flag.String("gitops_pr_title", "GitOps deployment <branch>", "a title for deployment PR")
	branchName             = flag.String("branch_name", "unknown", "Branch name to use in commit message")
	gitCommit              = flag.String("git_commit", "unknown", "Git commit to use in commit message")
	deploymentBranchSuffix = flag.String("deployment_branch_suffix", "", "suffix to add to all deployment branch names")
	gitHost                = flag.String("git_server", "bitbucket", "the git server api to use. 'bitbucket', 'github' or 'gitlab'")
	gitopsKind             SliceFlags
	gitopsRuleName         SliceFlags
	gitopsRuleAttr         SliceFlags
	dryRun                 = flag.Bool("dry_run", false, "Do not create PRs, just print what would be done")
)

func init() {
	flag.Var(&gitopsKind, "gitops_dependencies_kind", "dependency kind(s) to run during gitops phase. Can be specified multiple times. Default is 'k8s_container_push'")
	flag.Var(&gitopsRuleName, "gitops_dependencies_name", "dependency name(s) to run during gitops phase. Can be specified multiple times. Default is empty")
	flag.Var(&gitopsRuleAttr, "gitops_dependencies_attr", "dependency attribute(s) to run during gitops phase. Use attribute=value format. Can be specified multiple times. Default is empty")
}

func bazelQuery(query string) *analysis.CqueryResult {
	log.Println("Executing bazel cquery ", query)
	cmd := oe.Command(*bazelCmd, "cquery", query, "--output=proto")
	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Fatal(err)
	}
	go func() {
		io.Copy(os.Stderr, stderr)
	}()
	buildproto, err := cmd.Output()
	if err != nil {
		log.Fatal(err)
	}
	qr := &analysis.CqueryResult{}
	if err := proto.Unmarshal(buildproto, qr); err != nil {
		log.Fatal(err)
	}
	return qr
}

func main() {
	flag.Parse()
	if *workspace != "" {
		if err := os.Chdir(*workspace); err != nil {
			log.Fatal(err)
		}
	}
	if len(gitopsKind) == 0 {
		gitopsKind = []string{"k8s_container_push"}
	}

	var gitServer git.Server
	switch *gitHost {
	case "github":
		gitServer = git.ServerFunc(github.CreatePR)
	case "gitlab":
		gitServer = git.ServerFunc(gitlab.CreatePR)
	case "bitbucket":
		gitServer = git.ServerFunc(bitbucket.CreatePR)
	default:
		log.Fatalf("unknown vcs host: %s", *gitHost)
	}

	q := fmt.Sprintf("attr(deployment_branch, \".+\", attr(release_branch_prefix, \"%s\", kind(gitops, %s)))", *releaseBranch, *target)
	qr := bazelQuery(q)
	releaseTrains := make(map[string][]string)
	for _, t := range qr.Results {
		var releaseTrain string
		for _, a := range t.Target.GetRule().GetAttribute() {
			if a.GetName() == "deployment_branch" {
				releaseTrain = a.GetStringValue()
			}
		}
		releaseTrains[releaseTrain] = append(releaseTrains[releaseTrain], t.Target.Rule.GetName())
	}
	if (len(releaseTrains)) == 0 {
		log.Println("No matching targets found")
		return
	}

	for train, targets := range releaseTrains {
		fmt.Println(train)
		for _, t := range targets {
			fmt.Println(" ", t)
		}
	}

	gitopsdir, err := ioutil.TempDir(*gitopsTmpDir, "gitops")
	if err != nil {
		log.Fatalf("Unable to create tempdir in %s: %v", *gitopsTmpDir, err)
	}
	defer os.RemoveAll(gitopsdir)
	workdir, err := git.Clone(*repo, gitopsdir, *gitMirror, *prInto, *gitopsPath)
	if err != nil {
		log.Fatalf("Unable to clone repo: %v", err)
	}

	var updatedGitopsTargets []string
	var updatedGitopsBranches []string

	for train, targets := range releaseTrains {
		log.Println("train", train)
		branch := fmt.Sprintf("deploy/%s%s", train, *deploymentBranchSuffix)
		newBranch := workdir.SwitchToBranch(branch, *prInto)
		if !newBranch {
			// Find if we need to recreate the branch because target was deleted
			msg := workdir.GetLastCommitMessage()
			targetset := make(map[string]bool)
			for _, t := range targets {
				targetset[t] = true
			}
			oldtargets := commitmsg.ExtractTargets(msg)
			for _, t := range oldtargets {
				if !targetset[t] {
					// target t is not present in a new list
					workdir.RecreateBranch(branch, *prInto)
					break
				}
			}
		}
		for _, target := range targets {
			log.Println("train", train, "target", target)
			bin := bazel.TargetToExecutable(target)
			exec.Mustex("", bin, "--nopush", "--nobazel", "--deployment_root", gitopsdir)
		}
		if workdir.Commit(fmt.Sprintf("GitOps for release branch %s from %s commit %s\n%s", *releaseBranch, *branchName, *gitCommit, commitmsg.Generate(targets)), *gitopsPath) {
			log.Println("branch", branch, "has changes, push is required")
			updatedGitopsTargets = append(updatedGitopsTargets, targets...)
			updatedGitopsBranches = append(updatedGitopsBranches, branch)
		}
	}
	if len(updatedGitopsTargets) == 0 {
		log.Println("No gitops changes to push")
		return
	}

	// Push images

	// Create space separated set('//a' '//b' ... '//z') of targets.
	// Target names need to be quoted to protect from + and other special characters
	depsList := "set('" + strings.Join(updatedGitopsTargets, "' '") + "')"
	var qv []string
	for _, kind := range gitopsKind {
		q := fmt.Sprintf("kind(%s, deps(%s))", kind, depsList)
		qv = append(qv, q)
	}
	for _, name := range gitopsRuleName {
		q := fmt.Sprintf("filter(%s, deps(%s))", name, depsList)
		qv = append(qv, q)
	}
	for _, attr := range gitopsRuleAttr {
		name, value, found := strings.Cut(attr, "=")
		if !found {
			value = ".*"
		}
		q := fmt.Sprintf("attr(%s, %s, deps(%s))", name, value, depsList)
		qv = append(qv, q)
	}

	query := strings.Join(qv, " union ")
	qr = bazelQuery(query)
	targetsCh := make(chan string)
	var wg sync.WaitGroup
	wg.Add(*pushParallelism)
	for i := 0; i < *pushParallelism; i++ {
		go func() {
			defer wg.Done()
			for target := range targetsCh {
				bin := bazel.TargetToExecutable(target)
				exec.Mustex("", bin)
			}
		}()
	}
	for _, t := range qr.Results {
		targetsCh <- t.Target.Rule.GetName()
	}
	close(targetsCh)
	wg.Wait()

	if *dryRun {
		log.Println("dry-run: updated gitops branches: ", updatedGitopsBranches)
		log.Println("dry-run: skipping push")
	} else {
		workdir.Push(updatedGitopsBranches)
	}

	for _, branch := range updatedGitopsBranches {
		if *dryRun {
			log.Println("dry-run: skipping PR creation: branch ", branch, "into ", *prInto)
			continue
		}

		if *prTitle == "" {
			*prTitle = fmt.Sprintf("GitOps deployment %s", branch)
		}

		err := gitServer.CreatePR(branch, *prInto, *prTitle, *prBody)
		if err != nil {
			log.Fatal("unable to create PR: ", err)
		}
	}
}
