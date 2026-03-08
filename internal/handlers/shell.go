package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"agent-tools-sandbox/internal/config"
	"agent-tools-sandbox/internal/models"
	"agent-tools-sandbox/internal/pathutil"
)

// Shell holds config for shell run (timeout default).
type Shell struct {
	DefaultTimeoutSec int
}

// NewShell returns a Shell with config applied.
func NewShell(cfg config.Config) Shell {
	return Shell{DefaultTimeoutSec: cfg.ShellTimeoutSec}
}

// Run handles POST /shell/run. Runs command as-is (e.g. sh -c "..." so sudo works).
func (s *Shell) Run(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req models.ShellRunRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}
	req.Command = strings.TrimSpace(req.Command)
	if req.Command == "" {
		writeJSONError(w, "command is required", http.StatusBadRequest)
		return
	}
	timeoutSec := req.TimeoutSeconds
	if timeoutSec <= 0 {
		timeoutSec = s.DefaultTimeoutSec
	}
	ctx, cancel := context.WithTimeout(r.Context(), time.Duration(timeoutSec)*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "sh", "-c", req.Command)
	if req.Cwd != "" {
		absCwd, err := pathutil.ResolveAbsolute(req.Cwd)
		if err != nil {
			writeJSONError(w, "invalid cwd: "+err.Error(), http.StatusBadRequest)
			return
		}
		cmd.Dir = absCwd
	}
	stdout, err := cmd.Output()
	stderr := []byte(nil)
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			stderr = exitErr.Stderr
			exitCode = exitErr.ExitCode()
		} else {
			if ctx.Err() == context.DeadlineExceeded {
				writeJSONError(w, "command timed out", http.StatusGatewayTimeout)
				return
			}
			writeJSONError(w, "exec failed: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}
	writeJSON(w, http.StatusOK, models.ShellRunResponse{
		Stdout:   string(stdout),
		Stderr:   string(stderr),
		ExitCode: exitCode,
	})
}
