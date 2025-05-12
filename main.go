package main

import (
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/gorilla/handlers"
	"gopkg.in/yaml.v2"
	"golang.org/x/crypto/acme/autocert"
)

type Config struct {
	Server struct {
		Port      int    `yaml:"port"`
		StaticDir string `yaml:"static_dir"`
		Backend   string `yaml:"backend"`
	} `yaml:"server"`
	Logging struct {
		Level  string `yaml:"level"`
		Format string `yaml:"format"`
	} `yaml:"logging"`
	Proxy struct {
		Routes []struct {
			Path    string `yaml:"path"`
			Backend string `yaml:"backend"`
		} `yaml:"routes"`
	} `yaml:"proxy"`
}

func loadConfig(path string) (*Config, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}
	return &config, nil
}

func loadConfigs(mainConfigPath string, confDir string) (*Config, error) {
	mainConfig, err := loadConfig(mainConfigPath)
	if err != nil {
		return nil, err
	}

	files, err := ioutil.ReadDir(confDir)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".conf") {
			confPath := filepath.Join(confDir, file.Name())
			additionalConfig, err := loadConfig(confPath)
			if err != nil {
				log.Printf("Failed to load config %s: %v", confPath, err)
				continue
			}

			// Merge additionalConfig into mainConfig
			mainConfig.Proxy.Routes = append(mainConfig.Proxy.Routes, additionalConfig.Proxy.Routes...)
		}
	}

	return mainConfig, nil
}

func main() {
	configPath := flag.String("config", "gate.conf", "Path to main configuration file")
	confDir := flag.String("conf-dir", "conf.d", "Path to additional configuration directory")
	flag.Parse()

	config, err := loadConfigs(*configPath, *confDir)
	if err != nil {
		log.Fatalf("Failed to load configurations: %v", err)
	}

	// Use configuration values
	port := config.Server.Port
	if port == 0 {
		port = 80 // Default to port 80 for HTTP
	}
	staticDir := config.Server.StaticDir

	// Validate directory
	if _, err := os.Stat(staticDir); os.IsNotExist(err) {
		log.Fatalf("Directory does not exist: %s", staticDir)
	}

	// Set up handlers based on configuration
	for _, route := range config.Proxy.Routes {
		backendURL, err := url.Parse(route.Backend)
		if err != nil {
			log.Printf("Invalid backend URL for path %s: %v", route.Path, err)
			continue
		}
		proxy := httputil.NewSingleHostReverseProxy(backendURL)
		http.HandleFunc(route.Path, func(w http.ResponseWriter, r *http.Request) {
			proxy.ServeHTTP(w, r)
		})
		log.Printf("Proxying path %s to backend: %s", route.Path, route.Backend)
	}

	// Add gzip compression and caching for static files
	fileServer := http.FileServer(http.Dir(staticDir))
	gzipHandler := handlers.CompressHandler(fileServer)
	cacheHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "public, max-age=31536000")
		gzipHandler.ServeHTTP(w, r)
	})
	http.Handle("/", cacheHandler)
	log.Printf("Serving static files from: %s", staticDir)

	// Set up autocert manager for Let's Encrypt
	autoCertManager := &autocert.Manager{
		Cache:      autocert.DirCache("certs"), // Directory to store certificates
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist("example.com", "www.example.com"), // Replace with your domain(s)
	}

	// Start HTTPS server with autocert
	httpsServer := &http.Server{
		Addr:      ":443",
		Handler:   nil,
		TLSConfig: autoCertManager.TLSConfig(),
	}

	// Start HTTP server to redirect to HTTPS
	httpServer := &http.Server{
		Addr: ":80",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "https://"+r.Host+r.URL.String(), http.StatusMovedPermanently)
		}),
	}

	go func() {
		log.Printf("Starting HTTP server on :80")
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server failed: %v", err)
		}
	}()

	log.Printf("Starting HTTPS server on :443")
	if err := httpsServer.ListenAndServeTLS("", ""); err != nil {
		log.Fatalf("HTTPS server failed: %v", err)
	}
}
