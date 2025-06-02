// package main

// import (
// 	"fmt"
// 	"log"
// 	"plugin"

// 	"example/pluginapi" // replace with your actual module path

// 	"github.com/gin-gonic/gin"
// )

// func main() {
// 	// Create Gin router
// 	router := gin.Default()

// 	// Path to your plugin .so file (adjust accordingly)
// 	soPath := "kubestellar-cluster-plugin.so"

// 	// Load the plugin dynamically
// 	pluginInstance, err := loadPlugin(soPath, router)
// 	if err != nil {
// 		log.Fatalf("Failed to load plugin: %v", err)
// 	}

// 	// Print plugin metadata info
// 	meta := pluginInstance.GetMetadata()
// 	fmt.Printf("Loaded plugin: %s (Version: %s, ID: %s)\n", meta.Name, meta.Version, meta.ID)

// 	// Start HTTP server
// 	fmt.Println("Starting server on :8090")
// 	if err := router.Run(":8090"); err != nil {
// 		log.Fatalf("Failed to start server: %v", err)
// 	}
// }

// // loadPlugin loads the plugin from .so file and registers handlers on the router
// func loadPlugin(soPath string, router *gin.Engine) (pluginapi.KubestellarPlugin, error) {
// 	p, err := plugin.Open(soPath)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to open plugin: %w", err)
// 	}

// 	// Lookup NewPlugin symbol
// 	const PluginSymbol = "NewPlugin"
// 	sym, err := p.Lookup(PluginSymbol)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to find NewPlugin symbol: %w", err)
// 	}

// 	// Assert symbol type to func() pluginapi.KubestellarPlugin
// 	newPluginFunc, ok := sym.(func() pluginapi.KubestellarPlugin)
// 	if !ok {
// 		return nil, fmt.Errorf("NewPlugin has wrong type signature")
// 	}

// 	// Call constructor to get plugin instance
// 	pluginInstance := newPluginFunc()

// 	// Initialize plugin with empty config (or real config)
// 	if err := pluginInstance.Initialize(map[string]interface{}{}); err != nil {
// 		return nil, fmt.Errorf("plugin initialization failed: %w", err)
// 	}

// 	// Register plugin handlers with Gin router, prefix with /api/plugins/{pluginID}
// 	pluginID := pluginInstance.GetMetadata().ID
// 	for path, handler := range pluginInstance.GetHandlers() {
// 		fullPath := fmt.Sprintf("/api/plugins/%s%s", pluginID, path)

// 		// Simple heuristic for HTTP method (GET for /status, POST otherwise)
// 		if path == "/status" {
// 			router.GET(fullPath, handler)
// 		} else {
// 			router.POST(fullPath, handler)
// 		}

// 		fmt.Printf("Registered route %s\n", fullPath)
// 	}

// 	return pluginInstance, nil
// }

package main

import (
	"fmt"
	"log"
	"net/http"
	"plugin"

	"github.com/gin-gonic/gin"

	"github.com/Per0x1de-1337/pluginapi" // replace with your actual module path
)

func main() {
	plug, err := plugin.Open("kubestellar-cluster-plugin.so")
	if err != nil {
		log.Fatalf("Error opening plugin: %v", err)
	}

	symNewPlugin, err := plug.Lookup("NewPlugin")
	if err != nil {
		log.Fatalf("Error looking up NewPlugin: %v", err)
	}

	newPluginFunc, ok := symNewPlugin.(func() pluginapi.KubestellarPlugin)
	if !ok {
		log.Fatal("NewPlugin has wrong signature")
	}

	kubePlugin := newPluginFunc()

	if err := kubePlugin.Initialize(nil); err != nil {
		log.Fatalf("Plugin initialization failed: %v", err)
	}

	// Set up Gin router
	router := gin.Default()

	// Register handlers from plugin
	for path, handler := range kubePlugin.GetHandlers() {
		switch path {
		case "/onboard", "/detach":
			router.POST(path, handler)
		case "/status":
			router.GET(path, handler)
		case "/available":
			router.GET(path, handler)
		default:
			log.Printf("Skipping unrecognized path: %s", path)
		}
	}

	// Log plugin metadata
	meta := kubePlugin.GetMetadata()
	fmt.Printf("✅ Loaded plugin: %s (%s)\n📄 Description: %s\n👤 Author: %s\n\n",
		meta.Name, meta.Version, meta.Description, meta.Author)

	// Start HTTP server
	if err := router.Run(":8080"); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server error: %v", err)
	}
}
