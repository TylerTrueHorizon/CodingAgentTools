package models

// ----- Files -----

// ReadFileResponse is the response for GET /files/read.
type ReadFileResponse struct {
	Content string `json:"content"`
}

// ListDirResponse is the response for GET /files/list (and glob).
type ListDirResponse struct {
	Entries []DirEntry `json:"entries"`
}

// DirEntry is a single file or directory entry.
type DirEntry struct {
	Name string `json:"name"`
	Type string `json:"type"` // "file" or "dir"
}

// WriteFileRequest is the body for POST /files/write.
type WriteFileRequest struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

// WriteFileResponse is the response for POST /files/write.
type WriteFileResponse struct {
	Path string `json:"path"`
}

// EditFileRequest is the body for POST /files/edit.
type EditFileRequest struct {
	Path     string `json:"path"`
	EditType string `json:"edit_type"` // "str_replace" or "insert"
	OldStr   string `json:"old_str,omitempty"`
	NewStr   string `json:"new_str,omitempty"`
	Line     int    `json:"line,omitempty"`     // 1-based for insert
	Content  string `json:"content,omitempty"`  // for insert
}

// EditFileResponse is the response for POST /files/edit.
type EditFileResponse struct {
	Path    string `json:"path"`
	Snippet string `json:"snippet,omitempty"`
}

// ----- Shell -----

// ShellRunRequest is the body for POST /shell/run.
type ShellRunRequest struct {
	Command        string `json:"command"`
	Cwd            string `json:"cwd,omitempty"`
	TimeoutSeconds int    `json:"timeout_seconds,omitempty"`
}

// ShellRunResponse is the response for POST /shell/run.
type ShellRunResponse struct {
	Stdout   string `json:"stdout"`
	Stderr   string `json:"stderr"`
	ExitCode int    `json:"exit_code"`
}
