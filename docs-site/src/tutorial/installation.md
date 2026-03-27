# Installation

Pick whichever method suits your situation. If you just want to try mxcli without installing anything, start with the [Playground](#playground-zero-install). If you're planning to use mxcli on your own project with an AI coding assistant, skip to the [Dev Container](#dev-container-recommended) section.

## Playground (zero install)

The fastest way to try mxcli. The [mxcli Playground](https://github.com/mendixlabs/mxcli-playground) is a GitHub repository with a pre-configured Mendix project, example scripts, and tutorials. Open it in a Codespace and start using mxcli immediately -- nothing to install on your machine.

[![Open in GitHub Codespaces](https://github.com/codespaces/badge.svg)](https://codespaces.new/mendixlabs/mxcli-playground)

The Codespace comes with mxcli, a JDK, Docker-in-Docker, Claude Code, and a sample Mendix 11.x project ready to explore and modify. It includes:

- **5 example scripts** -- explore, create entities, microflows, pages, and security
- **Step-by-step tutorials** -- from first steps through linting and testing
- **AI tool configs** -- pre-configured for Claude Code, GitHub Copilot, OpenCode, Cursor, Windsurf, Continue.dev, and Aider

Once the Codespace is running:

```bash
./mxcli -p App.mpr -c "SHOW STRUCTURE"          # Explore the project
./mxcli exec scripts/01-explore.mdl -p App.mpr   # Run an example script
./mxcli                                           # Start interactive REPL
```

When you're ready to work on your own Mendix project, use one of the installation methods below.

## Binary download

Pre-built binaries are available for Linux, macOS, and Windows on both amd64 and arm64 architectures.

1. Go to the [GitHub Releases page](https://github.com/mendixlabs/mxcli/releases).
2. Download the archive for your platform (e.g., `mxcli_linux_amd64.tar.gz` or `mxcli_darwin_arm64.tar.gz`).
3. Extract the binary and move it somewhere on your `PATH`:

```bash
# Example for Linux/macOS
tar xzf mxcli_linux_amd64.tar.gz
sudo mv mxcli /usr/local/bin/
```

On Windows, extract the `.zip` and add the folder containing `mxcli.exe` to your system PATH.

## Build from source

Building from source requires **Go 1.24 or later** and **Make**. No C compiler is needed -- mxcli uses a pure-Go SQLite driver.

```bash
git clone https://github.com/mendixlabs/mxcli.git
cd mxcli
make build
```

The binary lands at `./bin/mxcli`. You can copy it to a directory on your PATH or run it directly:

```bash
./bin/mxcli --version
```

## Dev Container (recommended)

The Dev Container approach is the recommended way to work with mxcli, particularly when using AI coding assistants. It creates a sandboxed environment so that agents can only access your project files, not the rest of your system.

Here's how to set it up:

**Step 1:** Install mxcli using one of the methods above (you need it locally to run `init`).

**Step 2:** Run `mxcli init` on your Mendix project:

```bash
mxcli init /path/to/my-mendix-project
```

This creates a `.devcontainer/` folder (along with skill files, agent configs, and other goodies) inside your project directory.

**Step 3:** Open the project folder in VS Code and click **"Reopen in Container"** when prompted (or use the Command Palette: `Dev Containers: Reopen in Container`).

VS Code will build and start the container. This takes a minute or two the first time.

### What's inside the Dev Container

The container comes with everything you need pre-installed:

| Component | What it's for |
|-----------|---------------|
| **mxcli** | The CLI itself, copied into the project |
| **JDK 21** (Adoptium) | Required by MxBuild for project validation |
| **Docker-in-Docker** | Running Mendix apps locally with `mxcli docker run` |
| **Node.js** | Playwright testing support |
| **PostgreSQL client** | Database connectivity for demo data |
| **Claude Code** | AI coding assistant (auto-installed on container creation) |

Once the container is running, mxcli is ready to use -- no further setup needed.

### Specifying AI tools

By default, `mxcli init` configures for Claude Code. You can target other tools too:

```bash
# Cursor only
mxcli init --tool cursor /path/to/my-mendix-project

# Multiple tools
mxcli init --tool claude --tool cursor /path/to/my-mendix-project

# Everything
mxcli init --all-tools /path/to/my-mendix-project
```

Run `mxcli init --list-tools` to see all supported tools.

## Verify your installation

Whichever method you used, confirm that mxcli is working:

```bash
mxcli --version
```

You should see version and build information printed to the terminal. If you get a "command not found" error, double-check that the binary is on your PATH.

You're ready to open a project.
