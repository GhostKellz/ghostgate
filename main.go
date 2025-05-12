package main

import (
	"flag"
	"io/ioutil"
	"log"
	"mime"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"gopkg.in/yaml.v2"
	"golang.org/x/crypto/acme/autocert"
	"golang.org/x/time/rate"
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

func serveStaticFiles(staticDir string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		filePath := filepath.Join(staticDir, r.URL.Path)
		info, err := os.Stat(filePath)
		if os.IsNotExist(err) {
			http.Error(w, "404 - Not Found", http.StatusNotFound)
			return
		}
		if err != nil {
			http.Error(w, "403 - Forbidden", http.StatusForbidden)
			return
		}
		if info.IsDir() {
			// Serve directory index
			files, err := os.ReadDir(filePath)
			if err != nil {
				http.Error(w, "403 - Forbidden", http.StatusForbidden)
				return
			}
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("<html><body><ul>"))
			for _, file := range files {
				w.Write([]byte("<li><a href=\"" + file.Name() + "\">" + file.Name() + "</a></li>"))
			}
			w.Write([]byte("</ul></body></html>"))
			return
		}
		// Serve file with MIME type detection
		mimeType := mime.TypeByExtension(filepath.Ext(filePath))
		if mimeType == "" {
			mimeType = "application/octet-stream"
		}
		w.Header().Set("Content-Type", mimeType)
		http.ServeFile(w, r, filePath)
	})
}

func rateLimitedProxy(proxy *httputil.ReverseProxy, limiter *rate.Limiter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !limiter.Allow() {
			http.Error(w, "429 - Too Many Requests", http.StatusTooManyRequests)
			return
		}
		// Inject custom headers
		r.Header.Set("X-Forwarded-For", r.RemoteAddr)
		proxy.ServeHTTP(w, r)
	}
}

func main() {
	configPath := flag.String("config", "gate.conf", "Path to main configuration file")
	confDir := flag.String("conf-dir", "conf.d", "Path to additional configuration directory")
	flag.Parse()

	reloadChan := make(chan os.Signal, 1)
	signal.Notify(reloadChan, syscall.SIGHUP)

	config, err := loadConfigs(*configPath, *confDir)
	if err != nil {
		log.Fatalf("Failed to load configurations: %v", err)
	}

	// Function to reload configurations
	reloadConfig := func() {
		log.Println("Reloading configurations...")
		newConfig, err := loadConfigs(*configPath, *confDir)
		if err != nil {
			log.Printf("Failed to reload configurations: %v", err)
			return
		}
		config = newConfig
		log.Println("Configurations reloaded successfully.")
	}

	// Start a goroutine to listen for reload signals
	go func() {
		for {
			<-reloadChan
			reloadConfig()
		}
	}()

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
		limiter := rate.NewLimiter(1, 5) // 1 request per second with a burst of 5
		http.HandleFunc(route.Path, rateLimitedProxy(proxy, limiter))
		log.Printf("Proxying path %s to backend: %s", route.Path, route.Backend)
	}

	// Replace static file handler with enhanced version
	http.Handle("/", serveStaticFiles(config.Server.StaticDir))

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
