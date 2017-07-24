package e2e

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"net/http"
	"net/url"

	"github.com/google/go-github/github"
	_ "github.com/joho/godotenv/autoload"
	"github.com/marwan-at-work/baghdad/utils"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
)

var projectName = "baghdad-e2e-test"
var baseURL = os.Getenv("BAGHDAD_E2E_API_URL")
var githubUser = os.Getenv("BAGHDAD_E2E_GITHUB_USER")
var adminToken = os.Getenv("BAGHDAD_E2E_ADMIN_TOKEN")
var yes = true
var tempDir string

var projectURL = fmt.Sprintf(
	"https://github.com/%v/%v",
	githubUser,
	projectName,
)
var projectGitURL = fmt.Sprintf(
	"https://%v@github.com/%v/%v",
	adminToken,
	githubUser,
	projectName,
)

var tagName = "master-1.0.0-build.1"

func TestMain(m *testing.M) {
	code := m.Run()
	cleanupTests()
	os.Exit(code)
}

func TestInit(t *testing.T) {
	utils.ValidateEnvVars(getRequiredVars()...)
	testRoot(t)
}

func testRoot(t *testing.T) {
	resp, err := http.Get(baseURL)
	if err != nil {
		t.Fatalf("expected GET %v to not error out but got: %v", baseURL, err)
	} else if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected GET %v to equal 200 but got %v", baseURL, resp.StatusCode)
	}

	testRepoCreation(t)
}

func testRepoCreation(t *testing.T) {
	c := utils.GetGithub(adminToken)
	_, _, err := c.Repositories.Create(context.Background(), "", &github.Repository{
		Name:     &projectName,
		Private:  &yes,
		AutoInit: &yes,
	})
	if err != nil {
		t.Fatalf("expected repo creation to be succesful but got: %v", err)
	}

	testCloneRepo(t)
}

func testCloneRepo(t *testing.T) {
	gitURL, _ := url.Parse(projectGitURL)
	gitURL.User = url.UserPassword(adminToken, "x-oauth-basic")
	var err error
	tempDir, err = ioutil.TempDir("", "baghdad-e2e-test")
	if err != nil {
		t.Fatalf("could not create temp dir: %v", err)
	}

	_, err = git.PlainClone(tempDir, false, &git.CloneOptions{
		URL:               gitURL.String(),
		Progress:          os.Stdout,
		RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
		ReferenceName:     plumbing.ReferenceName("refs/heads/master"),
	})
	if err != nil {
		t.Fatal(err)
	}

	testCreateWebhook(t)
}

func testCreateWebhook(t *testing.T) {
	c := utils.GetGithub(adminToken)
	web := "web"
	_, _, err := c.Repositories.CreateHook(context.Background(), githubUser, projectName, &github.Hook{
		Name:   &web,
		Active: &yes,
		Events: []string{"yes"},
		Config: map[string]interface{}{
			"url":          filepath.Join(baseURL, "/hooks/github"),
			"content_type": "json",
		},
	})

	if err != nil {
		t.Fatalf("could not create webhook: %v", err)
	}
}

func cleanupTests() {
	c := utils.GetGithub(adminToken)
	_, err := c.Repositories.Delete(context.Background(), githubUser, projectName)
	if err != nil {
		fmt.Println(err)
	}

	if tempDir != "" {
		os.RemoveAll(tempDir)
	}
}
