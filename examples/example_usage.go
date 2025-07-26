//go:build ignore

// Package main demonstrates basic usage of the libconfig library.
package main

import (
	"fmt"
	"log"

	"github.com/kuzmik/go-libconfig"
)

func main() {
	// Example 1: Parse from string
	fmt.Println("=== Example 1: Parse from String ===")

	configStr := `
		name = "MyApp";
		version = "1.0";
		port = 8080;
		debug = true;

		database = {
			host = "localhost";
			port = 5432;
			credentials = {
				username = "admin";
				password = "secret";
			};
		};

		servers = [ "web1", "web2", "web3" ];
		weights = [ 100, 80, 60 ];
	`

	config, err := libconfig.ParseString(configStr)
	if err != nil {
		log.Fatal("Failed to parse config string:", err)
	}

	// Look up various values
	name, _ := config.LookupString("name")
	fmt.Printf("Application name: %s\n", name)

	port, _ := config.LookupInt("port")
	fmt.Printf("Port: %d\n", port)

	debug, _ := config.LookupBool("debug")
	fmt.Printf("Debug mode: %t\n", debug)

	dbHost, _ := config.LookupString("database.host")
	fmt.Printf("Database host: %s\n", dbHost)

	username, _ := config.LookupString("database.credentials.username")
	fmt.Printf("DB Username: %s\n", username)

	// Example 2: Working with arrays
	fmt.Println("\n=== Example 2: Working with Arrays ===")

	serversValue, err := config.Lookup("servers")
	if err == nil && serversValue.Type == libconfig.TypeArray {
		fmt.Printf("Found %d servers:\n", len(serversValue.ArrayVal))

		for i, server := range serversValue.ArrayVal {
			fmt.Printf("  Server %d: %s\n", i+1, server.StrVal)
		}
	}

	weightsValue, err := config.Lookup("weights")
	if err == nil && weightsValue.Type == libconfig.TypeArray {
		fmt.Printf("Server weights: ")

		for i, weight := range weightsValue.ArrayVal {
			if i > 0 {
				fmt.Print(", ")
			}

			fmt.Printf("%d", weight.IntVal)
		}

		fmt.Println()
	}

	// Example 3: Parse from file (if it exists)
	fmt.Println("\n=== Example 3: Parse from File ===")

	fileConfig, err := libconfig.ParseFile("example.cfg")
	if err != nil {
		fmt.Printf("Could not parse file (this is expected if running from different directory): %v\n", err)
	} else {
		appName, _ := fileConfig.LookupString("application.name")
		fmt.Printf("Application from file: %s\n", appName)

		buildNumber, _ := fileConfig.LookupInt64("build_number")
		fmt.Printf("Build number: %d\n", buildNumber)
	}
}
