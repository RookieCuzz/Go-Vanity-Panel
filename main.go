// Package main provides the go-vanity CLI tool, a basic server implementation
// capable of providing custom URLs to be used by the standard go tools.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// 全局配置变量，用于动态更新
var globalConfig *Configuration
var globalHandler *handler

func main() {
	// Validate arguments
	file, port := getParameters()
	if file == "" {
		fmt.Println("error: a configuration file is required")
		os.Exit(-1)
	}

	// Read configuration file
	contents, err := os.ReadFile(filepath.Clean(file))
	if err != nil {
		fmt.Println("failed to read configuration file: ", err)
		os.Exit(-1)
	}

	// Decode configuration file
	globalConfig = NewServerConfig()
	if strings.HasSuffix(file, ".yaml") || strings.HasSuffix(file, ".yml") {
		if err := yaml.Unmarshal(contents, globalConfig); err != nil {
			fmt.Println("failed to decode YAML configuration file: ", err)
			os.Exit(-1)
		}
	}
	if strings.HasSuffix(file, ".json") {
		if err := json.Unmarshal(contents, globalConfig); err != nil {
			fmt.Println("failed to decode JSON configuration file: ", err)
			os.Exit(-1)
		}
	}
	if len(globalConfig.Paths) == 0 {
		fmt.Println("no valid configuration to use")
		os.Exit(-1)
	}

	// Start server
	globalHandler = newHandler(globalConfig)
	fmt.Println("serving on port:", port)
	srv := http.Server{
		Addr:              fmt.Sprintf(":%d", port),
		Handler:           logMiddleware(getServerMux()),
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      5 * time.Second,
		ReadTimeout:       10 * time.Second,
	}
	if err := srv.ListenAndServe(); err != nil {
		fmt.Println("server error: ", err)
		os.Exit(-1)
	}
}

func getParameters() (string, int) {
	// Define flags
	file := ""
	port := 9090
	ff := flag.String("config", file, "configuration file")
	fp := flag.Int("port", port, "TCP port")
	flag.Parse()

	// Read file from ENV variable and flag
	if ef := os.Getenv("GOVANITY_CONFIG"); ef != "" {
		file = ef
	}
	if *ff != "" {
		file = *ff
	}

	// Read port from ENV variable and flag
	if ep := os.Getenv("GOVANITY_PORT"); ep != "" {
		var err error
		port, err = strconv.Atoi(ep)
		if err != nil {
			port = 9090
		}
	}
	if *fp != port {
		port = *fp
	}
	return file, port
}

func setHeaders(res http.ResponseWriter, ct string, cache string, code int) {
	res.Header().Add("Cache-Control", cache)
	res.Header().Add("Content-Type", ct)
	res.Header().Add("X-Content-Type-Options", "nosniff")
	res.Header().Add("X-Go-Vanity-Server-Build", buildCode)
	res.Header().Add("X-Go-Vanity-Server-Version", coreVersion)
	res.WriteHeader(code)
}

func logMiddleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/ping" {
			addr := r.RemoteAddr
			if xf := r.Header.Get("X-Real-Ip"); xf != "" {
				addr = xf
			}
			log.Printf("%s %s %s [%s]\n", addr, r.Method, r.URL, r.UserAgent())
		}
		handler.ServeHTTP(w, r)
	})
}

