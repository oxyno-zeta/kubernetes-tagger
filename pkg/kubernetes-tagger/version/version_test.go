package version

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetVersionWithDefault(t *testing.T) {
	// Call code
	res := GetVersion()

	assert.Exactly(t, "-unreleased", res.Version)
	assert.Empty(t, res.GitCommit)
	assert.Empty(t, res.BuildDate)
}

func TestGetVersionNotReleased(t *testing.T) {
	Version = "version"
	GitCommit = "sha1"
	BuildDate = "04-03-2019"

	// Call code
	res := GetVersion()

	assert.Exactly(t, "version-unreleased", res.Version)
	assert.Exactly(t, "sha1", res.GitCommit)
	assert.Exactly(t, "04-03-2019", res.BuildDate)
}

func TestGetVersionReleased(t *testing.T) {
	Version = "version"
	GitCommit = "sha1"
	BuildDate = "04-03-2019"
	Metadata = ""

	// Call code
	res := GetVersion()

	assert.Exactly(t, "version", res.Version)
	assert.Exactly(t, "sha1", res.GitCommit)
	assert.Exactly(t, "04-03-2019", res.BuildDate)
}
