package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

func getLatestActionsTag(token, name string) (response string, ok bool) {
	// https://api.github.com/repos/actions/checkout/releases/latest
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", name)

	var bearer = "Bearer " + token

	req, err := http.NewRequest("GET", url, nil)
	if len(token) > 0 {
		req.Header.Add("Authorization", bearer)
	}
	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		log.Printf("[ERROR] Unable to fetch tags, error: %v", err.Error())
	}
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("[ERROR] Issue with API response: %v\n", string(body))
		return "", false
	}
	defer resp.Body.Close()
	var tag struct {
		TagName string `json:"tag_name"`
	}
	json.NewDecoder(resp.Body).Decode(&tag)
	return tag.TagName, true
}

func getWorkflowFiles(token, repo string) (response []string, ok bool) {
	var workflow_urls []string
	// https://api.github.com/repos/OWNER/REPO/actions/workflows
	url := fmt.Sprintf("https://api.github.com/repos/%s/actions/workflows", repo)
	var bearer = "Bearer " + token

	req, err := http.NewRequest("GET", url, nil)
	if len(token) > 0 {
		req.Header.Add("Authorization", bearer)
	}
	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		log.Printf("[ERROR] Unable to fetch tags, error: %v", err.Error())
	}
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("[ERROR] Issue with API response: %v\n", string(body))
		return nil, false
	}
	defer resp.Body.Close()
	var Workflows struct {
		Workflows []struct {
			Path string `json:"path"`
		} `json:"workflows"`
	}
	json.NewDecoder(resp.Body).Decode(&Workflows)

	for _, workflow := range Workflows.Workflows {
		// https://api.github.com/repos/{username}/{repository_name}/contents/{file_path}
		workflow_url := fmt.Sprintf("https://api.github.com/repos/%s/contents/%s", repo, workflow.Path)
		workflow_urls = append(workflow_urls, workflow_url)
	}
	return workflow_urls, true

}

func getUrlContent(url, token string) (io.ReadCloser, error) {
	var bearer = "Bearer " + token
	req, err := http.NewRequest("GET", url, nil)
	if len(token) > 0 {
		req.Header.Add("Authorization", bearer)
	}
	req.Header.Set("Accept", "application/vnd.github.raw")
	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		log.Fatalf("[ERROR] Unable to download file: %v", err.Error())
	}
	return resp.Body, nil
}

func storeRemoteWorkflows(token, repo, temp_workflow_dir string) {
	workflow_urls, ok := getWorkflowFiles(token, repo)
	if !ok {
		log.Fatalf("[ERROR] Unable to get workflow from repo: %v", repo)
	}

	for _, workflow_url := range workflow_urls {
		log.Printf("[DEBUG] Download workflow content from url: %v", workflow_url)
		content, err := getUrlContent(workflow_url, token)
		if err != nil {
			log.Fatalf("[ERROR] Unable to fetch content: %v", err.Error())
		}

		f, err := os.CreateTemp(temp_workflow_dir, "gh.*.yml")
		if err != nil {
			log.Fatalf("[ERROR] Create temp file: %v", err.Error())
		}

		defer f.Close()
		_, err = io.Copy(f, content)
		if err != nil {
			log.Fatalf("[Error] Unable to store content in temp file: %v", err.Error())
		}
	}
}
