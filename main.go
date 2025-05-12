package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"
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

	http.Handle("/", http.FileServer(http.Dir(staticDir)))
	log.Printf("Serving static files from: %s", staticDir)

	addr := fmt.Sprintf(":%d", port)
	log.Printf("GhostGate running on http://localhost%s\n", addr)
	err = http.ListenAndServe(addr, nil)
	if err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
