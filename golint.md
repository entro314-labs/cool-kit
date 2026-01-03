 Running [/home/runner/golangci-lint-2.7.2-linux-amd64/golangci-lint config path] in [/home/runner/work/cool-kit/cool-kit] ...
  Running [/home/runner/golangci-lint-2.7.2-linux-amd64/golangci-lint config verify] in [/home/runner/work/cool-kit/cool-kit] ...
  Running [/home/runner/golangci-lint-2.7.2-linux-amd64/golangci-lint run  --timeout=5m] in [/home/runner/work/cool-kit/cool-kit] ...
  Error: internal/config/project.go:85:11: Error return value of `os.Remove` is not checked (errcheck)
  	os.Remove(configDir) // Ignore error - directory might not be empty
  	         ^
  Error: internal/docker/build.go:49:14: Error return value of `os.Remove` is not checked (errcheck)
  				os.Remove(dockerfilePath)
  				         ^
  Error: internal/docker/build.go:53:13: Error return value of `os.Remove` is not checked (errcheck)
  			os.Remove(dockerfilePath)
  			         ^
  Error: internal/docker/push.go:35:11: Error return value of `cmd.StdoutPipe` is not checked (errcheck)
  		stdout, _ := cmd.StdoutPipe()
  		        ^
  Error: internal/docker/push.go:36:11: Error return value of `cmd.StderrPipe` is not checked (errcheck)
  		stderr, _ := cmd.StderrPipe()
  		        ^
  Error: internal/docker/push.go:94:11: Error return value of `cmd.StdoutPipe` is not checked (errcheck)
  		stdout, _ := cmd.StdoutPipe()
  		        ^
  Error: internal/docker/push.go:95:11: Error return value of `cmd.StderrPipe` is not checked (errcheck)
  		stderr, _ := cmd.StderrPipe()
  		        ^
  Error: internal/git/git.go:135:13: Error return value of `msgCmd.Output` is not checked (errcheck)
  	msgOutput, _ := msgCmd.Output()
  	           ^
  Error: internal/git/git.go:141:16: Error return value of `authorCmd.Output` is not checked (errcheck)
  	authorOutput, _ := authorCmd.Output()
  	              ^
  Error: internal/git/repo.go:100:11: Error return value of `cmd.StdoutPipe` is not checked (errcheck)
  		stdout, _ := cmd.StdoutPipe()
  		        ^
  Error: internal/git/repo.go:101:11: Error return value of `cmd.StderrPipe` is not checked (errcheck)
  		stderr, _ := cmd.StderrPipe()
  		        ^
  Error: internal/git/repo.go:188:11: Error return value of `cmd.StdoutPipe` is not checked (errcheck)
  		stdout, _ := cmd.StdoutPipe()
  		        ^
  Error: internal/git/repo.go:189:11: Error return value of `cmd.StderrPipe` is not checked (errcheck)
  		stderr, _ := cmd.StderrPipe()
  		        ^
  Error: internal/installer/azure.go:221:17: Error return value of `d.runAzCommand` is not checked (errcheck)
  		d.runAzCommand("network", "nsg", "rule", "create",
  		              ^
  Error: internal/mcp/server.go:45:6: Error return value is not checked (errcheck)
  	id, _ := req["id"].(float64)
  	    ^
  Error: internal/mcp/server.go:46:10: Error return value is not checked (errcheck)
  	method, _ := req["method"].(string)
  	        ^
  Error: internal/mcp/server.go:228:10: Error return value is not checked (errcheck)
  	params, _ := req["params"].(map[string]interface{})
  	        ^
  Error: internal/mcp/server.go:229:8: Error return value is not checked (errcheck)
  	name, _ := params["name"].(string)
  	      ^
  Error: internal/mcp/server.go:230:8: Error return value is not checked (errcheck)
  	args, _ := params["arguments"].(map[string]interface{})
  	      ^
  Error: internal/mcp/server.go:308:8: Error return value is not checked (errcheck)
  	uuid, _ := args["uuid"].(string)
  	      ^
  Error: internal/mcp/server.go:334:8: Error return value is not checked (errcheck)
  	uuid, _ := args["uuid"].(string)
  	      ^
  Error: internal/mcp/server.go:370:8: Error return value is not checked (errcheck)
  	uuid, _ := args["uuid"].(string)
  	      ^
  Error: internal/mcp/server.go:387:8: Error return value is not checked (errcheck)
  	uuid, _ := args["uuid"].(string)
  	      ^
  Error: internal/mcp/server.go:403:8: Error return value is not checked (errcheck)
  	uuid, _ := args["uuid"].(string)
  	      ^
  Error: internal/output/table.go:30:16: Error return value of `fmt.Fprintln` is not checked (errcheck)
  			fmt.Fprintln(f.opts.Writer)
  			            ^
  Error: internal/output/table.go:57:15: Error return value of `fmt.Fprintln` is not checked (errcheck)
  		fmt.Fprintln(w, "No data")
  		            ^
  Error: internal/output/table.go:70:15: Error return value of `fmt.Fprintf` is not checked (errcheck)
  			fmt.Fprintf(w, "%v\n", val.Index(i).Interface())
  			           ^
  Error: internal/output/table.go:79:14: Error return value of `fmt.Fprintln` is not checked (errcheck)
  	fmt.Fprintln(w, strings.Join(headersWithNum, "\t"))
  	            ^
  Error: internal/output/table.go:90:15: Error return value of `fmt.Fprintln` is not checked (errcheck)
  		fmt.Fprintln(w, strings.Join(rowWithNum, "\t"))
  		            ^
  Error: internal/output/table.go:99:14: Error return value of `fmt.Fprintln` is not checked (errcheck)
  	fmt.Fprintln(w, strings.Join(headers, "\t"))
  	            ^
  Error: internal/output/table.go:102:14: Error return value of `fmt.Fprintln` is not checked (errcheck)
  	fmt.Fprintln(w, strings.Join(row, "\t"))
  	            ^
  Error: internal/output/table.go:109:14: Error return value of `fmt.Fprintln` is not checked (errcheck)
  	fmt.Fprintln(w, "Key\tValue")
  	            ^
  Error: internal/output/table.go:115:14: Error return value of `fmt.Fprintf` is not checked (errcheck)
  		fmt.Fprintf(w, "%v\t%v\n", key.Interface(), f.formatValue(value))
  		           ^
  Error: internal/providers/azure/azure.go:291:17: Error return value of `p.runAzCommand` is not checked (errcheck)
  		p.runAzCommand("network", "nsg", "create",
  		              ^
  Error: internal/providers/azure/azure.go:300:18: Error return value of `p.runAzCommand` is not checked (errcheck)
  			p.runAzCommand("network", "nsg", "rule", "create",
  			              ^
  Error: internal/providers/azure/deploy.go:308:12: Error return value of `fmt.Sscanf` is not checked (errcheck)
  	fmt.Sscanf(backupNum, "%d", &selectedIndex)
  	          ^
  Error: internal/providers/azure/ssh.go:93:17: Error return value of `os.Remove` is not checked (errcheck)
  	defer os.Remove(tmpFile.Name())
  	               ^
  Error: internal/providers/azure/ssh.go:94:21: Error return value of `tmpFile.Close` is not checked (errcheck)
  	defer tmpFile.Close()
  	                   ^
  Error: internal/providers/baremetal/ssh.go:139:14: Error return value of `p.gitManager.GetLatestCommitInfo` is not checked (errcheck)
  	commitInfo, _ := p.gitManager.GetLatestCommitInfo()
  	            ^
  Error: internal/providers/baremetal/status.go:250:9: Error return value of `os.UserHomeDir` is not checked (errcheck)
  		home, _ := os.UserHomeDir()
  		      ^
  Error: internal/providers/digitalocean/ssh.go:145:12: Error return value of `fmt.Scanln` is not checked (errcheck)
  	fmt.Scanln(&confirm)
  	          ^
  Error: internal/providers/docker/docker.go:116:14: Error return value of `p.gitManager.GetLatestCommitInfo` is not checked (errcheck)
  	commitInfo, _ := p.gitManager.GetLatestCommitInfo()
  	            ^
  Error: internal/providers/docker/docker.go:141:14: Error return value of `cmd.Output` is not checked (errcheck)
  	dbPassword, _ := cmd.Output()
  	            ^
  Error: internal/providers/docker/docker.go:249:25: Error return value of `p.health.CheckWebSocket` is not checked (errcheck)
  	p.health.CheckWebSocket(wsURL, 5*time.Second)
  	                       ^
  Error: internal/providers/hetzner/ssh.go:175:12: Error return value of `fmt.Scanln` is not checked (errcheck)
  	fmt.Scanln(&confirm)
  	          ^
  Error: internal/templates/embed.go:41:9: Error return value of `filepath.Rel` is not checked (errcheck)
  			rel, _ := filepath.Rel("data", path)
  			     ^
  Error: internal/ui/ui.go:357:14: Error return value of `fmt.Fprintln` is not checked (errcheck)
  	fmt.Fprintln(l.writer, DimStyle.Render(msg))
  	            ^
  Error: internal/ui/ui.go:361:12: Error return value of `fmt.Fprint` is not checked (errcheck)
  	fmt.Fprint(l.writer, msg)
  	          ^
  Error: internal/utils/healthcheck.go:136:18: Error return value of `h.CheckWebSocket` is not checked (errcheck)
  	h.CheckWebSocket(wsURL, 10*time.Second)
  	                ^
  Error: internal/utils/logger.go:105:23: Error return value of `l.logFile.WriteString` is not checked (errcheck)
  	l.logFile.WriteString(logLine)
  	                     ^
  Error: cmd/login.go:35:1: cognitive complexity 40 of func `runLogin` is high (> 30) (gocognit)
  func runLogin(cmd *cobra.Command, args []string) error {
  ^
  Error: cmd/reset.go:23:1: cognitive complexity 44 of func `runReset` is high (> 30) (gocognit)
  func runReset(cmd *cobra.Command, args []string) error {
  ^
  Error: cmd/rollback.go:20:1: cognitive complexity 37 of func `runRollback` is high (> 30) (gocognit)
  func runRollback(cmd *cobra.Command, args []string) error {
  ^
  Error: internal/api/client.go:141:1: cognitive complexity 32 of func `(*Client).doWithRetry` is high (> 30) (gocognit)
  func (c *Client) doWithRetry(ctx context.Context, method, urlStr string, body []byte, v interface{}) error {
  ^
  Error: internal/detect/detector.go:81:1: cognitive complexity 38 of func `detectNodeProject` is high (> 30) (gocognit)
  func detectNodeProject(dir string) (*FrameworkInfo, error) {
  ^
  Error: internal/docker/build.go:29:1: cognitive complexity 39 of func `Build` is high (> 30) (gocognit)
  func Build(opts *BuildOptions) (err error) {
  ^
  Error: cmd/destroy.go:52:2: ifElseChain: rewrite if-else to switch statement (gocritic)
  	if len(args) > 0 {
  	^
  Error: internal/api/client.go:83:3: assignOp: replace `baseURL = baseURL + "/api/v1"` with `baseURL += "/api/v1"` (gocritic)
  		baseURL = baseURL + "/api/v1"
  		^
  Error: internal/azureconfig/validation.go:253:2: singleCaseSwitch: should rewrite switch statement to if statement (gocritic)
  	switch field {
  	^
  Error: internal/detect/detector.go:107:2: ifElseChain: rewrite if-else to switch statement (gocritic)
  	if isBun {
  	^
  Error: internal/providers/digitalocean/client.go:81:10: appendAssign: append result not assigned to the same slice (gocritic)
  	tags := append(opts.Tags, "coolify", "managed-by-cool-kit")
  	        ^
  Error: internal/ui/config_menu.go:72:2: singleCaseSwitch: should rewrite switch statement to if statement (gocritic)
  	switch msg := msg.(type) {
  	^
  Error: internal/azureconfig/validation.go:20:1: cyclomatic complexity 22 of func `(*Config).Validate` is high (> 20) (gocyclo)
  func (c *Config) Validate() error {
  ^
  Error: internal/local/setup.go:96:1: cyclomatic complexity 21 of func `collectConfiguration` is high (> 20) (gocyclo)
  func collectConfiguration() (*Config, error) {
  ^
  Error: internal/ui/progress.go:207:1: cyclomatic complexity 23 of func `(ProgressModel).Update` is high (> 20) (gocyclo)
  func (m ProgressModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
  ^
  Error: cmd/ci.go:63:12: G301: Expect directory permissions to be 0750 or less (gosec)
  	if err := os.MkdirAll(workflowDir, 0755); err != nil {
  	          ^
  Error: internal/appdeploy/setup.go:538:9: G306: Expect WriteFile permissions to be 0600 or less (gosec)
  	return os.WriteFile(readmePath, []byte(content), 0644)
  	       ^
  Error: internal/azureconfig/loader.go:63:12: G301: Expect directory permissions to be 0750 or less (gosec)
  	if err := os.MkdirAll(dir, 0755); err != nil {
  	          ^
  Error: internal/azureconfig/loader.go:74:12: G306: Expect WriteFile permissions to be 0600 or less (gosec)
  	if err := os.WriteFile(path, data, 0644); err != nil {
  	          ^
  Error: internal/config/config.go:151:12: G301: Expect directory permissions to be 0750 or less (gosec)
  	if err := os.MkdirAll(configDir, 0755); err != nil {
  	          ^
  Error: internal/config/global.go:49:12: G301: Expect directory permissions to be 0750 or less (gosec)
  	if err := os.MkdirAll(configDir, 0755); err != nil {
  	          ^
  Error: internal/config/project.go:44:12: G301: Expect directory permissions to be 0750 or less (gosec)
  	if err := os.MkdirAll(configDir, 0755); err != nil {
  	          ^
  Error: internal/config/project.go:55:12: G306: Expect WriteFile permissions to be 0600 or less (gosec)
  	if err := os.WriteFile(configPath, data, 0644); err != nil {
  	          ^
  Error: internal/docker/build.go:37:18: G306: Expect WriteFile permissions to be 0600 or less (gosec)
  		if writeErr := os.WriteFile(tempDockerfilePath, []byte(content), 0644); writeErr != nil {
  		               ^
  Error: internal/git/git.go:70:12: G301: Expect directory permissions to be 0750 or less (gosec)
  	if err := os.MkdirAll(parentDir, 0755); err != nil {
  	          ^
  Error: internal/installer/coolify.go:123:12: G301: Expect directory permissions to be 0750 or less (gosec)
  	if err := os.MkdirAll(tmpDir, 0755); err != nil {
  	          ^
  Error: internal/installer/coolify.go:151:12: G302: Expect file permissions to be 0600 or less (gosec)
  	if err := os.Chmod(scriptPath, 0755); err != nil {
  	          ^
  Error: internal/local/compose.go:141:12: G306: Expect WriteFile permissions to be 0600 or less (gosec)
  	if err := os.WriteFile(composePath, []byte(content), 0644); err != nil {
  	          ^
  Error: internal/local/config.go:98:12: G301: Expect directory permissions to be 0750 or less (gosec)
  	if err := os.MkdirAll(dir, 0755); err != nil {
  	          ^
  Error: internal/local/config.go:109:12: G306: Expect WriteFile permissions to be 0600 or less (gosec)
  	if err := os.WriteFile(path, data, 0644); err != nil {
  	          ^
  Error: internal/local/env.go:96:12: G301: Expect directory permissions to be 0750 or less (gosec)
  	if err := os.MkdirAll(workDir, 0755); err != nil {
  	          ^
  Error: internal/local/env.go:101:12: G306: Expect WriteFile permissions to be 0600 or less (gosec)
  	if err := os.WriteFile(envPath, []byte(content), 0644); err != nil {
  	          ^
  Error: internal/local/env.go:126:13: G301: Expect directory permissions to be 0750 or less (gosec)
  		if err := os.MkdirAll(path, 0755); err != nil {
  		          ^
  Error: internal/providers/azure/azure.go:362:13: G306: Expect WriteFile permissions to be 0600 or less (gosec)
  		if err := os.WriteFile(cloudInitFile, []byte(cloudInit), 0644); err != nil {
  		          ^
  Error: internal/providers/azure/azure.go:462:10: G204: Subprocess launched with a potential tainted input or cmd arguments (gosec)
  		cmd := exec.Command("ssh",
  			"-o", "StrictHostKeyChecking=no",
  			"-o", "ConnectTimeout=10",
  			"-o", "BatchMode=yes",
  			"-o", "UserKnownHostsFile=/dev/null",
  			"-o", "LogLevel=ERROR",
  			fmt.Sprintf("%s@%s", p.config.Azure.AdminUsername, publicIP),
  			"sudo docker ps 2>/dev/null")
  Error: internal/providers/azure/azure.go:499:13: G204: Subprocess launched with a potential tainted input or cmd arguments (gosec)
  	curlCmd := exec.Command("curl", "-f", "-s", "-o", "/dev/null", "--max-time", "10",
  		fmt.Sprintf("http://%s:8000", publicIP))
  Error: internal/providers/azure/provision.go:110:9: G204: Subprocess launched with a potential tainted input or cmd arguments (gosec)
  	cmd := exec.Command("az", "group", "create",
  		"--name", p.ctx.ResourceGroup,
  		"--location", p.ctx.Location,
  		"--output", "none")
  Error: internal/providers/azure/provision.go:128:13: G204: Subprocess launched with a potential tainted input or cmd arguments (gosec)
  	vnetCmd := exec.Command("az", "network", "vnet", "create",
  		"--resource-group", p.ctx.ResourceGroup,
  		"--name", p.ctx.VNetName,
  		"--address-prefix", p.ctx.Config.Networking.VNetAddressPrefix,
  		"--subnet-name", p.ctx.SubnetName,
  		"--subnet-prefix", p.ctx.Config.Networking.SubnetAddressPrefix,
  		"--output", "none")
  Error: internal/providers/azure/provision.go:149:12: G204: Subprocess launched with a potential tainted input or cmd arguments (gosec)
  	nsgCmd := exec.Command("az", "network", "nsg", "create",
  		"--resource-group", p.ctx.ResourceGroup,
  		"--name", p.ctx.NSGName,
  		"--output", "none")
  Error: internal/providers/azure/provision.go:171:14: G204: Subprocess launched with a potential tainted input or cmd arguments (gosec)
  		ruleCmd := exec.Command("az", "network", "nsg", "rule", "create",
  			"--resource-group", p.ctx.ResourceGroup,
  			"--nsg-name", p.ctx.NSGName,
  			"--name", rule.name,
  			"--protocol", "tcp",
  			"--direction", "inbound",
  			"--source-address-prefix", "*",
  			"--source-port-range", "*",
  			"--destination-address-prefix", "*",
  			"--destination-port-range", rule.port,
  			"--access", "allow",
  			"--priority", fmt.Sprintf("%d", rule.priority),
  			"--output", "none")
  Error: internal/providers/azure/provision.go:198:9: G204: Subprocess launched with a potential tainted input or cmd arguments (gosec)
  	cmd := exec.Command("az", "network", "public-ip", "create",
  		"--resource-group", p.ctx.ResourceGroup,
  		"--name", p.ctx.PublicIPName,
  		"--sku", "Standard",
  		"--allocation-method", "Static",
  		"--output", "none")
  Error: internal/providers/azure/provision.go:217:9: G204: Subprocess launched with a potential tainted input or cmd arguments (gosec)
  	cmd := exec.Command("az", "network", "nic", "create",
  		"--resource-group", p.ctx.ResourceGroup,
  		"--name", p.ctx.NICName,
  		"--vnet-name", p.ctx.VNetName,
  		"--subnet", p.ctx.SubnetName,
  		"--public-ip-address", p.ctx.PublicIPName,
  		"--network-security-group", p.ctx.NSGName,
  		"--output", "none")
  Error: internal/providers/azure/sdk_client.go:168:43: G115: integer overflow conversion int -> int32 (gosec)
  				Priority:                 to.Ptr(int32(100 + i*10)),
  				                                      ^
  Error: internal/providers/baremetal/ssh.go:281:9: G204: Subprocess launched with a potential tainted input or cmd arguments (gosec)
  	return exec.Command("ssh", fmt.Sprintf("%s@%s", user, host), command)
  	       ^
  Error: internal/providers/docker/docker.go:172:12: G306: Expect WriteFile permissions to be 0600 or less (gosec)
  	if err := os.WriteFile(envFile, []byte(envContent), 0644); err != nil {
  	          ^
  Error: internal/providers/production/production.go:85:9: G204: Subprocess launched with variable (gosec)
  	cmd := exec.Command("ssh", args...)
  	       ^
  Error: internal/providers/production/production.go:106:9: G204: Subprocess launched with variable (gosec)
  	cmd := exec.Command("ssh", args...)
  	       ^
  Error: internal/providers/production/production.go:115:8: G204: Subprocess launched with variable (gosec)
  	cmd = exec.Command("ssh", args...)
  	      ^
  Error: internal/providers/production/production.go:124:8: G204: Subprocess launched with variable (gosec)
  	cmd = exec.Command("ssh", args...)
  	      ^
  Error: internal/providers/production/production.go:145:9: G204: Subprocess launched with variable (gosec)
  	cmd := exec.Command("ssh", args...)
  	       ^
  Error: internal/providers/production/production.go:155:8: G204: Subprocess launched with variable (gosec)
  	cmd = exec.Command("ssh", args...)
  	      ^
  Error: internal/providers/production/production.go:196:9: G204: Subprocess launched with variable (gosec)
  	cmd := exec.Command("ssh", args...)
  	       ^
  Error: internal/providers/production/production.go:219:9: G204: Subprocess launched with variable (gosec)
  	cmd := exec.Command("ssh", args...)
  	       ^
  Error: internal/providers/production/production.go:231:9: G204: Subprocess launched with variable (gosec)
  		cmd = exec.Command("ssh", args...)
  		      ^
  Error: internal/providers/production/production.go:263:9: G204: Subprocess launched with variable (gosec)
  	cmd := exec.Command("ssh", args...)
  	       ^
  Error: internal/templates/embed.go:26:12: G306: Expect WriteFile permissions to be 0600 or less (gosec)
  	if err := os.WriteFile(targetPath, content, 0644); err != nil {
  	          ^
  Error: internal/utils/healthcheck.go:93:9: G204: Subprocess launched with a potential tainted input or cmd arguments (gosec)
  	cmd := exec.Command("ssh",
  		"-o", "ConnectTimeout=10",
  		"-o", "BatchMode=yes",
  		fmt.Sprintf("%s@%s", user, host),
  		"echo 'SSH OK'")
  Error: internal/utils/logger.go:23:18: G302: Expect file permissions to be 0600 or less (gosec)
  		logFile, err = os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
  		               ^
  Error: cmd/health.go:161:3: ineffectual assignment to statusDisplay (ineffassign)
  		statusDisplay := c.status
  		^
  Error: internal/providers/production/production.go:253:3: ineffectual assignment to sslEmail (ineffassign)
  		sslEmail = "admin@" + domain
  		^
  Error: cmd/badge.go:91:11: `Cancelled` is a misspelling of `Canceled` (misspell)
  		ui.Dim("Cancelled - no UUID provided")
  		        ^
  Error: cmd/deploy.go:104:23: `cancelled` is a misspelling of `canceled` (misspell)
  			ui.Dim("Deployment cancelled")
  			                   ^
  Error: cmd/destroy.go:109:13: `Cancelled` is a misspelling of `Canceled` (misspell)
  			ui.Info("Cancelled")
  			         ^
  Error: cmd/env.go:199:11: `Cancelled` is a misspelling of `Canceled` (misspell)
  		ui.Dim("Cancelled")
  		        ^
  Error: cmd/env.go:278:12: `Cancelled` is a misspelling of `Canceled` (misspell)
  			ui.Dim("Cancelled")
  			        ^
  Error: cmd/env.go:365:11: `Cancelled` is a misspelling of `Canceled` (misspell)
  		ui.Dim("Cancelled")
  		        ^
  Error: cmd/github.go:131:18: `Cancelled` is a misspelling of `Canceled` (misspell)
  				fmt.Println("Cancelled")
  				             ^
  Error: cmd/instances.go:194:11: `Cancelled` is a misspelling of `Canceled` (misspell)
  		ui.Dim("Cancelled")
  		        ^
  Error: cmd/link.go:37:12: `Cancelled` is a misspelling of `Canceled` (misspell)
  			ui.Dim("Cancelled")
  			        ^
  Error: cmd/logout.go:24:11: `Cancelled` is a misspelling of `Canceled` (misspell)
  		ui.Dim("Cancelled")
  		        ^
  Error: cmd/reset.go:61:22: `cancelled` is a misspelling of `canceled` (misspell)
  		return fmt.Errorf("cancelled")
  		                   ^
  Error: cmd/reset.go:70:22: `cancelled` is a misspelling of `canceled` (misspell)
  		return fmt.Errorf("cancelled")
  		                   ^
  Error: cmd/rollback.go:154:11: `Cancelled` is a misspelling of `Canceled` (misspell)
  		ui.Dim("Cancelled")
  		        ^
  Error: internal/appdeploy/watcher.go:185:27: `cancelled` is a misspelling of `canceled` (misspell)
  	case "failed", "error", "cancelled":
  	                         ^
  Error: internal/local/setup.go:296:19: `cancelled` is a misspelling of `canceled` (misspell)
  			ui.Info("Reset cancelled")
  			               ^
  Error: internal/providers/azure/deploy.go:33:23: `cancelled` is a misspelling of `canceled` (misspell)
  		ui.Info("Deployment cancelled")
  		                    ^
  Error: internal/providers/azure/deploy.go:320:21: `cancelled` is a misspelling of `canceled` (misspell)
  		ui.Info("Rollback cancelled")
  		                  ^
  Error: internal/providers/azure/deploy.go:401:19: `cancelled` is a misspelling of `canceled` (misspell)
  			ui.Info("Reset cancelled")
  			               ^
  Error: internal/providers/digitalocean/ssh.go:147:31: `cancelled` is a misspelling of `canceled` (misspell)
  		return fmt.Errorf("deletion cancelled")
  		                            ^
  Error: internal/providers/hetzner/ssh.go:177:31: `cancelled` is a misspelling of `canceled` (misspell)
  		return fmt.Errorf("deletion cancelled")
  		                            ^
  Error: internal/api/deployments.go:179:2: Consider pre-allocating `lines` (prealloc)
  	var lines []string
  	^
  Error: internal/mcp/server.go:291:2: Consider pre-allocating `result` (prealloc)
  	var result []map[string]interface{}
  	^
  Error: internal/mcp/server.go:459:2: Consider pre-allocating `result` (prealloc)
  	var result []map[string]interface{}
  	^
  Error: internal/providers/azure/backup.go:146:2: Consider pre-allocating `backups` (prealloc)
  	var backups []BackupInfo
  	^
  Error: internal/providers/digitalocean/client.go:75:2: Consider pre-allocating `sshKeys` (prealloc)
  	var sshKeys []godo.DropletCreateSSHKey
  	^
  Error: internal/service/server.go:62:2: Consider pre-allocating `newInstances` (prealloc)
  	var newInstances []config.Instance
  	^
  Error: internal/ui/main_menu.go:153:2: Consider pre-allocating `bounds` (prealloc)
  	var bounds []int
  	^
  Error: internal/api/application_getters.go:1:1: package-comments: should have a package comment (revive)
  package api
  ^
  Error: internal/api/client.go:1:9: var-naming: avoid meaningless package names (revive)
  package api
          ^
  Error: internal/api/client.go:27:6: exported: type name will be used as api.APIError by other packages, and that stutters; consider calling this Error (revive)
  type APIError struct {
       ^
  Error: internal/azureconfig/config.go:1:1: package-comments: should have a package comment (revive)
  package azureconfig
  ^
  Error: internal/config/config.go:1:1: package-comments: should have a package comment (revive)
  package config
  ^
  Error: internal/detect/detector.go:1:1: package-comments: should have a package comment (revive)
  package detect
  ^
  Error: internal/mcp/server.go:1:1: package-comments: should have a package comment (revive)
  package mcp
  ^
  Error: internal/mcp/server.go:17:6: exported: type name will be used as mcp.MCPServer by other packages, and that stutters; consider calling this Server (revive)
  type MCPServer struct {
       ^
  Error: internal/output/formatter.go:1:1: package-comments: should have a package comment (revive)
  package output
  ^
  Error: internal/output/table.go:20:1: exported: exported method TableFormatter.Format should have comment or be unexported (revive)
  func (f *TableFormatter) Format(data any) (err error) {
  ^
  Error: internal/smart/detector.go:1:1: package-comments: should have a package comment (revive)
  package smart
  ^
  Error: internal/smart/detector.go:40:6: exported: type name will be used as smart.SmartDetector by other packages, and that stutters; consider calling this Detector (revive)
  type SmartDetector struct {
       ^
  Error: internal/templates/embed.go:1:1: package-comments: should have a package comment (revive)
  package templates
  ^
  Error: internal/ui/config_menu.go:1:1: package-comments: should have a package comment (revive)
  package ui
  ^
  Error: internal/ui/model.go:55:2: exported: exported const StateSelectProvider should have comment (or a comment on this block) or be unexported (revive)
  	StateSelectProvider State = iota
  	^
  Error: internal/ui/model.go:363:1: exported: comment on exported type ProgressMsg should be of the form "ProgressMsg ..." (with optional leading article) (revive)
  // Message types
  ^
  Error: internal/ui/model.go:369:6: exported: exported type ErrorMsg should have comment or be unexported (revive)
  type ErrorMsg struct {
       ^
  Error: internal/ui/model.go:373:6: exported: exported type CompleteMsg should have comment or be unexported (revive)
  type CompleteMsg struct{}
       ^
  Error: internal/ui/ui.go:80:1: exported: exported function Print should have comment or be unexported (revive)
  func Print(msg string) {
  ^
  Error: internal/ui/ui.go:85:1: exported: exported function Success should have comment or be unexported (revive)
  func Success(msg string) {
  ^
  Error: internal/ui/ui.go:90:1: exported: exported function Error should have comment or be unexported (revive)
  func Error(msg string) {
  ^
  Error: internal/ui/ui.go:95:1: exported: exported function Warning should have comment or be unexported (revive)
  func Warning(msg string) {
  ^
  Error: internal/ui/ui.go:100:1: exported: exported function Info should have comment or be unexported (revive)
  func Info(msg string) {
  ^
  Error: internal/ui/ui.go:105:1: exported: exported function Dim should have comment or be unexported (revive)
  func Dim(msg string) {
  ^
  Error: internal/ui/ui.go:110:1: exported: exported function Bold should have comment or be unexported (revive)
  func Bold(msg string) {
  ^
  Error: internal/ui/ui.go:115:1: exported: exported function Spacer should have comment or be unexported (revive)
  func Spacer() {
  ^
  Error: internal/ui/ui.go:120:1: exported: exported function Divider should have comment or be unexported (revive)
  func Divider() {
  ^
  Error: internal/ui/ui.go:154:1: exported: exported function Code should have comment or be unexported (revive)
  func Code(msg string) {
  ^
  Error: internal/ui/ui.go:158:1: exported: exported function Section should have comment or be unexported (revive)
  func Section(title string) {
  ^
  Error: internal/ui/ui.go:164:1: exported: exported function KeyValue should have comment or be unexported (revive)
  func KeyValue(key, value string) {
  ^
  Error: internal/ui/ui.go:168:1: exported: exported function List should have comment or be unexported (revive)
  func List(items []string) {
  ^
  Error: internal/ui/ui.go:233:1: exported: exported function Input should have comment or be unexported (revive)
  func Input(prompt, placeholder string) (string, error) {
  ^
  Error: internal/ui/ui.go:243:1: exported: exported function InputWithDefault should have comment or be unexported (revive)
  func InputWithDefault(prompt, defaultValue string) (string, error) {
  ^
  Error: internal/ui/ui.go:259:1: exported: exported function Password should have comment or be unexported (revive)
  func Password(prompt string) (string, error) {
  ^
  Error: internal/ui/ui.go:269:1: exported: exported function Confirm should have comment or be unexported (revive)
  func Confirm(prompt string) (bool, error) {
  ^
  Error: internal/ui/ui.go:280:1: exported: exported function Select should have comment or be unexported (revive)
  func Select(prompt string, options []string) (string, error) {
  ^
  Error: internal/ui/ui.go:299:1: exported: exported function SelectWithKeys should have comment or be unexported (revive)
  func SelectWithKeys(prompt string, options map[string]string) (string, error) {
  ^
  Error: internal/ui/ui.go:318:1: exported: exported function MultiSelect should have comment or be unexported (revive)
  func MultiSelect(prompt string, options []string) ([]string, error) {
  ^
  Error: internal/ui/ui.go:337:1: exported: exported function Form should have comment or be unexported (revive)
  func Form(groups ...*huh.Group) error {
  ^
  Error: internal/ui/ui.go:341:1: exported: exported function ConfirmAction should have comment or be unexported (revive)
  func ConfirmAction(action, resource string) (bool, error) {
  ^
  Error: internal/ui/ui.go:352:1: exported: exported function NewLogStream should have comment or be unexported (revive)
  func NewLogStream() *LogStream {
  ^
  Error: internal/ui/ui.go:360:1: exported: exported method LogStream.WriteRaw should have comment or be unexported (revive)
  func (l *LogStream) WriteRaw(msg string) {
  ^
  Error: internal/ui/ui.go:367:1: exported: exported function NewCmdOutput should have comment or be unexported (revive)
  func NewCmdOutput() *CmdOutput {
  ^
  Error: internal/ui/ui.go:388:1: exported: exported function NewStatus should have comment or be unexported (revive)
  func NewStatus(message string) *Status {
  ^
  Error: internal/ui/ui.go:392:1: exported: exported method Status.Update should have comment or be unexported (revive)
  func (s *Status) Update(message string) {
  ^
  Error: internal/ui/ui.go:397:1: exported: exported method Status.Done should have comment or be unexported (revive)
  func (s *Status) Done() {
  ^
  Error: internal/ui/ui.go:403:1: exported: exported function NextSteps should have comment or be unexported (revive)
  func NextSteps(steps []string) {
  ^
  Error: internal/ui/ui.go:411:1: exported: exported function ErrorWithSuggestion should have comment or be unexported (revive)
  func ErrorWithSuggestion(err error, suggestion string) {
  ^
  Error: internal/ui/ui.go:419:1: exported: exported function StepProgress should have comment or be unexported (revive)
  func StepProgress(current, total int, stepName string) {
  ^
  Error: internal/utils/healthcheck.go:1:9: var-naming: avoid meaningless package names (revive)
  package utils
          ^
  Error: cmd/services.go:100:3: QF1003: could use tagged switch on db.Status (staticcheck)
  		if db.Status == "running" {
  		^
  Error: internal/config/config.go:373:11: ST1005: error strings should not be capitalized (staticcheck)
  			return fmt.Errorf("Azure location is required")
  			       ^
  Error: internal/config/config.go:376:11: ST1005: error strings should not be capitalized (staticcheck)
  			return fmt.Errorf("Azure resource group is required")
  			       ^
  Error: internal/config/config.go:380:11: ST1005: error strings should not be capitalized (staticcheck)
  			return fmt.Errorf("Local work directory is required")
  			       ^
  Error: internal/config/config.go:384:11: ST1005: error strings should not be capitalized (staticcheck)
  			return fmt.Errorf("Production domain is required")
  			       ^
  Error: internal/config/config.go:396:11: ST1005: error strings should not be capitalized (staticcheck)
  			return fmt.Errorf("Bare Metal host is required")
  			       ^
  Error: internal/git/repo.go:258:13: S1039: unnecessary use of fmt.Sprintf (staticcheck)
  	message := fmt.Sprintf("Deploy via cdp")
  	           ^
  Error: internal/installer/github.go:80:18: S1039: unnecessary use of fmt.Sprintf (staticcheck)
  	dashboardURL := fmt.Sprintf("http://localhost:8000")
  	                ^
  Error: internal/local/docker.go:24:10: ST1005: error strings should not be capitalized (staticcheck)
  		return fmt.Errorf("Docker is not running. Please start Docker")
  		       ^
  Error: internal/providers/azure/azure.go:174:11: ST1005: error strings should not be capitalized (staticcheck)
  			return fmt.Errorf("Azure SDK validation failed: %w", err)
  			       ^
  Error: internal/providers/azure/azure.go:180:11: ST1005: error strings should not be capitalized (staticcheck)
  			return fmt.Errorf("Azure CLI not authenticated. Run 'az login': %w", err)
  			       ^
  Error: internal/providers/azure/backup.go:319:10: ST1005: error strings should not be capitalized (staticcheck)
  		return fmt.Errorf("Redis restore failed: %w", err)
  		       ^
  Error: internal/providers/azure/deploy.go:54:10: ST1005: error strings should not be capitalized (staticcheck)
  		return fmt.Errorf("Coolify installation failed: %w", err)
  		       ^
  Error: internal/providers/azure/sdk_client.go:500:2: S1000: should use for range instead of for { select {} } (staticcheck)
  	for {
  	^
  Error: internal/providers/hetzner/client.go:238:15: SA1019: server.Datacenter is deprecated: [Server.Datacenter] is deprecated and will be removed after 1 July 2026. Use [Server.Location] instead. See https://docs.hetzner.cloud/changelog#2025-12-16-phasing-out-datacenters (staticcheck)
  		Location:   server.Datacenter.Location.Name,
  		            ^
  Error: internal/providers/production/production.go:257:14: S1039: unnecessary use of fmt.Sprintf (staticcheck)
  	sslCheck := fmt.Sprintf(`
  	            ^
  Error: internal/providers/azure/sdk_client.go:434:19: unnecessary conversion (unconvert)
  		State:    string(*resp.Properties.ProvisioningState),
  		                ^
  Error: internal/providers/azure/sdk_client.go:462:22: unnecessary conversion (unconvert)
  		info.State = string(*resp.Properties.ProvisioningState)
  		                   ^
  Error: cmd/helpers.go:47:6: func getVerboseFlag is unused (unused)
  func getVerboseFlag(cmd *cobra.Command) bool {
       ^
  Error: cmd/helpers.go:53:6: func requireLogin is unused (unused)
  func requireLogin(cmd *cobra.Command, args []string) error {
       ^
  Error: internal/api/client.go:222:18: func (*Client).request is unused (unused)
  func (c *Client) request(method, path string, body interface{}, result interface{}) error {
                   ^
  Error: internal/detect/detector.go:444:6: func detectPython is unused (unused)
  func detectPython(dir string) (*FrameworkInfo, error) {
       ^
  Error: internal/ui/config_menu.go:24:2: field width is unused (unused)
  	width   int
  	^
  Error: internal/ui/config_menu.go:25:2: field height is unused (unused)
  	height  int
  	^
  Error: internal/ui/messages.go:6:6: type taskStartMsg is unused (unused)
  type taskStartMsg struct{}
       ^
  Error: internal/ui/model.go:66:2: field config is unused (unused)
  	config      *config.Config
  	^
  Error: internal/ui/model.go:69:2: field selected is unused (unused)
  	selected    int
  	^
  Error: internal/ui/model.go:70:2: field loading is unused (unused)
  	loading     bool
  	^
  215 issues:
  * errcheck: 50
  * gocognit: 6
  * gocritic: 6
  * gocyclo: 3
  * gosec: 43
  * ineffassign: 2
  * misspell: 20
  * prealloc: 7
  * revive: 50
  * staticcheck: 16
  * unconvert: 2
  * unused: 10
  
  level=warning msg="[linters_context] gocritic: no need to disable check \"hugeParam\": it's already disabled"
  
  Error: issues found
  Ran golangci-lint in 94839ms