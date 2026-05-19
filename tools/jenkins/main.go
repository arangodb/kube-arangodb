// Jenkins job downloader.
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

type Build struct {
	Number      int        `json:"number"`
	DisplayName string     `json:"displayName"`
	Result      string     `json:"result"`
	Artifacts   []Artifact `json:"artifacts"`
}

type Artifact struct {
	FileName     string `json:"fileName"`
	RelativePath string `json:"relativePath"`
}

type client struct {
	baseURL  string
	user     string
	token    string
	http     *http.Client
	buildURL string
}

var (
	flagURL    string
	flagUser   string
	flagToken  string
	flagOutput string
)

func main() {
	cmd := &cobra.Command{
		Use:   "jenkins <job-name> [build-id]",
		Short: "Download Jenkins job output and artifacts",
		Long: `Download console output, test results, and artifacts from a Jenkins build.

The job name is the simple name as shown in the Jenkins UI (e.g. "gateway-matrix").
The build ID defaults to "lastBuild" if not specified.

Examples:
  jenkins gateway-matrix
  jenkins gateway-matrix 42
  jenkins --url https://jenkins.example.com gateway-matrix lastBuild`,
		Args: cobra.RangeArgs(1, 2),
		RunE: run,
	}

	cmd.Flags().StringVar(&flagURL, "url", "", "Jenkins base URL (or JENKINS_URL env)")
	cmd.Flags().StringVar(&flagUser, "user", "", "Jenkins username (or JENKINS_USER env)")
	cmd.Flags().StringVar(&flagToken, "token", "", "Jenkins API token (or JENKINS_TOKEN env)")
	cmd.Flags().StringVar(&flagOutput, "output", "./jobs", "Output directory")

	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func envOr(flag, env string) string {
	if flag != "" {
		return flag
	}
	return os.Getenv(env)
}

func run(cmd *cobra.Command, args []string) error {
	c := &client{
		baseURL: envOr(flagURL, "JENKINS_URL"),
		user:    envOr(flagUser, "JENKINS_USER"),
		token:   envOr(flagToken, "JENKINS_TOKEN"),
		http:    &http.Client{},
	}

	if c.baseURL == "" {
		return fmt.Errorf("Jenkins URL is required (--url or JENKINS_URL)")
	}
	if c.user == "" || c.token == "" {
		return fmt.Errorf("Jenkins credentials required (--user/--token or JENKINS_USER/JENKINS_TOKEN)")
	}

	c.baseURL = strings.TrimRight(c.baseURL, "/")

	jobName := args[0]
	buildID := "lastBuild"
	if len(args) > 1 {
		buildID = args[1]
	}

	c.buildURL = fmt.Sprintf("%s/job/%s/%s", c.baseURL, jobName, buildID)

	return c.download(jobName, flagOutput)
}

func (c *client) download(jobPath, outputDir string) error {
	fmt.Printf(">> Fetching build info from %s\n", c.buildURL)

	var build Build
	buildJSON, err := c.getJSON(c.buildURL+"/api/json", &build)
	if err != nil {
		return fmt.Errorf("failed to fetch build info: %w", err)
	}

	safeName := strings.ReplaceAll(jobPath, "/job/", "_")
	safeName = strings.ReplaceAll(safeName, "/", "_")
	dest := filepath.Join(outputDir, safeName, fmt.Sprintf("%d", build.Number))

	fmt.Printf(">> Build: %s (#%d) - %s\n", build.DisplayName, build.Number, build.Result)
	fmt.Printf(">> Output: %s\n", dest)

	if err := os.MkdirAll(dest, 0o755); err != nil {
		return err
	}

	if err := os.WriteFile(filepath.Join(dest, "build.json"), buildJSON, 0o644); err != nil {
		return err
	}
	fmt.Println("   Saved build.json")

	c.downloadFile(c.buildURL+"/consoleText", filepath.Join(dest, "console.txt"), "console.txt")
	c.downloadFile(c.buildURL+"/testReport/api/json?pretty=true", filepath.Join(dest, "test-report.json"), "test-report.json")

	fmt.Printf(">> Downloading %d artifact(s)...\n", len(build.Artifacts))
	for _, a := range build.Artifacts {
		artifactPath := filepath.Join(dest, "artifacts", a.RelativePath)
		if err := os.MkdirAll(filepath.Dir(artifactPath), 0o755); err != nil {
			fmt.Printf("   %s... MKDIR FAILED: %v\n", a.RelativePath, err)
			continue
		}
		c.downloadFile(c.buildURL+"/artifact/"+a.RelativePath, artifactPath, a.RelativePath)
	}

	fmt.Printf("\n>> Done. Contents of %s:\n", dest)
	filepath.Walk(dest, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		rel, _ := filepath.Rel(dest, path)
		fmt.Printf("   %s (%d bytes)\n", rel, info.Size())
		return nil
	})

	return nil
}

func (c *client) get(url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(c.user, c.token)
	return c.http.Do(req)
}

func (c *client) getJSON(url string, v interface{}) ([]byte, error) {
	resp, err := c.get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d from %s", resp.StatusCode, url)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return data, json.Unmarshal(data, v)
}

func (c *client) downloadFile(url, destPath, label string) {
	resp, err := c.get(url)
	if err != nil {
		fmt.Printf("   %s... FAILED: %v\n", label, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("   %s... not available (HTTP %d)\n", label, resp.StatusCode)
		return
	}

	f, err := os.Create(destPath)
	if err != nil {
		fmt.Printf("   %s... FAILED: %v\n", label, err)
		return
	}
	defer f.Close()

	n, err := io.Copy(f, resp.Body)
	if err != nil {
		fmt.Printf("   %s... FAILED: %v\n", label, err)
		return
	}

	fmt.Printf("   %s (%d bytes)\n", label, n)
}
