package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"time"

	"compress/gzip"

	"github.com/fsnotify/fsnotify"
	"github.com/patrickmn/go-cache"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/crypto/acme/autocert"
	"golang.org/x/time/rate"
	"gopkg.in/yaml.v2"
)

const version = "1.0.0" // Define the version of GhostGate

type ProxyRoute struct {
	Path      string            `yaml:"path"`
	Backend   string            `yaml:"backend"`
	Regex     bool              `yaml:"regex"`
	RateLimit int               `yaml:"rate_limit"`
	Headers   map[string]string `yaml:"headers"`
}

type DomainConfig struct {
	Domain          string       `yaml:"domain"`
	DomainRegex     bool         `yaml:"domain_regex"`
	StaticDir       string       `yaml:"static_dir"`
	ProxyRoutes     []ProxyRoute `yaml:"proxy_routes"`
	Autocert        bool         `yaml:"autocert"`
	ACMEEmail       string       `yaml:"acme_email"`
	RedirectToHTTPS bool         `yaml:"redirect_to_https"`
	HSTS            bool         `yaml:"hsts"`
	CSP             string       `yaml:"csp"`
	AllowIPs        []string     `yaml:"allow_ips"`
	DenyIPs         []string     `yaml:"deny_ips"`
}

type Config struct {
	Server struct {
		Port     int    `yaml:"port"`
		TLSCert  string `yaml:"tls_cert"`
		TLSKey   string `yaml:"tls_key"`
		AdminAPI string `yaml:"admin_api"`
	} `yaml:"server"`
	Logging struct {
		Level  string `yaml:"level"`
		Format string `yaml:"format"`
	} `yaml:"logging"`
	Domains        []DomainConfig `yaml:"domains"`
	ReloadOnChange bool           `yaml:"reload_on_change"`
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
			mainConfig.Domains = append(mainConfig.Domains, additionalConfig.Domains...)
		}
	}

	return mainConfig, nil
}

// Add caching headers for static files
func serveStaticFilesWithCache(staticDir string) http.Handler {
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
		// Serve file with caching headers
		mimeType := mime.TypeByExtension(filepath.Ext(filePath))
		if mimeType == "" {
			mimeType = "application/octet-stream"
		}
		w.Header().Set("Content-Type", mimeType)
		w.Header().Set("Cache-Control", "public, max-age=3600")
		http.ServeFile(w, r, filePath)
	})
}

// Add health check endpoint
func serveHealthCheck() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
}

func serveWelcomePage() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`
			<html>
			<head><title>Welcome to GhostGate</title></head>
			<body>
			<h1>Welcome to GhostGate</h1>
			<p>If you see this page, the GhostGate server is running successfully.</p>
			<p>Configure your server by editing <code>gate.conf</code> or adding files to <code>conf.d/</code>.</p>
			</body>
			</html>
		`))
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

func setupLogger(level, format string) {
	var logOutput io.Writer = os.Stdout
	log.SetOutput(logOutput)

	// Set log format
	if strings.ToLower(format) == "json" {
		log.SetFlags(0) // Disable default timestamp
		log.SetOutput(io.MultiWriter(logOutput, &jsonLogWriter{}))
	} else {
		log.SetFlags(log.LstdFlags) // Default text format with timestamp
	}

	// Set log level (basic implementation)
	if strings.ToLower(level) == "error" {
		log.SetOutput(io.Discard) // Discard all logs except errors (placeholder)
	}
}

type jsonLogWriter struct{}

func (j *jsonLogWriter) Write(p []byte) (n int, err error) {
	logEntry := map[string]string{
		"message":   strings.TrimSpace(string(p)),
		"timestamp": time.Now().Format(time.RFC3339),
	}
	logJSON, _ := json.Marshal(logEntry)
	os.Stdout.Write(logJSON)
	os.Stdout.Write([]byte("\n"))
	return len(p), nil
}

func validateConfig(configPath string) error {
	log.Printf("Validating configuration file: %s", configPath)
	config, err := loadConfig(configPath)
	if err != nil {
		return fmt.Errorf("validation failed: %v", err)
	}

	// Check for required fields
	if config.Server.Port == 0 {
		return fmt.Errorf("server port is not defined")
	}
	if len(config.Domains) == 0 {
		return fmt.Errorf("at least one domain must be defined")
	}

	log.Println("Configuration is valid.")
	return nil
}

