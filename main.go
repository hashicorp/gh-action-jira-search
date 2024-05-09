package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"

	"github.com/hashicorp/gh-action-jira/config"
	"github.com/hashicorp/gh-action-jira/gha"
	"github.com/hashicorp/gh-action-jira/jira"
)

func main() {
	err := search()
	if err != nil {
		log.Fatal(err)
	}
}

func search() error {
	jql := os.Getenv("INPUT_JQL")
	if jql == "" {
		return errors.New("no jql query provided as input")
	}
	config, err := config.ReadConfig()
	if err != nil {
		return err
	}

	issueKeys, err := findIssueKeys(config, jql)
	if err != nil {
		return err
	}
	if len(issueKeys) == 0 {
		fmt.Println("Successfully queried API but did not find any issues")
		return nil
	} else if len(issueKeys) > 1 {
		return errors.New("jql does not uniquely identify an issue")
	}

	key := issueKeys[0]
	fmt.Printf("Found issue %s\n", key)

	if err := gha.SetOutput("key", key); err != nil {
		return err
	}

	return nil
}

type searchResponse struct {
	Issues []struct {
		Key string `json:"key"`
	} `json:"issues"`
}

func findIssueKeys(config config.JiraConfig, jql string) ([]string, error) {
	query := url.Values{
		"jql":    {jql},
		"fields": {"summary"}, // Specify fields summary purely to minimise the size of all the unused fields in the response.
	}
	respBody, err := jira.DoRequest(config, "GET", "/rest/api/3/search", query, nil)
	if err != nil {
		return nil, err
	}

	var response searchResponse
	json.Unmarshal(respBody, &response)

	result := []string{}
	for _, issue := range response.Issues {
		result = append(result, issue.Key)
	}

	return result, nil
}
