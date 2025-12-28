# LazyWork - AI-powered Git workflow automation

# Version info from git
version := `git describe --tags --always --dirty 2>/dev/null || echo "dev"`
commit := `git rev-parse --short HEAD 2>/dev/null || echo "none"`
date := `date -u +"%Y-%m-%dT%H:%M:%SZ"`

# Module path
module := "github.com/miltonparedes/lazywork"

# Build flags for version injection
ldflags := "-X " + module + "/cmd.Version=" + version + " -X " + module + "/cmd.Commit=" + commit + " -X " + module + "/cmd.BuildDate=" + date

# Default recipe: build dev version
default: dev

# Development build (fast, shows 'dev' version)
dev:
    go build -o lazywork .

# Production build (with version info from git)
build:
    go build -ldflags "{{ldflags}}" -o lazywork .

# Install to GOPATH/bin
install:
    go install -ldflags "{{ldflags}}" .

# Run the application (dev mode)
run *args:
    go run . {{args}}

# Run all tests
test:
    go test ./...

# Run tests with coverage
test-coverage:
    go test -coverprofile=coverage.out ./...
    go tool cover -html=coverage.out -o coverage.html

# Go bin path
gobin := `go env GOPATH` + "/bin"

# Format code with gofumpt
fmt:
    @test -x {{gobin}}/gofumpt || { echo "Installing gofumpt..."; go install mvdan.cc/gofumpt@latest; }
    {{gobin}}/gofumpt -w .

# Check format without modifying files
fmt-check:
    @test -x {{gobin}}/gofumpt || { echo "Installing gofumpt..."; go install mvdan.cc/gofumpt@latest; }
    @{{gobin}}/gofumpt -l . | grep . && { echo "Files need formatting:"; {{gobin}}/gofumpt -l .; exit 1; } || echo "All files formatted correctly"

# Lint code
lint: fmt
    @test -x {{gobin}}/staticcheck || { echo "Installing staticcheck..."; go install honnef.co/go/tools/cmd/staticcheck@latest; }
    {{gobin}}/staticcheck ./...

# Clean build artifacts
clean:
    rm -f lazywork
    rm -f coverage.out coverage.html
    rm -rf dist/

# Add/update dependencies
deps:
    go get github.com/spf13/cobra@latest
    go get github.com/charmbracelet/huh@latest
    go get github.com/charmbracelet/lipgloss@latest
    go get golang.org/x/term@latest
    go mod tidy

# Goreleaser snapshot (test release without publishing)
snapshot:
    goreleaser release --snapshot --clean

# Full release (requires GITHUB_TOKEN)
release:
    goreleaser release --clean
