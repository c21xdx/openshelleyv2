package main

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"embed"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"strconv"
	"sync"
	"time"

	"github.com/creack/pty"
	"github.com/gorilla/websocket"
)

//go:embed static/*
var staticFS embed.FS

var (
	authToken   string
	shelleyURL  = "http://localhost:9001" // 开源Shelley内部端口
	portalPort  = "8000"
	baseDir     string
	mgmtMutex   sync.Mutex
)

type FileInfo struct {
	Name    string `json:"name"`
	Path    string `json:"path"`
	IsDir   bool   `json:"isDir"`
	Size    int64  `json:"size"`
	ModTime string `json:"modTime"`
	Mode    string `json:"mode"`
}

func main() {
	// Generate or use provided token
	authToken = os.Getenv("PORTAL_TOKEN")
	if authToken == "" {
		authToken = generateToken()
	}

	if envPort := os.Getenv("PORTAL_PORT"); envPort != "" {
		portalPort = envPort
	}

	if envShelley := os.Getenv("SHELLEY_URL"); envShelley != "" {
		shelleyURL = envShelley
	}

	// Get base directory from env or use executable's parent directory
	baseDir = os.Getenv("BASE_DIR")
	if baseDir == "" {
		exePath, _ := os.Executable()
		baseDir = filepath.Dir(exePath)
	}

	log.Printf("Portal starting on port %s", portalPort)
	log.Printf("Auth Token: %s", authToken)
	log.Printf("Open Shelley URL: %s", shelleyURL)

	mux := http.NewServeMux()

	// Login page (no auth required)
	mux.HandleFunc("/login", handleLogin)
	mux.HandleFunc("/logout", handleLogout)

	// Portal static files
	mux.HandleFunc("/portal/", handlePortalStatic)

	// Portal pages (require auth)
	mux.HandleFunc("/portal", authMiddleware(handlePortalHome))
	mux.HandleFunc("/portal/terminal", authMiddleware(handleTerminalPage))
	mux.HandleFunc("/portal/files", authMiddleware(handleFilesPage))

	// Portal API endpoints (require auth)
	mux.HandleFunc("/portal/api/files/", authMiddleware(handleFilesAPI))
	mux.HandleFunc("/portal/api/file/", authMiddleware(handleFileAPI))

	// Management API endpoints
	mux.HandleFunc("/portal/api/mgmt/status", authMiddleware(handleMgmtStatus))
	mux.HandleFunc("/portal/api/mgmt/token", authMiddleware(handleMgmtToken))
	mux.HandleFunc("/portal/api/mgmt/start", authMiddleware(handleMgmtStart))
	mux.HandleFunc("/portal/api/mgmt/stop", authMiddleware(handleMgmtStop))
	mux.HandleFunc("/portal/api/mgmt/restart", authMiddleware(handleMgmtRestart))
	mux.HandleFunc("/portal/api/mgmt/check-update", authMiddleware(handleMgmtCheckUpdate))
	mux.HandleFunc("/portal/api/mgmt/update", authMiddleware(handleMgmtUpdate))
	mux.HandleFunc("/portal/api/mgmt/backups", authMiddleware(handleMgmtBackups))
	mux.HandleFunc("/portal/api/mgmt/rollback", authMiddleware(handleMgmtRollback))

	// WebSocket for terminal
	mux.HandleFunc("/portal/ws/terminal", authMiddleware(handleTerminalWS))

	// Everything else goes to Shelley (require auth)
	mux.HandleFunc("/", authMiddleware(handleShelleyProxy))

	log.Fatal(http.ListenAndServe(":"+portalPort, mux))
}

func generateToken() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func handlePortalStatic(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/portal/")
	if path == "" {
		path = "index.html"
	}
	data, err := staticFS.ReadFile("static/" + path)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	switch {
	case strings.HasSuffix(path, ".css"):
		w.Header().Set("Content-Type", "text/css")
	case strings.HasSuffix(path, ".js"):
		w.Header().Set("Content-Type", "application/javascript")
	case strings.HasSuffix(path, ".html"):
		w.Header().Set("Content-Type", "text/html")
	}

	w.Write(data)
}

func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("portal_token")
		if err != nil || cookie.Value != authToken {
			headerToken := r.Header.Get("Authorization")
			if headerToken != "Bearer "+authToken {
				if strings.HasPrefix(r.URL.Path, "/portal/api/") || strings.HasPrefix(r.URL.Path, "/portal/ws/") {
					http.Error(w, "Unauthorized", http.StatusUnauthorized)
				} else {
					http.Redirect(w, r, "/login", http.StatusSeeOther)
				}
				return
			}
		}
		next(w, r)
	}
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		r.ParseForm()
		token := r.FormValue("token")
		if token == authToken {
			http.SetCookie(w, &http.Cookie{
				Name:     "portal_token",
				Value:    token,
				Path:     "/",
				HttpOnly: true,
				MaxAge:   86400 * 30,
			})
			http.Redirect(w, r, "/portal", http.StatusSeeOther)
			return
		}
		http.Redirect(w, r, "/login?error=invalid", http.StatusSeeOther)
		return
	}

	data, _ := staticFS.ReadFile("static/login.html")
	w.Header().Set("Content-Type", "text/html")
	w.Write(data)
}

