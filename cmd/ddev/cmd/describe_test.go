package cmd

import (
	"strings"
	"testing"

	"encoding/json"

	"github.com/drud/ddev/pkg/ddevapp"
	"github.com/drud/ddev/pkg/exec"
	"github.com/drud/ddev/pkg/testcommon"
	"github.com/drud/ddev/pkg/util"
	log "github.com/sirupsen/logrus"
	asrt "github.com/stretchr/testify/assert"
)

// TestDescribeBadArgs ensures the binary behaves as expected when used with invalid arguments or working directories.
func TestDescribeBadArgs(t *testing.T) {
	assert := asrt.New(t)

	// Create a temporary directory and switch to it for the duration of this test.
	tmpdir := testcommon.CreateTmpDir("badargs")
	defer testcommon.CleanupDir(tmpdir)
	defer testcommon.Chdir(tmpdir)()

	// Ensure it fails if we run the vanilla describe outside of an application root.
	args := []string{"describe"}
	out, err := exec.RunCommand(DdevBin, args)
	assert.Error(err)
	assert.Contains(string(out), "Please specify a project name or change directories")

	// Ensure we get a failure if we run a describe on a named application which does not exist.
	args = []string{"describe", util.RandString(16)}
	out, err = exec.RunCommand(DdevBin, args)
	assert.Error(err)
	assert.Contains(string(out), "Unable to find any active project")

	// Ensure we get a failure if using too many arguments.
	args = []string{"describe", util.RandString(16), util.RandString(16)}
	out, err = exec.RunCommand(DdevBin, args)
	assert.Error(err)
	assert.Contains(string(out), "Too many arguments provided")

}

// TestDescribe tests that the describe command works properly when using the binary.
func TestDescribe(t *testing.T) {
	assert := asrt.New(t)

	for _, v := range DevTestSites {
		// First, try to do a describe from another directory.
		tmpdir := testcommon.CreateTmpDir("")
		cleanup := testcommon.Chdir(tmpdir)
		defer testcommon.CleanupDir(tmpdir)

		args := []string{"describe", v.Name}
		out, err := exec.RunCommand(DdevBin, args)
		assert.NoError(err)
		assert.Contains(string(out), "NAME")
		assert.Contains(string(out), "LOCATION")
		assert.Contains(string(out), v.Name)
		assert.Contains(string(out), "running")

		cleanup()

		cleanup = v.Chdir()

		args = []string{"describe"}
		out, err = exec.RunCommand(DdevBin, args)
		assert.NoError(err)
		assert.Contains(string(out), "NAME")
		assert.Contains(string(out), "LOCATION")
		assert.Contains(string(out), v.Name)
		assert.Contains(string(out), "running")

		// Test describe in current directory with json flag
		args = []string{"describe", "-j"}
		out, err = exec.RunCommand(DdevBin, args)
		assert.NoError(err)
		logItems, err := unmarshallJSONLogs(out)
		assert.NoError(err)

		// The description log should be the last item; there may be a warning
		// or other info before that.
		data := logItems[len(logItems)-1]
		assert.EqualValues(data["level"], "info")
		raw, ok := data["raw"].(map[string]interface{})
		assert.True(ok)
		assert.EqualValues(raw["status"], "running")
		assert.EqualValues(raw["name"], v.Name)
		assert.EqualValues(raw["shortroot"].(string), ddevapp.RenderHomeRootedDir(v.Dir))
		assert.EqualValues(raw["approot"].(string), v.Dir)

		assert.NotEmpty(data["msg"])

		cleanup()
	}
}

