package labeling

import (
	"testing"

	"github.com/google/go-github/v29/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const open = "open"

func closePR(pr pullRequest) pullRequest { pr.state = ""; return pr }

type pullRequest struct {
	title string
	state string
	files []string

	expectedLabels []string
	*github.PullRequest
}

var (
	prModifyAppsPlugin = pullRequest{
		title: "Modify apps.plugin",
		state: open,
		files: []string{"collectors/apps.plugin/apps_plugin.c"},
	}
	prModifyPythonExample = pullRequest{
		title: "Modify python.d example module",
		state: open,
		files: []string{"collectors/python.d.plugin/example/example.chart.py"},
	}
	prModifyPythonApache = pullRequest{
		title: "Modify python.d apache module",
		state: open,
		files: []string{"collectors/python.d.plugin/apache/apache.chart.py"},
	}
	prModifyBashExample = pullRequest{
		title: "Modify charts.d example module",
		state: open,
		files: []string{"collectors/charts.d.plugin/example/example.sh"},
	}
	prModifyBashApache = pullRequest{
		title: "Modify charts.d apache module",
		state: open,
		files: []string{"collectors/charts.d.plugin/apache/apache.sh"},
	}
	prClosedModifyBashTomcat = pullRequest{
		title: "Closed modify charts.d tomcat module",
		files: []string{"collectors/charts.d.plugin/tomcat/tomcat.sh"},
	}
)

func TestNew(t *testing.T) {

}

type applyLabelsTest struct {
	pr             pullRequest
	expectedLabels []string
}

func TestLabeler_ApplyLabels(t *testing.T) {
	tests := []applyLabelsTest{
		{pr: prModifyAppsPlugin, expectedLabels: []string{"collectors"}},
		{pr: prModifyPythonExample, expectedLabels: []string{"collectors", "python.d"}},
		{pr: prModifyPythonApache, expectedLabels: []string{"collectors", "python.d", "python.d/apache"}},
		{pr: prModifyBashExample, expectedLabels: []string{"collectors", "charts.d"}},
		{pr: prModifyBashApache, expectedLabels: []string{"collectors", "charts.d", "charts.d/apache"}},
		{pr: prClosedModifyBashTomcat},
	}

	labeler, _ := prepareApplyLabelsLabeler(tests)

	err := labeler.ApplyLabels()
	require.NoError(t, err)
	ensurePullRequestsHaveExpectedLabels(t, tests)
}

func TestLabeler_ApplyLabels_DoesntApplyLabelsInDryRunMode(t *testing.T) {
	tests := []applyLabelsTest{
		{pr: prModifyAppsPlugin},
		{pr: prModifyPythonExample},
		{pr: prModifyPythonApache},
		{pr: prModifyBashExample},
		{pr: prModifyBashApache},
	}

	labeler, _ := prepareApplyLabelsLabeler(tests)
	labeler.DryRun = true

	err := labeler.ApplyLabels()
	require.NoError(t, err)
	ensurePullRequestsHaveExpectedLabels(t, tests)
}

func TestLabeler_ApplyLabels_SuccessfulWhenZeroPullRequest(t *testing.T) {
	labeler, _ := prepareApplyLabelsLabeler(nil)

	assert.NoError(t, labeler.ApplyLabels())
}

func TestLabeler_ApplyLabels_SuccessfulWhenZeroOpenPullRequest(t *testing.T) {
	tests := []applyLabelsTest{
		{pr: closePR(prModifyAppsPlugin)},
		{pr: closePR(prModifyPythonExample)},
		{pr: closePR(prModifyPythonApache)},
		{pr: closePR(prModifyBashExample)},
		{pr: closePR(prModifyBashApache)},
	}

	labeler, _ := prepareApplyLabelsLabeler(tests)

	assert.NoError(t, labeler.ApplyLabels())
	ensurePullRequestsHaveExpectedLabels(t, tests)
}

func TestLabeler_ApplyLabels_ReturnsErrorIfOpenPullRequestsFails(t *testing.T) {
	labeler, rs := prepareApplyLabelsLabeler(nil)
	rs.errOnOpenPullRequests = true

	assert.Error(t, labeler.ApplyLabels())
}

func TestLabeler_ApplyLabels_ReturnsErrorIfPullRequestModifiedFilesFails(t *testing.T) {
	tests := []applyLabelsTest{
		{pr: prModifyAppsPlugin},
	}
	labeler, rs := prepareApplyLabelsLabeler(tests)
	rs.errOnPullRequestModifiedFiles = true

	assert.Error(t, labeler.ApplyLabels())
}

func TestLabeler_ApplyLabels_ReturnsErrorIfAddLabelsToPullRequestFails(t *testing.T) {
	tests := []applyLabelsTest{
		{pr: prModifyAppsPlugin},
	}
	labeler, rs := prepareApplyLabelsLabeler(tests)
	rs.errOnAddLabelsToPullRequest = true

	assert.Error(t, labeler.ApplyLabels())
}

func ensurePullRequestsHaveExpectedLabels(t *testing.T, tests []applyLabelsTest) {
	for _, test := range tests {
		if len(test.expectedLabels) > 0 {
			diff := difference(test.expectedLabels, test.pr.Labels)
			assert.Zerof(t, diff, "PR#%d ('%s') has no following labels: %v", *test.pr.Number, *test.pr.Title, diff)
		} else {
			assert.Zerof(t, test.pr.Labels, "PR#%d ('%s') has following labels: %v", *test.pr.Number, *test.pr.Title, test.pr.Labels)
		}
	}
}

func prepareApplyLabelsLabeler(cases []applyLabelsTest) (*Labeler, *mockRepository) {
	rs := prepareRepository()
	ms := prepareMappings()

	for i, c := range cases {
		cases[i].pr.expectedLabels = c.expectedLabels
		pull, files := convertPullRequest(c.pr)
		cases[i].pr.PullRequest = pull
		rs.addPullRequest(pull, files)
	}
	return New(rs, ms), rs
}

func convertPullRequest(pr pullRequest) (*github.PullRequest, []*github.CommitFile) {
	pull := &github.PullRequest{
		Title: &pr.title,
		State: &pr.state,
	}
	var cf []*github.CommitFile
	for _, name := range pr.files {
		name := name
		cf = append(cf, &github.CommitFile{Filename: &name})
	}
	return pull, cf
}
