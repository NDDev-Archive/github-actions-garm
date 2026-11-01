package repositorycontract

import (
	"os"
	"sort"
	"testing"

	"github.com/NDDev-OpenNetwork/github-actions/internal/garmbootstrap"
	"gopkg.in/yaml.v3"
)

// githubHostedLabelsActionlintPredates are GitHub-hosted images this
// repository's workflows name that the pinned actionlint does not know yet
// (rhysd/actionlint#682). GitHub serves them, not the fleet, so they are the
// only labels allowed beside the published classes.
var githubHostedLabelsActionlintPredates = map[string]bool{
	"ubuntu-26.04":     true,
	"ubuntu-26.04-arm": true,
}

// Actionlint's runner vocabulary is a public API: a missing label makes a
// published class unusable, while an extra label lets workflows request a
// class the fleet can never serve. Hold both directions to one authority.
func TestActionlintLabelsExactlyMatchPublishedClasses(t *testing.T) {
	t.Parallel()
	raw, err := os.ReadFile(repositoryRoot + "/.github/actionlint.yaml")
	if err != nil {
		t.Fatal(err)
	}
	var config struct {
		SelfHostedRunner struct {
			Labels []string `yaml:"labels"`
		} `yaml:"self-hosted-runner"`
	}
	if err := yaml.Unmarshal(raw, &config); err != nil {
		t.Fatal(err)
	}
	declared := make([]string, 0, len(config.SelfHostedRunner.Labels))
	for _, label := range config.SelfHostedRunner.Labels {
		if !githubHostedLabelsActionlintPredates[label] {
			declared = append(declared, label)
		}
	}
	wanted := make([]string, 0, len(garmbootstrap.PublishedScaleSets()))
	for _, class := range garmbootstrap.PublishedScaleSets() {
		wanted = append(wanted, class.Name)
	}
	sort.Strings(declared)
	sort.Strings(wanted)
	if len(declared) != len(wanted) {
		t.Fatalf("actionlint declares %d runner labels, fleet publishes %d\ndeclared=%v\npublished=%v", len(declared), len(wanted), declared, wanted)
	}
	for index := range wanted {
		if declared[index] != wanted[index] {
			t.Fatalf("actionlint runner labels differ from the published fleet classes\ndeclared=%v\npublished=%v", declared, wanted)
		}
		if index > 0 && declared[index] == declared[index-1] {
			t.Fatalf("actionlint runner label %q is duplicated", declared[index])
		}
	}
}
