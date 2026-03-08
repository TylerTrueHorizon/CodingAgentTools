package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"agent-tools-sandbox/internal/models"
	"agent-tools-sandbox/internal/pathutil"
)

// Files holds dependencies for file handlers.
type Files struct{}

// Read handles GET /files/read?path=...&start_line=&end_line=
func (f *Files) Read(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	path := r.URL.Query().Get("path")
	if path == "" {
		writeJSONError(w, "path is required", http.StatusBadRequest)
		return
	}
	absPath, err := pathutil.ResolveAbsolute(path)
	if err != nil {
		writeJSONError(w, "invalid path: "+err.Error(), http.StatusBadRequest)
		return
	}
	info, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			writeJSONError(w, "not found", http.StatusNotFound)
			return
		}
		writeJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if info.IsDir() {
		writeJSONError(w, "path is a directory; use list", http.StatusBadRequest)
		return
	}
	content, err := os.ReadFile(absPath)
	if err != nil {
		writeJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	text := string(content)
	startLine, _ := strconv.Atoi(r.URL.Query().Get("start_line"))
	endLine, _ := strconv.Atoi(r.URL.Query().Get("end_line"))
	if startLine > 0 || endLine > 0 {
		text = sliceLines(text, startLine, endLine)
	}
	writeJSON(w, http.StatusOK, models.ReadFileResponse{Content: text})
}

// List handles GET /files/list?path=...&pattern=
func (f *Files) List(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	path := r.URL.Query().Get("path")
	if path == "" {
		writeJSONError(w, "path is required", http.StatusBadRequest)
		return
	}
	absPath, err := pathutil.ResolveAbsolute(path)
	if err != nil {
		writeJSONError(w, "invalid path: "+err.Error(), http.StatusBadRequest)
		return
	}
	info, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			writeJSONError(w, "not found", http.StatusNotFound)
			return
		}
		writeJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !info.IsDir() {
		writeJSONError(w, "path is not a directory", http.StatusBadRequest)
		return
	}
	pattern := r.URL.Query().Get("pattern")
	var entries []models.DirEntry
	if pattern == "" {
		entries, err = listDir(absPath)
	} else {
		entries, err = globDir(absPath, pattern)
	}
	if err != nil {
		writeJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, models.ListDirResponse{Entries: entries})
}

// Write handles POST /files/write.
func (f *Files) Write(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req models.WriteFileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}
	if req.Path == "" {
		writeJSONError(w, "path is required", http.StatusBadRequest)
		return
	}
	absPath, err := pathutil.ResolveAbsolute(req.Path)
	if err != nil {
		writeJSONError(w, "invalid path: "+err.Error(), http.StatusBadRequest)
		return
	}
	if err := os.MkdirAll(filepath.Dir(absPath), 0755); err != nil {
		writeJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := os.WriteFile(absPath, []byte(req.Content), 0644); err != nil {
		writeJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, models.WriteFileResponse{Path: absPath})
}

// Edit handles POST /files/edit.
func (f *Files) Edit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req models.EditFileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}
	if req.Path == "" {
		writeJSONError(w, "path is required", http.StatusBadRequest)
		return
	}
	absPath, err := pathutil.ResolveAbsolute(req.Path)
	if err != nil {
		writeJSONError(w, "invalid path: "+err.Error(), http.StatusBadRequest)
		return
	}
	content, err := os.ReadFile(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			writeJSONError(w, "not found", http.StatusNotFound)
			return
		}
		writeJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	text := string(content)
	switch req.EditType {
	case "str_replace":
		if !strings.Contains(text, req.OldStr) {
			writeJSONError(w, "old_str not found in file", http.StatusBadRequest)
			return
		}
		text = strings.Replace(text, req.OldStr, req.NewStr, 1)
	case "insert":
		if req.Line < 1 {
			writeJSONError(w, "line must be >= 1", http.StatusBadRequest)
			return
		}
		text, err = insertAtLine(text, req.Line, req.Content)
		if err != nil {
			writeJSONError(w, err.Error(), http.StatusBadRequest)
			return
		}
	default:
		writeJSONError(w, "edit_type must be str_replace or insert", http.StatusBadRequest)
		return
	}
	if err := os.WriteFile(absPath, []byte(text), 0644); err != nil {
		writeJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	snippet := text
	if len(snippet) > 200 {
		snippet = snippet[:200] + "..."
	}
	writeJSON(w, http.StatusOK, models.EditFileResponse{Path: absPath, Snippet: snippet})
}

func listDir(dir string) ([]models.DirEntry, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	out := make([]models.DirEntry, 0, len(entries))
	for _, e := range entries {
		t := "file"
		if e.IsDir() {
			t = "dir"
		}
		out = append(out, models.DirEntry{Name: e.Name(), Type: t})
	}
	return out, nil
}

func globDir(dir, pattern string) ([]models.DirEntry, error) {
	fullPattern := filepath.Join(dir, pattern)
	matches, err := filepath.Glob(fullPattern)
	if err != nil {
		return nil, err
	}
	out := make([]models.DirEntry, 0, len(matches))
	for _, m := range matches {
		info, err := os.Stat(m)
		if err != nil {
			continue
		}
		t := "file"
		if info.IsDir() {
			t = "dir"
		}
		out = append(out, models.DirEntry{Name: filepath.Base(m), Type: t})
	}
	return out, nil
}

func sliceLines(text string, start, end int) string {
	// 1-based; -1 or 0 for end means to end of file
	lines := strings.Split(text, "\n")
	if start < 1 {
		start = 1
	}
	if end < 1 {
		end = len(lines)
	}
	if start > len(lines) {
		return ""
	}
	if end > len(lines) {
		end = len(lines)
	}
	return strings.Join(lines[start-1:end], "\n")
}

func insertAtLine(text string, line int, content string) (string, error) {
	lines := strings.Split(text, "\n")
	if line < 1 || line > len(lines)+1 {
		return "", errors.New("line out of range")
	}
	idx := line - 1
	newLines := make([]string, 0, len(lines)+1)
	newLines = append(newLines, lines[:idx]...)
	newLines = append(newLines, content)
	newLines = append(newLines, lines[idx:]...)
	return strings.Join(newLines, "\n"), nil
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeJSONError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}