func handleLogout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:   "portal_token",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func handlePortalHome(w http.ResponseWriter, r *http.Request) {
	data, _ := staticFS.ReadFile("static/index.html")
	w.Header().Set("Content-Type", "text/html")
	w.Write(data)
}

func handleTerminalPage(w http.ResponseWriter, r *http.Request) {
	data, _ := staticFS.ReadFile("static/terminal.html")
	w.Header().Set("Content-Type", "text/html")
	w.Write(data)
}

func handleFilesPage(w http.ResponseWriter, r *http.Request) {
	data, _ := staticFS.ReadFile("static/files.html")
	w.Header().Set("Content-Type", "text/html")
	w.Write(data)
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func handleTerminalWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	defer conn.Close()

	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/bash"
	}

	cmd := exec.Command(shell)
	cmd.Env = append(os.Environ(), "TERM=xterm-256color")

	ptmx, err := pty.Start(cmd)
	if err != nil {
		log.Printf("PTY start error: %v", err)
		return
	}
	defer ptmx.Close()

	pty.Setsize(ptmx, &pty.Winsize{Rows: 24, Cols: 80})

	go func() {
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				return
			}
			if len(msg) > 0 && msg[0] == '{' {
				var resize struct {
					Type string `json:"type"`
					Cols int    `json:"cols"`
					Rows int    `json:"rows"`
				}
				if json.Unmarshal(msg, &resize) == nil && resize.Type == "resize" {
					pty.Setsize(ptmx, &pty.Winsize{
						Rows: uint16(resize.Rows),
						Cols: uint16(resize.Cols),
					})
					continue
				}
			}
			ptmx.Write(msg)
		}
	}()

	buf := make([]byte, 4096)
	for {
		n, err := ptmx.Read(buf)
		if err != nil {
			return
		}
		if err := conn.WriteMessage(websocket.BinaryMessage, buf[:n]); err != nil {
			return
		}
	}
}

func handleFilesAPI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	path := strings.TrimPrefix(r.URL.Path, "/portal/api/files")
	if path == "" || path == "/" {
		path = os.Getenv("HOME")
	}

	switch r.Method {
	case "GET":
		entries, err := os.ReadDir(path)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		files := make([]FileInfo, 0, len(entries))
		for _, entry := range entries {
			info, err := entry.Info()
			if err != nil {
				continue
			}
			files = append(files, FileInfo{
				Name:    entry.Name(),
				Path:    filepath.Join(path, entry.Name()),
				IsDir:   entry.IsDir(),
				Size:    info.Size(),
				ModTime: info.ModTime().Format(time.RFC3339),
				Mode:    info.Mode().String(),
			})
		}
		sort.Slice(files, func(i, j int) bool {
			if files[i].IsDir != files[j].IsDir {
				return files[i].IsDir
			}
			return files[i].Name < files[j].Name
		})
		json.NewEncoder(w).Encode(map[string]interface{}{"path": path, "files": files})

	case "POST":
		var req struct {
			Type string `json:"type"`
			Name string `json:"name"`
		}
		json.NewDecoder(r.Body).Decode(&req)
		newPath := filepath.Join(path, req.Name)
		if req.Type == "dir" {
			err := os.MkdirAll(newPath, 0755)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		} else {
			f, err := os.Create(newPath)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			f.Close()
		}
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})

	case "DELETE":
		err := os.RemoveAll(path)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}
}

func handleFileAPI(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/portal/api/file")

	switch r.Method {
	case "GET":
		info, err := os.Stat(path)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		if info.Size() > 10*1024*1024 {
			http.Error(w, "File too large", http.StatusBadRequest)
			return
		}
		content, err := os.ReadFile(path)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"path": path, "content": string(content), "size": info.Size(), "modTime": info.ModTime().Format(time.RFC3339),
		})

	case "PUT":
		var req struct {
			Content string `json:"content"`
		}
		json.NewDecoder(r.Body).Decode(&req)
		err := os.WriteFile(path, []byte(req.Content), 0644)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})

	case "POST":
		var req struct {
			NewPath string `json:"newPath"`
		}
		json.NewDecoder(r.Body).Decode(&req)
		err := os.Rename(path, req.NewPath)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}
}

// ============== Management API Handlers ==============

