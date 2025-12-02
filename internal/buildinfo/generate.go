package buildinfo

//go:generate sh -c "echo \"version=$(git describe --tags --always --dirty 2>/dev/null || echo unknown)\" > .version; echo \"commit=$(git rev-parse HEAD 2>/dev/null || echo unknown)\" >> .version; echo \"date=$(date -u +%Y-%m-%dT%H:%M:%SZ)\" >> .version"
