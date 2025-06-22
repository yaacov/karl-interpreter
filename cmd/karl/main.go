package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/yaacov/karl-interpreter/pkg/karl"
	"gopkg.in/yaml.v2"
)

func main() {
	var (
		outputFormat = flag.String("format", "yaml", "Output format: yaml or json")
		prettyOutput = flag.Bool("pretty", false, "Pretty print output")
		showHelp     = flag.Bool("help", false, "Show help message")
	)
	flag.Parse()

	if *showHelp {
		printHelp()
		return
	}

	args := flag.Args()
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "Error: No KARL rule provided\n")
		printUsage()
		os.Exit(1)
	}

	karlRule := args[0]

	// Create interpreter and parse the rule
	interpreter := karl.NewKARLInterpreter()
	err := interpreter.Parse(karlRule)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Parse error: %v\n", err)
		os.Exit(1)
	}

	// Validate the rule
	err = interpreter.Validate()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Validation error: %v\n", err)
		os.Exit(1)
	}

	// Convert to Kubernetes affinity
	affinity, err := interpreter.ToAffinity()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Conversion error: %v\n", err)
		os.Exit(1)
	}

	// Output in requested format
	switch *outputFormat {
	case "json":
		if *prettyOutput {
			output, err := json.MarshalIndent(affinity, "", "  ")
			if err != nil {
				fmt.Fprintf(os.Stderr, "JSON marshal error: %v\n", err)
				os.Exit(1)
			}
			fmt.Println(string(output))
		} else {
			output, err := json.Marshal(affinity)
			if err != nil {
				fmt.Fprintf(os.Stderr, "JSON marshal error: %v\n", err)
				os.Exit(1)
			}
			fmt.Println(string(output))
		}
	case "yaml":
		output, err := yaml.Marshal(affinity)
		if err != nil {
			fmt.Fprintf(os.Stderr, "YAML marshal error: %v\n", err)
			os.Exit(1)
		}
		fmt.Print(string(output))
	default:
		fmt.Fprintf(os.Stderr, "Error: Unknown format '%s'. Use 'yaml' or 'json'\n", *outputFormat)
		os.Exit(1)
	}
}

func printHelp() {
	fmt.Printf(`KARL - Kubernetes Affinity Rule Language Interpreter

USAGE:
    karl [OPTIONS] "<KARL_RULE>"

DESCRIPTION:
    Convert a single KARL rule into Kubernetes Affinity/Anti-Affinity YAML or JSON.
    Outputs only the affinity structure, ready to be used in pod specifications.

KARL RULE SYNTAX:
    RULE_TYPE TARGET_SELECTOR on TOPOLOGY [weight=N]

    Rule Types:
        REQUIRE  - Hard affinity constraint (must schedule near target pods)
        PREFER   - Soft affinity constraint (prefer to schedule near target pods with weight)
        AVOID    - Hard anti-affinity constraint (must not schedule near target pods)
        REPEL    - Soft anti-affinity constraint (prefer not to schedule near target pods with weight)

    Target Selectors:
        pods(label_selector)     - Select pods by expressive label selectors:
          pods(app=web)          - Simple equality
          pods(app=web,tier=frontend) - Multiple labels (AND operation)
          pods(app in [web,api]) - Label value in list
          pods(app not in [batch,test]) - Label value not in list
          pods(has monitoring)   - Label exists
          pods(not has debug)    - Label does not exist

    Topology Keys:
        node     - Same Kubernetes node (kubernetes.io/hostname)
        zone     - Same availability zone (topology.kubernetes.io/zone)
        region   - Same region (topology.kubernetes.io/region)
        rack     - Same rack (topology.kubernetes.io/rack)

OPTIONS:
    -format string
            Output format: yaml or json (default "yaml")
    -pretty
            Pretty print output (for JSON)
    -help
            Show this help message

EXAMPLES:
    # Hard affinity: require database pods on same node
    karl "REQUIRE pods(app=database) on node"

    # Soft anti-affinity: spread web/frontend pods across zones
    karl "REPEL pods(app in [web,frontend]) on zone weight=80"

    # Hard anti-affinity: avoid pods with debug labels
    karl "AVOID pods(has debug) on node"

    # Soft affinity: prefer to be near production workloads
    karl "PREFER pods(not has test) on zone weight=90"

    # Complex label selector with multiple conditions
    karl "REQUIRE pods(app=web,tier=frontend,env not in [test,debug]) on node"

OUTPUT:
    The tool outputs only the affinity structure in YAML or JSON format.
    You can use this output directly in your pod specifications under spec.affinity.
`)
}

func printUsage() {
	fmt.Printf(`Usage: karl [OPTIONS] "<KARL_RULE>"

Try 'karl -help' for more information.
`)
}