func reloadGhostGate() {
	pid := os.Getpid()
	log.Printf("Sending SIGHUP signal to process %d to reload configurations...", pid)
	if err := syscall.Kill(pid, syscall.SIGHUP); err != nil {
		log.Fatalf("Failed to send SIGHUP signal: %v", err)
	}
	log.Println("Reload signal sent successfully.")
}

func issueCertificate(domains []string) {
	log.Printf("Issuing certificate for domains: %v", domains)
	acmeArgs := []string{"--issue", "--dns", "dns_pdns"}
	for _, d := range domains {
		acmeArgs = append(acmeArgs, "-d", d)
	}
	acmeArgs = append(acmeArgs, "--dnssleep", "20", "--log", "--debug")
	cmd := exec.Command("acme.sh", acmeArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Fatalf("Failed to issue certificate for domains %v: %v", domains, err)
	}
	// Use the first domain for cert storage
	domainDir := filepath.Join("/etc/ghostgate/certs", domains[0])
	if err := os.MkdirAll(domainDir, 0755); err != nil {
		log.Fatalf("Failed to create directory for certificates: %v", err)
	}
	installArgs := []string{"--install-cert", "-d", domains[0],
		"--cert-file", filepath.Join(domainDir, "cert.pem"),
		"--key-file", filepath.Join(domainDir, "privkey.pem"),
		"--fullchain-file", filepath.Join(domainDir, "fullchain.pem"),
		"--reloadcmd", "ghostgate reload"}
	installCmd := exec.Command("acme.sh", installArgs...)
	installCmd.Stdout = os.Stdout
	installCmd.Stderr = os.Stderr
	if err := installCmd.Run(); err != nil {
		log.Fatalf("Failed to install certificate for domain %s: %v", domains[0], err)
	}
	log.Printf("Certificate successfully issued and installed for domains: %v", domains)
}

// Add a global cache instance
var responseCache = cache.New(5*time.Minute, 10*time.Minute)

// Middleware to handle reverse proxy response caching
func cacheMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cacheKey := r.URL.String()
		if cachedResponse, found := responseCache.Get(cacheKey); found {
			w.Write(cachedResponse.([]byte))
			return
		}

		recorder := httptest.NewRecorder()
		next.ServeHTTP(recorder, r)

		for k, v := range recorder.Header() {
			w.Header()[k] = v
		}
		w.WriteHeader(recorder.Code)
		responseBody := recorder.Body.Bytes()
		w.Write(responseBody)

		responseCache.Set(cacheKey, responseBody, cache.DefaultExpiration)
	})
}

// Middleware to add gzip compression
func gzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		w.Header().Set("Content-Encoding", "gzip")
		gz := gzip.NewWriter(w)
		defer gz.Close()
		gzipWriter := &gzipResponseWriter{Writer: gz, ResponseWriter: w}
		next.ServeHTTP(gzipWriter, r)
	})
}

type gzipResponseWriter struct {
	http.ResponseWriter
	io.Writer
}

func (w *gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func checkCertExpiry(certPath string) {
	certData, err := ioutil.ReadFile(certPath)
	if err != nil {
		log.Printf("[WARN] Could not read certificate for expiry check: %v", err)
		return
	}
	certs, err := tls.X509KeyPair(certData, certData)
	if err != nil {
		log.Printf("[WARN] Could not parse certificate for expiry check: %v", err)
		return
	}
	for _, cert := range certs.Certificate {
		parsed, err := x509.ParseCertificate(cert)
		if err == nil && time.Until(parsed.NotAfter) < 30*24*time.Hour {
			log.Printf("[WARN] Certificate expires soon: %s", parsed.NotAfter)
		}
	}
}

// hostHandler only serves requests for the given domain
func hostHandler(domain string, domainRegex bool, handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if domainRegex {
			matched, _ := regexp.MatchString(domain, r.Host)
			if matched {
				handler.ServeHTTP(w, r)
				return
			}
		} else {
			if r.Host == domain || strings.HasPrefix(r.Host, domain+":") {
				handler.ServeHTTP(w, r)
				return
			}
		}
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("404 - Not Found"))
	})
}