// TestDescribeAppFunction performs unit tests on the describeApp function from the working directory.
func TestDescribeAppFunction(t *testing.T) {
	assert := asrt.New(t)
	for _, v := range DevTestSites {
		cleanup := v.Chdir()

		app, err := ddevapp.GetActiveApp("")
		assert.NoError(err)

		desc, err := app.Describe()
		assert.NoError(err)
		assert.EqualValues(desc["status"], ddevapp.SiteRunning)
		assert.EqualValues(app.GetName(), desc["name"])
		assert.EqualValues(ddevapp.RenderHomeRootedDir(v.Dir), desc["shortroot"].(string))
		assert.EqualValues(v.Dir, desc["approot"].(string))

		out, _ := json.Marshal(desc)
		assert.Contains(string(out), app.GetHTTPURL())
		assert.Contains(string(out), app.GetName())
		assert.Contains(string(out), "\"router_status\":\"healthy\"")
		assert.Contains(string(out), ddevapp.RenderHomeRootedDir(v.Dir))

		// Stop the router using docker and then check the describe
		_, err = exec.RunCommand("docker", []string{"stop", "ddev-router"})
		assert.NoError(err)
		desc, err = app.Describe()
		assert.NoError(err)
		out, _ = json.Marshal(desc)
		assert.NoError(err)
		assert.Contains(string(out), "router_status\":\"exited")
		_, err = exec.RunCommand("docker", []string{"start", "ddev-router"})
		assert.NoError(err)

		cleanup()
	}
}

// TestDescribeAppUsingSitename performs unit tests on the describeApp function using the sitename as an argument.
func TestDescribeAppUsingSitename(t *testing.T) {
	assert := asrt.New(t)

	// Create a temporary directory and switch to it for the duration of this test.
	tmpdir := testcommon.CreateTmpDir("describeAppUsingSitename")
	defer testcommon.CleanupDir(tmpdir)
	defer testcommon.Chdir(tmpdir)()

	for _, v := range DevTestSites {
		app, err := ddevapp.GetActiveApp(v.Name)
		assert.NoError(err)
		desc, err := app.Describe()
		assert.NoError(err)
		assert.EqualValues(desc["status"], ddevapp.SiteRunning)
		assert.EqualValues(app.GetName(), desc["name"])
		assert.EqualValues(ddevapp.RenderHomeRootedDir(v.Dir), desc["shortroot"].(string))
		assert.EqualValues(v.Dir, desc["approot"].(string))

		out, _ := json.Marshal(desc)
		assert.NoError(err)
		assert.Contains(string(out), "running")
		assert.Contains(string(out), ddevapp.RenderHomeRootedDir(v.Dir))
	}
}

// TestDescribeAppWithInvalidParams performs unit tests on the describeApp function using a variety of invalid parameters.
func TestDescribeAppWithInvalidParams(t *testing.T) {
	assert := asrt.New(t)

	// Create a temporary directory and switch to it for the duration of this test.
	tmpdir := testcommon.CreateTmpDir("TestDescribeAppWithInvalidParams")
	defer testcommon.CleanupDir(tmpdir)
	defer testcommon.Chdir(tmpdir)()

	// Ensure describeApp fails from an invalid working directory.
	_, err := ddevapp.GetActiveApp("")
	assert.Error(err)

	// Ensure describeApp fails with invalid site-names.
	_, err = ddevapp.GetActiveApp(util.RandString(16))
	assert.Error(err)

	// Change to a site's working directory and ensure a failure still occurs with a invalid site name.
	cleanup := DevTestSites[0].Chdir()
	_, err = ddevapp.GetActiveApp(util.RandString(16))
	assert.Error(err)
	cleanup()
}

// unmarshallJSONLogs takes a string buffer and splits it into lines,
// discards empty lines, and unmarshalls into an array of logs
func unmarshallJSONLogs(in string) ([]log.Fields, error) {
	logData := make([]log.Fields, 0)
	logStrings := strings.Split(in, "\n")
	data := make(log.Fields, 4)

	for _, logLine := range logStrings {
		if logLine != "" {
			err := json.Unmarshal([]byte(logLine), &data)
			if err != nil {
				return []log.Fields{}, err
			}
			logData = append(logData, data)
		}
	}
	return logData, nil
}