// Check if Shelley process is running
func isShelleyRunning() bool {
	cmd := exec.Command("pgrep", "-f", "shelley.*serve")
	return cmd.Run() == nil
}

// Get current Shelley version
func getCurrentVersion() string {
	binaryPath := filepath.Join(baseDir, "shelley")
	cmd := exec.Command(binaryPath, "version")
	output, err := cmd.Output()
	if err != nil {
		return "unknown"
	}
	var ver struct {
		Tag string `json:"tag"`
	}
	if json.Unmarshal(output, &ver) == nil {
		return ver.Tag
	}
	return "unknown"
}

// Get latest version from GitHub
func getLatestVersion() (string, error) {
	resp, err := http.Get("https://api.github.com/repos/boldsoftware/shelley/releases/latest")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	
	var release struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", err
	}
	return release.TagName, nil
}

func handleMgmtToken(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"token": authToken,
	})
}

func handleMgmtStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	running := isShelleyRunning()
	currentVer := getCurrentVersion()
	latestVer, _ := getLatestVersion()
	
	hasUpdate := false
	if latestVer != "" && currentVer != "unknown" && currentVer != latestVer {
		hasUpdate = true
	}
	
	json.NewEncoder(w).Encode(map[string]interface{}{
		"shelley_running":  running,
		"current_version":  currentVer,
		"latest_version":   latestVer,
		"has_update":       hasUpdate,
	})
}

func handleMgmtStart(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	mgmtMutex.Lock()
	defer mgmtMutex.Unlock()
	
	if isShelleyRunning() {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"output":  "Shelley is already running",
		})
		return
	}
	
	// Start Shelley using the start script
	scriptPath := filepath.Join(baseDir, "start.sh")
	cmd := exec.Command("bash", scriptPath)
	cmd.Dir = baseDir
	
	if err := cmd.Start(); err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}
	
	// Wait a bit for process to start
	time.Sleep(3 * time.Second)
	
	if isShelleyRunning() {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"output":  "Shelley started successfully",
		})
	} else {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Process failed to start. Check logs.",
		})
	}
}

func handleMgmtStop(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	mgmtMutex.Lock()
	defer mgmtMutex.Unlock()
	
	cmd := exec.Command("pkill", "-f", "shelley.*serve")
	cmd.Run() // Ignore error if not running
	
	time.Sleep(1 * time.Second)
	
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
	})
}

func handleMgmtRestart(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	mgmtMutex.Lock()
	defer mgmtMutex.Unlock()
	
	// Stop
	stopScript := filepath.Join(baseDir, "stop.sh")
	exec.Command("bash", stopScript).Run()
	time.Sleep(2 * time.Second)
	
	// Start
	startScript := filepath.Join(baseDir, "start.sh")
	cmd := exec.Command("bash", startScript)
	cmd.Dir = baseDir
	
	if err := cmd.Start(); err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}
	
	time.Sleep(3 * time.Second)
	
	if isShelleyRunning() {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
		})
	} else {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Process failed to restart",
		})
	}
}

func handleMgmtCheckUpdate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	currentVer := getCurrentVersion()
	latestVer, err := getLatestVersion()
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Failed to check latest version: " + err.Error(),
		})
		return
	}
	
	hasUpdate := currentVer != "unknown" && currentVer != latestVer
	
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":         true,
		"current_version": currentVer,
		"latest_version":  latestVer,
		"has_update":      hasUpdate,
	})
}

func handleMgmtUpdate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	mgmtMutex.Lock()
	defer mgmtMutex.Unlock()
	
	// Run update script
	scriptPath := filepath.Join(baseDir, "update-shelley.sh")
	cmd := exec.Command("bash", scriptPath)
	cmd.Dir = baseDir
	
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	
	err := cmd.Run()
	output := stdout.String() + stderr.String()
	
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   err.Error(),
			"output":  output,
		})
		return
	}
	
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"output":  output,
	})
}

// List available backups
func handleMgmtBackups(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	binaryPath := filepath.Join(baseDir, "shelley")
	pattern := binaryPath + ".backup.*"
	
	matches, err := filepath.Glob(pattern)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}
	
	type BackupInfo struct {
		Name    string `json:"name"`
		Path    string `json:"path"`
		Size    int64  `json:"size"`
		ModTime string `json:"modTime"`
	}
	
	backups := make([]BackupInfo, 0)
	for _, match := range matches {
		info, err := os.Stat(match)
		if err != nil {
			continue
		}
		backups = append(backups, BackupInfo{
			Name:    filepath.Base(match),
			Path:    match,
			Size:    info.Size(),
			ModTime: info.ModTime().Format("2006-01-02 15:04:05"),
		})
	}
	
	// Sort by modification time (newest first)
	sort.Slice(backups, func(i, j int) bool {
		return backups[i].ModTime > backups[j].ModTime
	})
	
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"backups": backups,
	})
}