func getServerMux() *http.ServeMux {
	mux := http.NewServeMux()

	// Ping
	mux.HandleFunc("/api/ping", func(res http.ResponseWriter, _ *http.Request) {
		setHeaders(res, "text/plain; charset=utf-8", globalHandler.cache(), http.StatusOK)
		_, _ = res.Write([]byte("pong"))
	})

	// Version
	mux.HandleFunc("/api/version", func(res http.ResponseWriter, _ *http.Request) {
		js, _ := json.MarshalIndent(versionInfo(), "", "  ")
		setHeaders(res, "application/json", globalHandler.cache(), http.StatusOK)
		_, _ = res.Write(js)
	})

	// Configuration
	mux.HandleFunc("/api/conf", func(res http.ResponseWriter, _ *http.Request) {
		js, _ := json.MarshalIndent(globalConfig, "", "  ")
		setHeaders(res, "application/json", globalHandler.cache(), http.StatusOK)
		_, _ = res.Write(js)
	})

	// Configuration Panel
	mux.HandleFunc("/config/panel", func(res http.ResponseWriter, _ *http.Request) {
		panel, err := globalHandler.getConfigPanel()
		if err != nil {
			setHeaders(res, "text/plain; charset=utf-8", globalHandler.cache(), http.StatusInternalServerError)
			_, _ = res.Write([]byte(err.Error()))
			return
		}
		setHeaders(res, "text/html; charset=utf-8", globalHandler.cache(), http.StatusOK)
		_, _ = res.Write(panel)
	})

	// Update Configuration API
	mux.HandleFunc("/api/config", func(res http.ResponseWriter, req *http.Request) {
		if req.Method != "POST" {
			setHeaders(res, "text/plain; charset=utf-8", globalHandler.cache(), http.StatusMethodNotAllowed)
			_, _ = res.Write([]byte("Method not allowed"))
			return
		}

		body, err := io.ReadAll(req.Body)
		if err != nil {
			setHeaders(res, "text/plain; charset=utf-8", globalHandler.cache(), http.StatusBadRequest)
			_, _ = res.Write([]byte("Failed to read request body"))
			return
		}

		var newConfig Configuration
		if err := json.Unmarshal(body, &newConfig); err != nil {
			setHeaders(res, "text/plain; charset=utf-8", globalHandler.cache(), http.StatusBadRequest)
			_, _ = res.Write([]byte("Invalid JSON format"))
			return
		}

		// 验证配置
		if len(newConfig.Paths) == 0 {
			setHeaders(res, "text/plain; charset=utf-8", globalHandler.cache(), http.StatusBadRequest)
			_, _ = res.Write([]byte("At least one path configuration is required"))
			return
		}

		// 更新全局配置
		globalConfig = &newConfig
		if err := globalHandler.updateConfig(globalConfig); err != nil {
			setHeaders(res, "text/plain; charset=utf-8", globalHandler.cache(), http.StatusInternalServerError)
			_, _ = res.Write([]byte("Failed to update configuration"))
			return
		}

		setHeaders(res, "application/json", globalHandler.cache(), http.StatusOK)
		response := map[string]string{"status": "success", "message": "Configuration updated successfully"}
		js, _ := json.Marshal(response)
		_, _ = res.Write(js)
	})

	// Reload Configuration API
	mux.HandleFunc("/api/config/reload", func(res http.ResponseWriter, req *http.Request) {
		if req.Method != "POST" {
			setHeaders(res, "text/plain; charset=utf-8", globalHandler.cache(), http.StatusMethodNotAllowed)
			_, _ = res.Write([]byte("Method not allowed"))
			return
		}

		// 这里可以添加从文件重新加载配置的逻辑
		// 目前只是返回成功状态
		setHeaders(res, "application/json", globalHandler.cache(), http.StatusOK)
		response := map[string]string{"status": "success", "message": "Configuration reloaded successfully"}
		js, _ := json.Marshal(response)
		_, _ = res.Write(js)
	})

	// Main index
	mux.HandleFunc("/index.html", func(res http.ResponseWriter, _ *http.Request) {
		index, err := globalHandler.getIndex()
		if err != nil {
			setHeaders(res, "text/plain; charset=utf-8", globalHandler.cache(), http.StatusInternalServerError)
			_, _ = res.Write([]byte(err.Error()))
			return
		}
		setHeaders(res, "text/html; charset=utf-8", globalHandler.cache(), http.StatusOK)
		_, _ = res.Write(index)
	})

	// Catch-all path
	mux.HandleFunc("/", func(res http.ResponseWriter, req *http.Request) {
		repo, err := globalHandler.getRepo(strings.TrimSuffix(req.URL.Path, "/"))
		if err != nil {
			setHeaders(res, "text/plain; charset=utf-8", globalHandler.cache(), http.StatusNotFound)
			_, _ = res.Write([]byte(err.Error()))
			return
		}
		setHeaders(res, "text/html; charset=utf-8", globalHandler.cache(), http.StatusOK)
		_, _ = res.Write(repo)
	})

	return mux
}
