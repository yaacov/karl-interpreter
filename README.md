# KARL - Kubernetes Affinity Rule Language

[![Go Report Card](https://goreportcard.com/badge/github.com/yaacov/karl-interpreter)](https://goreportcard.com/report/github.com/yaacov/karl-interpreter)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.24-blue)](https://golang.org/)

KARL (Kubernetes Affinity Rule Language) is a human-readable domain-specific language for expressing Kubernetes pod affinity and anti-affinity rules. It simplifies the complex YAML syntax into intuitive, English-like statements.

## 🚀 Quick Start

### Installation

```bash
# Install using go install
go install github.com/yaacov/karl-interpreter/cmd/karl@latest

# Or download binary from releases
curl -L https://github.com/yaacov/karl-interpreter/releases/latest/download/karl-linux-amd64 -o karl
chmod +x karl
```

### Usage

Convert a single KARL rule to Kubernetes affinity YAML:

```bash
# Hard constraint: database pods must be on same node for performance
karl "REQUIRE pods(app=database) on node"

# Soft constraint: spread web pods across different nodes for availability
karl "REPEL pods(app in [web,frontend]) on node weight=100"

# Hard constraint: cache pods must be in same zone for low latency
karl "REQUIRE pods(tier=cache) on zone"

# Hard anti-affinity: avoid pods with debug labels
karl "AVOID pods(has debug) on node"

# Soft anti-affinity: prefer not to run with batch workloads
karl "REPEL pods(type not in [production,critical]) on zone weight=80"
```

The tool outputs only the affinity structure that can be used directly in pod specifications.

## 📖 Language Syntax

### Basic Structure

```karl
RULE_TYPE TARGET_SELECTOR on TOPOLOGY [weight=N]
```

### Rule Types

- **REQUIRE** - Hard affinity constraint (must schedule near target pods)
- **PREFER** - Soft affinity constraint (prefer to schedule near target pods with weight)
- **AVOID** - Hard anti-affinity constraint (must not schedule near target pods)
- **REPEL** - Soft anti-affinity constraint (prefer not to schedule near target pods with weight)

### Target Selectors

- `pods(label_selector)` - Select pods by labels:
  - `pods(app=web)` - Simple equality match
  - `pods(app=web,tier=frontend)` - Multiple labels (AND)
  - `pods(app in [web,api])` - Label value in list
  - `pods(app not in [batch,test])` - Label value not in list
  - `pods(has monitoring)` - Label exists
  - `pods(not has debug)` - Label does not exist

### Topology Keys

- `node` - Same Kubernetes node (`kubernetes.io/hostname`)
- `zone` - Same availability zone (`topology.kubernetes.io/zone`)
- `region` - Same region (`topology.kubernetes.io/region`)
- `rack` - Same rack (`topology.kubernetes.io/rack`)

## 🎯 Label Selectors

KARL supports human-readable label selector syntax that's more intuitive than Kubernetes YAML:

```karl
# Simple equality
REQUIRE pods(app=database) on node

# Multiple labels (AND operation)
REQUIRE pods(app=web,tier=frontend) on zone

# Value in list (OR operation)
REPEL pods(app in [test,staging,dev]) on node weight=90

# Value not in list
PREFER pods(env not in [test,debug]) on zone weight=80

# Label exists
AVOID pods(has experimental) on node

# Label does not exist
REQUIRE pods(not has deprecated) on zone
```

### Comparison with Kubernetes YAML

**KARL:**

```karl
REPEL pods(app in [web,api],not has debug) on zone weight=80
```

**Equivalent Kubernetes YAML:**

```yaml
podAntiAffinity:
  preferredDuringSchedulingIgnoredDuringExecution:
  - weight: 80
    podAffinityTerm:
      labelSelector:
        matchLabels:
          # Multiple matchLabels for AND
        matchExpressions:
        - key: app
          operator: In
          values: [web, api]
        - key: debug
          operator: DoesNotExist
      topologyKey: topology.kubernetes.io/zone
```

## 📝 Examples

### Database High Availability

```karl
# Primary database must avoid other primaries (hard anti-affinity)
AVOID pods(db-role=primary) on node;

# Replicas should prefer to spread across zones (soft anti-affinity)
REPEL pods(app in [db-replica,db-secondary]) on zone weight=100;

# All database pods must be in high-memory zones (hard affinity)
REQUIRE pods(has database) on zone
```

### Microservices Co-location

```karl
# Web tier prefers to be near API tier (soft affinity)
PREFER pods(tier=api) on zone weight=80;

# Web tier should spread across nodes for availability (soft anti-affinity)
REPEL pods(app in [web,frontend]) on node weight=100;

# Avoid being on same node as batch jobs (soft anti-affinity)
REPEL pods(type not in [production,critical]) on node weight=90
```

### Multi-Zone Deployment

```karl
# Spread across zones for disaster recovery (soft anti-affinity)
REPEL pods(app=frontend) on zone weight=100;

# Keep session store close to app (hard affinity)
REQUIRE pods(component=session-store) on node;

# Avoid noisy neighbor workloads (soft anti-affinity)
REPEL pods(not has production) on node weight=80
```

## 🛠 Go API Usage

### Basic Usage

```go
package main

import (
    "fmt"
    "github.com/yaacov/karl-interpreter/pkg/karl"
)

func main() {
    interpreter := karl.NewKARLInterpreter()
    
    karlRule := "REPEL pods(app in [test,debug]) zone weight=80"
    
    // Parse single KARL rule
    err := interpreter.Parse(karlRule)
    if err != nil {
        panic(err)
    }
    
    // Convert to Kubernetes affinity
    affinity, err := interpreter.ToAffinity()
    if err != nil {
        panic(err)
    }
    
    // Use in pod specification
    podSpec.Affinity = affinity
}
```

### Advanced Usage with Validation

```go
interpreter := karl.NewKARLInterpreter()

// Parse single rule
karlRule := "REPEL pods(app=web) zone weight=75"
err := interpreter.Parse(karlRule)
if err != nil {
    log.Fatalf("Parse error: %v", err)
}

// Validate rule
err = interpreter.Validate()
if err != nil {
    log.Fatalf("Validation error: %v", err)
}

// Convert to affinity
affinity, _ := interpreter.ToAffinity()
```

## 🧪 Testing

Run the test suite:

```bash
# Unit tests
go test ./pkg/karl/

# All tests with coverage
make test
```

## 🔧 CLI Commands

### Parse and Convert

```bash
# Parse KARL rule and output as YAML affinity
karl "ATTRACT pods(app=database) on node"

# Output as JSON
karl -format json "REPEL pods(type=batch) on node"

# Pretty JSON output
karl -format json -pretty "PREFER pods(app=web) not_on zone weight=80"
```

## 📄 License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.