func routeHandler(route ProxyRoute, proxy *httputil.ReverseProxy) http.HandlerFunc {
	limiter := rate.NewLimiter(rate.Limit(route.RateLimit), 5)
	return func(w http.ResponseWriter, r *http.Request) {
		if route.RateLimit > 0 && !limiter.Allow() {
			http.Error(w, "429 - Too Many Requests", http.StatusTooManyRequests)
			return
		}
		for k, v := range route.Headers {
			r.Header.Set(k, v)
		}
		proxy.ServeHTTP(w, r)
	}
}

func watchFiles(paths []string, reloadFunc func()) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Printf("[WARN] Could not start file watcher: %v", err)
		return
	}
	defer watcher.Close()
	for _, p := range paths {
		watcher.Add(p)
	}
	for {
		select {
		case event := <-watcher.Events:
			if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create {
				log.Println("[INFO] Config or cert file changed, reloading...")
				reloadFunc()
			}
		case err := <-watcher.Errors:
			log.Printf("[WARN] File watcher error: %v", err)
		}
	}
}

func serveAdminAPI(reloadFunc func()) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" && r.URL.Path == "/admin/reload" {
			reloadFunc()
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Reloaded"))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	})
}

func getAutocertManager(domains []DomainConfig) *autocert.Manager {
	hosts := []string{}
	var email string
	for _, d := range domains {
		if d.Autocert {
			hosts = append(hosts, d.Domain)
			if d.ACMEEmail != "" {
				email = d.ACMEEmail
			}
		}
	}
	return &autocert.Manager{
		Cache:      autocert.DirCache("certs"),
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(hosts...),
		Email:      email,
	}
}

func loggingMiddleware(domain string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[ACCESS] domain=%s method=%s path=%s remote=%s", domain, r.Method, r.URL.Path, r.RemoteAddr)
		next.ServeHTTP(w, r)
	})
}

