# GitHub Actions Workflow for Automated Builds and Releases

This document outlines the GitHub Actions workflow design for automated building and publishing of releases for the CAPES Periódicos search tool.

## Workflow File Location

The workflow file should be created at:
```
.github/workflows/release.yml
```

## Workflow Configuration

```yaml
name: Build and Release

on:
  push:
    tags:
      - 'v*'

jobs:
  test:
    name: Run tests
    runs-on: ubuntu-latest
    steps:
      - name: Checkout code
        uses: actions/checkout@v3
      
      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.24.x'
          cache: true
      
      - name: Test
        run: go test -v ./...

  build:
    name: Build binaries
    needs: test
    runs-on: ubuntu-latest
    strategy:
      matrix:
        include:
          - os: windows
            arch: amd64
            extension: .exe
          - os: linux
            arch: amd64
            extension: ''
          - os: darwin
            arch: amd64
            extension: ''
    
    steps:
      - name: Checkout code
        uses: actions/checkout@v3
      
      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.24.x'
          cache: true
      
      - name: Build binary
        run: |
          GOOS=${{ matrix.os }} GOARCH=${{ matrix.arch }} go build -v -o capes-search${{ matrix.extension }} cmd/capes-search/main.go
      
      - name: Rename binary for upload
        run: |
          mv capes-search${{ matrix.extension }} capes-search-${{ matrix.os }}-${{ matrix.arch }}${{ matrix.extension }}
      
      - name: Upload binary
        uses: actions/upload-artifact@v3
        with:
          name: capes-search-${{ matrix.os }}-${{ matrix.arch }}
          path: capes-search-${{ matrix.os }}-${{ matrix.arch }}${{ matrix.extension }}

  release:
    name: Create release
    needs: build
    runs-on: ubuntu-latest
    permissions:
      contents: write
    
    steps:
      - name: Checkout code
        uses: actions/checkout@v3
      
      - name: Download all artifacts
        uses: actions/download-artifact@v3
      
      - name: Create release
        id: create_release
        uses: softprops/action-gh-release@v1
        with:
          name: Release ${{ github.ref_name }}
          draft: false
          prerelease: false
          files: |
            capes-search-windows-amd64/capes-search-windows-amd64.exe
            capes-search-linux-amd64/capes-search-linux-amd64
            capes-search-darwin-amd64/capes-search-darwin-amd64
```

## Workflow Explanation

### Triggers

The workflow will trigger whenever a tag matching the pattern 'v*' is pushed to the repository. For example:
- v1.0.0
- v1.2.3
- v0.1.0-beta

### Jobs

#### 1. Test

Runs the Go tests to ensure everything is working correctly before building the binaries.

#### 2. Build

Sets up a matrix to build the application for multiple platforms:
- Windows (amd64)
- Linux (amd64)
- macOS (amd64)

For each platform, it:
- Builds the binary with the appropriate GOOS and GOARCH
- Renames the binary to include OS and architecture information
- Uploads the binary as an artifact

#### 3. Release

Creates a GitHub release with the binaries as downloadable assets.

## Implementation Instructions

1. Create the `.github/workflows` directory
2. Create the `release.yml` file with the above content
3. Commit and push these changes to the repository
4. To trigger a release, create and push a tag with the 'v' prefix:
   ```
   git tag v1.0.0
   git push origin v1.0.0
   ```

## Usage Notes

- The workflow automatically runs tests before building
- The workflow creates a GitHub release with the appropriate binaries
- The release is named after the tag (e.g., "Release v1.0.0")
- Users can download the binaries directly from the GitHub releases page