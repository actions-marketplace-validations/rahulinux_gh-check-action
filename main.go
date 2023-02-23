package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/logutils"
)

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func findWorkflowFiles(workflow_dirs []string) ([]string, error) {
	var files []string
	for _, workflow_dir := range workflow_dirs {
		err := filepath.Walk(workflow_dir, func(path string, info fs.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() && (filepath.Ext(path) == ".yml" || filepath.Ext(path) == ".yaml") {
				files = append(files, path)
			}
			return err
		})
		if err != nil {
			msg := fmt.Errorf("[ERROR] Unable process files: %v", err.Error())
			log.Print(msg)
			return nil, err
		}
	}
	return files, nil
}

func cleanUp(temp_dirs []string) {
	for _, dir := range temp_dirs {
		os.RemoveAll(dir)
	}
}

func main() {
	var local bool
	var loglevel string
	var ignore_action string
	var workflow_dir string
	var token string
	var remote_repos string
	var file_dir string
	var workflow_dirs []string
	var temp_dirs []string
	action_updates := make(map[string]map[string]string)

	flag.BoolVar(&local, "local", false, "Use to run local mode")
	flag.StringVar(&workflow_dir, "workflow_dir", ".github/workflows", "Specify workflow dir")
	flag.StringVar(&loglevel, "loglevel", "info", "Specify log level: debug, warn, error")
	flag.StringVar(&ignore_action, "ignore_action", "", "comma separated list of actions to ignore")
	flag.StringVar(&remote_repos, "remote_repos", "", "comma separated list of remote repos")
	flag.StringVar(&token, "token", "", "Github auth token")
	flag.Parse()

	filter := &logutils.LevelFilter{
		Levels:   []logutils.LogLevel{"DEBUG", "WARN", "ERROR"},
		MinLevel: logutils.LogLevel(strings.ToUpper(loglevel)),
		Writer:   os.Stderr,
	}
	log.SetOutput(filter)

	if !local {
		log.Print("[INFO] CI mode")
		ignore_action = getEnv("INPUT_IGNOREACTIONS", "")
		workflow_dir = getEnv("INPUT_WORKFLOWDIR", "")
		loglevel = getEnv("INPUT_LOGLEVEL", "")
	} else {
		log.Print("[INFO] CI Local mode")
	}

	log.Printf("[DEBUG] loglevel: %v", loglevel)
	log.Printf("[DEBUG] ignore_action: %v", ignore_action)
	log.Printf("[DEBUG] workflow_dir: %v", workflow_dir)

	if len(remote_repos) > 0 {
		remote_repos := strings.Split(remote_repos, ",")
		for _, repo := range remote_repos {
			repo_name := strings.Replace(repo, "/", "-", -1)
			repo_name = fmt.Sprintf("*_%s", repo_name)
			temp_workflow_dir, err := os.MkdirTemp("", repo_name)
			if err != nil {
				log.Printf("[ERROR] Unable to create temp dir: %v", err.Error())
			}
			log.Printf("[DEBUG] Created temp dir: %v", temp_workflow_dir)
			workflow_dirs = append(workflow_dirs, temp_workflow_dir)
			temp_dirs = append(temp_dirs, temp_workflow_dir)
			storeRemoteWorkflows(token, repo, temp_workflow_dir)
		}
	}

	workflow_dirs = append(workflow_dirs, workflow_dir)
	files, err := findWorkflowFiles(workflow_dirs)

	log.Printf("[DEBUG] workflow files: %v", files)

	if err != nil {
		log.Printf("[ERROR] Unable process files: %v", err)
	}

	log.Printf("[INFO] processing workflow files: %v", files)

	for _, file := range files {
		for _, action := range getActions(file) {
			old_version := strings.Split(action, "@")[1]
			action_name := strings.Split(action, "@")[0]
			log.Printf("[DEBUG] Checking updates for action: %v", action_name)
			if new_version, ok := getLatestActionsTag(token, action_name); ok && len(new_version) > 0 {
				new_version = new_version[0:2]
				log.Printf("[DEBUG] Available version %v for action: %v", action_name, new_version)
				if old_version != new_version {
					updates := make(map[string]string)
					new_action := fmt.Sprintf("%s@%s", action_name, new_version)
					log.Printf("[DEBUG] Found new update for %s->%s\n", action, new_action)
					updates[action] = new_action
					if filepath.Dir(file) != workflow_dir {
						file_dir = filepath.Base(filepath.Dir(file))
						file_dir = strings.Split(file_dir, "_")[1]
					} else {
						file_dir = workflow_dir
					}
					action_updates[file_dir] = updates
				}
			}
		}
	}
	if len(action_updates) > 0 {
		output, err := json.MarshalIndent(action_updates, "", "    ")

		if err != nil {
			log.Printf("[ERROR] Something went wrong: %v", err.Error())
		}
		fmt.Println(fmt.Sprintf(`::set-output name=actions::%s`, output))
	}
	defer cleanUp(temp_dirs) // clean up
}
