# KPM - Kotlin Package Manager

KPM is a package manager designed for Kotlin projects, providing dependency management, package installation, and project lifecycle management similar to npm for Node.js or Cargo for Rust.

## 📋 Table of Contents

- [Overview](#overview)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [Commands](#commands)
- [Configuration Files](#configuration-files)
- [Project Structure](#project-structure)
- [Features](#features)
- [License](#license)

## 🎯 Overview

KPM streamlines Kotlin project management by providing a unified command-line interface for:

- **Project Initialization**: Create and configure new Kotlin projects
- **Dependency Management**: Automatically scan and manage project dependencies
- **Package Installation**: Download and install packages from Maven Central Repository
- **Package Search**: Search Maven Central for available packages
- **Script Execution**: Run custom build and development scripts
- **Language Server Integration**: Automatic Kotlin Language Server configuration
- **Project Maintenance**: Clean and sync project resources

## 📦 Installation

### Prerequisites

- Go 1.25.1 or higher
- Access to Maven Central Repository

### Build from Source

```bash
git clone <repository-url>
cd KPM
go build -o kpm main.go
```

The compiled binary (`kpm` or `kpm.exe` on Windows) can then be used as a command-line tool.

## 🚀 Quick Start

### 1. Initialize a New Project

```bash
kpm init
```

This interactive command will:
- Ask if you want to scan existing dependencies
- Prompt for project name (defaults to folder name)
- Set the main source directory (default: `./src`)
- Configure privacy settings
- Set initial version (default: `0.0.0`)

This generates:
- `package.kpm` - Project configuration file
- `.kls_classpath` - Kotlin Language Server classpath

### 2. Scan for Dependencies

```bash
kpm scan
# or
kpm extract
# or
kpm -s
```

Scans all Kotlin (`.kt`) files in your project and extracts import statements to identify dependencies.

### 3. Install Dependencies

```bash
# Install packages from package.kpm
kpm install

# Install a specific Maven package
kpm install org.jetbrains.kotlinx:kotlinx-coroutines-core:1.10.2

# Update packages
kpm update

# Install from URL
kpm install --url package_name https://example.com/package.jar

# Install from Maven
kpm install --maven org.jetbrains.kotlinx:kotlinx-coroutines-core:1.10.2
```

### 4. List Installed Packages

```bash
kpm list
# or
kpm -l
```

Displays a formatted table of all installed packages with their version, group, and paths.

### 5. Search for Packages

```bash
kpm search
# or
kpm -f
```

Search Maven Central Repository for available packages.

### 6. Run Scripts

Define scripts in `package.kpm`:

```json
{
  "scripts": {
    "dev": "command-to-run",
    "build": "another-command"
  }
}
```

Then execute:

```bash
kpm run dev
kpm run build
```

## 📖 Commands

| Command | Aliases | Description |
|---------|---------|-------------|
| `init` | `-i`, `--init` | Initialize a new KPM project |
| `install` | `i`, `get`, `-g` | Install packages from package.kpm or specified packages |
| `update` | — | Update existing packages |
| `scan` | `extract`, `-s`, `-e` | Scan project for dependencies |
| `list` | `-l`, `--list` | List installed packages in table format |
| `search` | `-f`, `--find` | Search Maven Central Repository |
| `run` | — | Execute scripts defined in package.kpm |
| `clean` | `--clean` | Clean project files |
| `sync` | — | Synchronize project resources |
| `version` | `-v`, `--version` | Display KPM version |
| `help` | `-h`, `--help` | Show help message |

## 📁 Configuration Files

### `package.kpm`

Project manifest file containing metadata and dependencies.

```json
{
  "name": "MyKotlinProject",
  "private": false,
  "version": "0.0.0",
  "path": "/home/user/MyKotlinProject",
  "mainDir": "./src",
  "dependencies": {
    "org.jetbrains.kotlinx:kotlinx-coroutines-core": "1.10.2",
    "io.kvision:kvision-server-javalin": "8.2.0"
  },
  "scripts": {
    "dev": "gradle run",
    "build": "gradle build"
  }
}
```

**Fields:**
- `name` - Project name
- `private` - Whether the project is private
- `version` - Semantic version (x.y.z format)
- `path` - Absolute path to project directory
- `mainDir` - Main source directory (relative path)
- `dependencies` - Map of package identifiers to versions
- `scripts` - Map of script names to commands

### `resource.kpm`

Stores information about downloaded and installed packages.

```json
{
  "kpm version": "1.0.0",
  "resources": [
    {
      "group": "org.jetbrains.kotlinx",
      "name": "kotlinx-coroutines-core",
      "version": "1.10.2",
      "source": "maven",
      "type": "jar",
      "domain": "https://repo1.maven.org/maven2/",
      "path": "https://repo1.maven.org/maven2/org/jetbrains/kotlinx/kotlinx-coroutines-core/1.10.2/kotlinx-coroutines-core-1.10.2.jar",
      "lpath": "./libs/org.jetbrains.kotlinx/kotlinx-coroutines-core/1.10.2/kotlinx-coroutines-core-1.10.2.jar",
      "gpath": "/home/user/project/libs/org.jetbrains.kotlinx/kotlinx-coroutines-core/1.10.2/kotlinx-coroutines-core-1.10.2.jar",
      "hash": "sha256:82af21f1c0e8ce74c5c..."
    }
  ]
}
```

**Resource Fields:**
- `group` - Maven group ID
- `name` - Package name/artifact ID
- `version` - Package version
- `source` - Package source (e.g., "maven")
- `type` - File type (e.g., "jar")
- `domain` - Base URL domain
- `path` - Full remote URL
- `lpath` - Local relative path
- `gpath` - Global absolute path
- `hash` - SHA256 hash for integrity verification

## 📂 Project Structure

```
KPM/
├── main.go                    # Main entry point and CLI router
├── go.mod                     # Go module definition
├── go.sum                     # Go dependencies checksums
├── package.kpm                # Project metadata
├── resource.kpm               # Installed packages manifest
├── LICENCE.md                 # Project license (MPEL v1.0)
├── README.md                  # This file
│
├── types/                     # Type definitions
│   ├── resource.go            # Resource struct definitions
│   ├── root.go                # Root directory detection
│   └── maven.go               # Maven-related types
│
├── init/                      # Project initialization
│   └── main.go                # Init command implementation
│
├── install/                   # Package installation
│   ├── main.go                # Install/update logic
│   ├── downloadFile.go        # File download utilities
│   ├── downloadPackage.go     # Package download handler
│   ├── getXmlFromMaven.go     # Maven XML parsing
│   ├── makeInstallFile.go     # Installation file generation
│   ├── loadresources.go       # Resource loading
│   └── install_test.go        # Tests
│
├── scanner/                   # Dependency scanning
│   ├── main.go                # Scanner orchestration
│   └── get_package.go         # Package detection
│
├── libscanner/                # Library scanning utilities
│   ├── kt_file_finder.go      # Kotlin file discovery
│   ├── main.go                # Scanner entry point
│   ├── readImports.go         # Import statement parsing
│   └── write_to_json.go       # JSON output formatting
│
├── list/                      # Package listing
│   └── main.go                # List command implementation
│
├── search/                    # Package search
│   └── main.go                # Search in Maven Central
│
├── run/                       # Script execution
│   └── main.go                # Run command implementation
│
├── clean/                     # Project cleaning
│   └── main.go                # Clean command implementation
│
├── sync/                      # Synchronization
│   └── main.go                # Sync command implementation
│
├── kotlin_language_server/    # Language Server support
│   ├── main.go                # KLS initialization
│   └── vscode.go              # VSCode integration
│
├── libs/                      # Downloaded package cache
│   ├── org.jetbrains.kotlinx/
│   │   └── kotlinx-coroutines-core/
│   │       └── 1.10.2/
│   └── io.kvision/
│       └── kvision-server-javalin/
│           └── 8.2.0/
│
├── .kls_classpath             # Kotlin Language Server classpath
├── .gitignore                 # Git ignore rules
└── .vscode/                   # VSCode configuration
```

## ✨ Features

### Automatic Dependency Discovery

KPM scans your Kotlin source files (`.kt`) to automatically identify import statements and detect dependencies from:
- Maven repositories
- External URLs
- Local files

### Maven Central Integration

Seamlessly interact with Maven Central Repository:
- Search for packages
- Download specific versions
- Resolve transitive dependencies
- Verify package integrity with SHA256 hashes

### Language Server Support

Automatic Kotlin Language Server (KLS) configuration:
- Generates `.kls_classpath` on project initialization
- Maintains classpath for IDE integration
- Updates classpath when packages are installed

### Package Caching

Efficient local caching system:
- Stores downloaded packages in structured `libs/` directory
- Maintains both local and global paths
- Enables offline usage of cached packages

### Script Management

Define and run custom scripts in `package.kpm`:
- Build scripts
- Development server commands
- Testing utilities
- Custom tooling

### Clean & Sync Operations

- **Clean**: Remove generated files and clean up project
- **Sync**: Synchronize installed packages with configuration

## 🔧 Dependencies

Key Go dependencies:

- `github.com/dgraph-io/badger/v3` - Embedded key-value database (alternative storage)
- `github.com/scagogogo/sonatype-central-sdk` - Maven Central API integration
- `github.com/shirou/gopsutil` - System utilities
- `go.etcd.io/bbolt` - Embedded database (BoltDB)

## 📝 License

This project is proprietary software licensed under the **Mervlot Proprietary Evaluation License (MPEL) v1.0**.

Copyright (c) 2025 Mervlot (Ogunmuyiwa Muhammad). All rights reserved.

The software is proprietary and confidential. Usage is permitted solely for internal evaluation, testing, and development with explicit written permission from Mervlot.

See [LICENCE.md](LICENCE.md) for full license terms.

## 👤 Author

**Mervlot (Ogunmuyiwa Muhammad)**

---

## 🤝 Contributing

This is a proprietary project. Contributions require explicit permission from the copyright holder.

## ⚠️ Status

**Current Version**: 0.0.0

This is an active development project. Features and APIs may change significantly between versions.

---

**For questions or issues, please contact the project owner.**
