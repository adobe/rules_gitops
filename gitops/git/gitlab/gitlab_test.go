package gitlab

import "testing"

func TestCreatePRRemote(t *testing.T) {
	t.Skip("Manual")
	var (
		testGitlabToken = "********"
	)
	accessToken = &testGitlabToken
	type args struct {
		from  string
		to    string
		title string
	}
	tests := []struct {
		repo    string
		args    args
		wantErr bool
	}{
		{
			repo: "cotocisternas/rules_gitops_gitlab_test",
			args: args{
				from:  "feature/gitlab-test",
				to:    "master",
				title: "test_gitlab",
			},
			wantErr: false,
		},
		{
			repo: "petabytecl/subgroup_rules_gitops_gitlab_test/rules_gitops_gitlab_test",
			args: args{
				from:  "feature/gitlab-test",
				to:    "master",
				title: "test_gitlab",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.repo, func(t *testing.T) {
			repo = &tt.repo
			if err := CreatePR(tt.args.from, tt.args.to, tt.args.title); (err != nil) != tt.wantErr {
				t.Errorf("CreatePR() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
