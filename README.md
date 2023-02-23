# gh-check-action

This tool will scan workflow files and check for any new update available for
action.

also, you can run local cli to update workflow file for new action update. [TODO]

### INSTALLATION

Copy and paste the following snippet into your workflow `.yml` file.

```yaml
- name: Check action updates
  uses: rahulinux/gh-check-action@v1.0.0
```

### CLI

```
  -ignore_action string
    	comma separated list of actions to ignore
  -local
    	Use to run local mode
  -loglevel string
    	Specify log level: debug, warn, error (default "info")
  -prettyprint
    	Json prettyprint output (default true)
  -remote_repos string
    	comma separated list of remote repos
  -token string
    	Github auth token
  -workflow_dir string
    	Specify workflow dir (default ".github/workflows")
```


### Sample output

```
> go run .  -local -remote_repos rahulinux/echo-service  -token XXXX
2023/02/23 13:59:20 [INFO] CI Local mode
2023/02/23 13:59:21 [INFO] processing workflow files: [/tmp/2367665663_rahulinux-echo-service/gh.1159213085.yml /tmp/2367665663_rahulinux-echo-service/gh.1721764720.yml /tmp/2367665663_rahulinux-echo-service/gh.2086260401.yml .github/workflows/go.yml .github/workflows/test.yml]
::set-output name=actions::{
    ".github/workflows": {
        "actions/checkout@v2": "actions/checkout@v3"
    },
    "rahulinux-echo-service": {
        "actions/setup-python@v3": "actions/setup-python@v4"
    }
}
```