// Rollback to a specific backup
func handleMgmtRollback(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	mgmtMutex.Lock()
	defer mgmtMutex.Unlock()
	
	var req struct {
		BackupName string `json:"backup_name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Invalid request",
		})
		return
	}
	
	binaryPath := filepath.Join(baseDir, "shelley")
	backupPath := filepath.Join(baseDir, req.BackupName)
	
	// Verify backup exists and is a valid backup file
	if !strings.HasPrefix(req.BackupName, "shelley.backup.") {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Invalid backup name",
		})
		return
	}
	
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Backup not found",
		})
		return
	}
	
	// Stop Shelley
	exec.Command("pkill", "-f", "shelley.*serve").Run()
	time.Sleep(2 * time.Second)
	
	// Backup current binary before rollback
	currentBackup := binaryPath + ".before-rollback." + time.Now().Format("20060102_150405")
	if _, err := os.Stat(binaryPath); err == nil {
		if err := copyFile(binaryPath, currentBackup); err != nil {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   "Failed to backup current binary: " + err.Error(),
			})
			return
		}
	}
	
	// Copy backup to binary
	if err := copyFile(backupPath, binaryPath); err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Failed to restore backup: " + err.Error(),
		})
		return
	}
	
	// Make executable
	os.Chmod(binaryPath, 0755)
	
	// Restart Shelley
	scriptPath := filepath.Join(baseDir, "start.sh")
	if _, err := os.Stat(scriptPath); err == nil {
		cmd := exec.Command("bash", scriptPath)
		cmd.Dir = baseDir
		cmd.Start()
	}
	
	time.Sleep(2 * time.Second)
	
	// Get version of restored binary
	var version string
	cmd := exec.Command(binaryPath, "version")
	if output, err := cmd.Output(); err == nil {
		var ver struct {
			Tag string `json:"tag"`
		}
		if json.Unmarshal(output, &ver) == nil {
			version = ver.Tag
		}
	}
	
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Rolled back to " + req.BackupName,
		"version": version,
	})
}

// Helper function to copy file
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()
	
	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()
	
	_, err = io.Copy(destFile, sourceFile)
	return err
}

// Portal button HTML/CSS to inject into Shelley pages
const portalButtonHTML = `
<style>
#portal-nav-btn {
    position: fixed;
    top: 12px;
    right: 12px;
    z-index: 9999;
    padding: 8px 16px;
    background: #2563eb;
    color: white;
    border: none;
    border-radius: 6px;
    font-family: "SF Mono", Monaco, monospace;
    font-size: 14px;
    cursor: pointer;
    text-decoration: none;
    display: flex;
    align-items: center;
    gap: 6px;
    box-shadow: 0 2px 8px rgba(0,0,0,0.15);
    transition: background 0.2s;
}
#portal-nav-btn:hover {
    background: #1d4ed8;
}
</style>
<a id="portal-nav-btn" href="/portal">
    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <path d="M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"></path>
        <polyline points="9 22 9 12 15 12 15 22"></polyline>
    </svg>
    Portal
</a>
`

func handleShelleyProxy(w http.ResponseWriter, r *http.Request) {
	target, _ := url.Parse(shelleyURL)

	// Check if this is an SSE request
	if strings.Contains(r.URL.Path, "/stream") || r.Header.Get("Accept") == "text/event-stream" {
		handleSSEProxy(w, r, target)
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(target)
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		req.Host = target.Host
	}
	
	// Inject portal button into HTML responses
	proxy.ModifyResponse = func(resp *http.Response) error {
		contentType := resp.Header.Get("Content-Type")
		if strings.Contains(contentType, "text/html") {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			
			// Inject portal button before </body>
			modified := strings.Replace(string(body), "</body>", portalButtonHTML+"</body>", 1)
			
			resp.Body = io.NopCloser(strings.NewReader(modified))
			resp.ContentLength = int64(len(modified))
			resp.Header.Set("Content-Length", strconv.Itoa(len(modified)))
		}
		return nil
	}
	
	proxy.ServeHTTP(w, r)
}

func handleSSEProxy(w http.ResponseWriter, r *http.Request, target *url.URL) {
	proxyURL := target.String() + r.URL.Path
	if r.URL.RawQuery != "" {
		proxyURL += "?" + r.URL.RawQuery
	}

	req, err := http.NewRequestWithContext(r.Context(), r.Method, proxyURL, r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for k, v := range r.Header {
		req.Header[k] = v
	}

	client := &http.Client{Timeout: 0}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	reader := bufio.NewReader(resp.Body)
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err != io.EOF {
				log.Printf("SSE read error: %v", err)
			}
			return
		}
		w.Write(line)
		flusher.Flush()
	}
}