func securityMiddleware(domain DomainConfig, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		remoteIP := strings.Split(r.RemoteAddr, ":")[0]
		if len(domain.AllowIPs) > 0 {
			allowed := false
			for _, ip := range domain.AllowIPs {
				if ip == remoteIP {
					allowed = true
					break
				}
			}
			if !allowed {
				http.Error(w, "403 - Forbidden", http.StatusForbidden)
				return
			}
		}
		for _, ip := range domain.DenyIPs {
			if ip == remoteIP {
				http.Error(w, "403 - Forbidden", http.StatusForbidden)
				return
			}
		}
		if domain.HSTS {
			w.Header().Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains; preload")
		}
		if domain.CSP != "" {
			w.Header().Set("Content-Security-Policy", domain.CSP)
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	mode := flag.String("mode", "run", "Mode of operation: run, check, reload, cert, or version")
	configPath := flag.String("config", "ghostgate.conf", "Path to main configuration file")
	confDir := flag.String("conf-dir", "conf.d", "Path to additional configuration directory")
	domainsFlag := flag.String("domain", "", "Comma-separated domains for certificate issuance (used with cert mode)")
	flag.Parse()

	if *mode == "version" {
		fmt.Printf("GhostGate version %s\n", version)
		os.Exit(0)
	}

	if *mode == "reload" {
		reloadGhostGate()
		os.Exit(0)
	}

	if *mode == "cert" {
		if *domainsFlag == "" {
			log.Fatalf("Domain(s) required for certificate issuance. Use -domain flag.")
		}
		domains := strings.Split(*domainsFlag, ",")
		issueCertificate(domains)
		os.Exit(0)
	}

	if *mode == "check" {
		if err := validateConfig(*configPath); err != nil {
			log.Fatalf("Configuration validation failed: %v", err)
		}
		os.Exit(0)
	}

	reloadChan := make(chan os.Signal, 1)
	signal.Notify(reloadChan, syscall.SIGHUP)

	config, err := loadConfigs(*configPath, *confDir)
	if err != nil {
		log.Fatalf("Failed to load configurations: %v", err)
	}

	// Setup logging
	setupLogger(config.Logging.Level, config.Logging.Format)

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

	if config.ReloadOnChange {
		go watchFiles([]string{*configPath}, reloadConfig)
	}
	if config.Server.AdminAPI != "" {
		go func() {
			log.Printf("Starting admin API on %s", config.Server.AdminAPI)
			adminMux := http.NewServeMux()
			adminMux.Handle("/admin/reload", serveAdminAPI(reloadConfig))
			http.ListenAndServe(config.Server.AdminAPI, adminMux)
		}()
	}

	// Use configuration values
	port := config.Server.Port
	if port == 0 {
		port = 80 // Default to port 80 for HTTP
	}

	// Per-domain (virtual host) handler setup
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	for _, d := range config.Domains {
		if d.StaticDir != "" {
			h := gzipMiddleware(cacheMiddleware(serveStaticFilesWithCache(d.StaticDir)))
			h = loggingMiddleware(d.Domain, h)
			h = securityMiddleware(d, h)
			mux.Handle("/", hostHandler(d.Domain, d.DomainRegex, h))
		}
		for _, route := range d.ProxyRoutes {
			backendURL, err := url.Parse(route.Backend)
			if err != nil {
				log.Printf("Invalid backend URL for domain %s path %s: %v", d.Domain, route.Path, err)
				continue
			}
			proxy := httputil.NewSingleHostReverseProxy(backendURL)
			var handler http.Handler
			if route.Regex {
				handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					matched, _ := regexp.MatchString(route.Path, r.URL.Path)
					if matched {
						routeHandler(route, proxy).ServeHTTP(w, r)
						return
					}
					w.WriteHeader(http.StatusNotFound)
				})
			} else {
				handler = routeHandler(route, proxy)
			}
			h := handler
			h = loggingMiddleware(d.Domain, h)
			h = securityMiddleware(d, h)
			mux.Handle(route.Path, hostHandler(d.Domain, d.DomainRegex, h))
		}
	}

	// Serve welcome page if no domains are configured
	if len(config.Domains) == 0 {
		mux.Handle("/", serveWelcomePage())
		log.Println("Serving default Welcome to GhostGate page")
	} else {
		mux.Handle("/health", serveHealthCheck())
	}

	// HTTP→HTTPS redirection per domain
	httpMux := http.NewServeMux()
	for _, d := range config.Domains {
		if d.RedirectToHTTPS {
			httpMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
				http.Redirect(w, r, "https://"+r.Host+r.URL.String(), http.StatusMovedPermanently)
			})
		}
	}
	if len(config.Domains) == 0 {
		httpMux.Handle("/", serveWelcomePage())
	}

	// Start HTTP server to redirect to HTTPS or serve welcome
	httpServer := &http.Server{
		Addr:    ":80",
		Handler: httpMux,
	}
	go func() {
		log.Printf("Starting HTTP server on :80")
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server failed: %v", err)
		}
	}()

	// Set up TLS config with OCSP stapling and custom cert/key support
	var tlsConfig *tls.Config
	if config.Server.TLSCert != "" && config.Server.TLSKey != "" {
		checkCertExpiry(config.Server.TLSCert)
		cert, err := tls.LoadX509KeyPair(config.Server.TLSCert, config.Server.TLSKey)
		if err != nil {
			log.Fatalf("Failed to load TLS cert/key: %v", err)
		}
		tlsConfig = &tls.Config{
			Certificates:             []tls.Certificate{cert},
			MinVersion:               tls.VersionTLS12,
			NextProtos:               []string{"h2", "http/1.1"},
			PreferServerCipherSuites: true,
			GetCertificate: func(chi *tls.ClientHelloInfo) (*tls.Certificate, error) {
				return &cert, nil
			},
			// OCSP stapling is handled automatically if the cert supports it
		}
	} else {
		autoCertManager := getAutocertManager(config.Domains)
		tlsConfig = autoCertManager.TLSConfig()
	}

	// Start HTTPS server with custom TLS config
	httpsServer := &http.Server{
		Addr:      ":443",
		Handler:   mux,
		TLSConfig: tlsConfig,
	}

	log.Printf("Starting HTTPS server on :443")
	if err := httpsServer.ListenAndServeTLS("", ""); err != nil {
		log.Fatalf("HTTPS server failed: %v", err)
	}
}
