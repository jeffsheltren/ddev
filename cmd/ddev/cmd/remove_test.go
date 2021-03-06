package cmd

import (
	"testing"

	"github.com/drud/ddev/pkg/exec"
	asrt "github.com/stretchr/testify/assert"
)

// TestDevRemove runs `ddev rm` on the test apps
func TestDevRemove(t *testing.T) {
	assert := asrt.New(t)

	// Make sure we have running sites.
	addSites()

	for _, site := range DevTestSites {
		cleanup := site.Chdir()

		out, err := exec.RunCommand(DdevBin, []string{"remove"})
		assert.NoError(err, "ddev remove should succeed but failed, err: %v, output: %s", err, out)
		assert.Contains(out, "Successfully removed")

		cleanup()
	}
	// Now put the sites back together so other tests can use them.
	addSites()
}
