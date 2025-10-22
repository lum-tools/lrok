# lrok v1.0.0 Release Checklist

## Pre-Release Setup

### 1. GitHub Repository ✅
- [x] Repository created: `lum-tools/lrok`
- [ ] Add repository description: "Expose local services with readable tunnel names"
- [ ] Add topics: `tunnel`, `ngrok`, `frp`, `reverse-proxy`, `https`
- [ ] Enable Issues
- [ ] Enable Discussions (optional)

### 2. npm Organization
- [ ] Verify npm account access
- [ ] Join/create `@lum-tools` organization (or use personal account)
- [ ] Generate npm token: https://www.npmjs.com/settings/tokens
  - Token type: Automation
  - Scope: Publish
- [ ] Add to GitHub Secrets as `NPM_TOKEN`

### 4. PyPI Account
- [ ] Create PyPI account: https://pypi.org/account/register/
- [ ] Generate API token: https://pypi.org/manage/account/token/
  - Scope: Entire account (or specific to lrok after first upload)
- [ ] Add to GitHub Secrets as `PYPI_TOKEN`

### 5. GitHub Secrets Configuration
Navigate to: https://github.com/lum-tools/lrok/settings/secrets/actions

Add the following secrets:
- [ ] `NPM_TOKEN` - npm automation token
- [ ] `PYPI_TOKEN` - PyPI API token
- [ ] `GITHUB_TOKEN` (automatically provided, verify workflows have access)

### 6. Local Testing

```bash
cd /home/ethan/Work/lum.tools/cli/lrok

# Test build
go build -o lrok cmd/lrok/main.go

# Test binary
./lrok version
./lrok --help
export LUM_API_KEY=lum_test_key
./lrok 8000 --name test-tunnel  # Should connect and fail auth (expected)

# Test GoReleaser locally (dry run)
goreleaser release --snapshot --clean --skip-publish
```

- [ ] All tests pass
- [ ] Binary works on current platform
- [ ] Help text looks good
- [ ] Version command works

## Release Process

### 7. Prepare Repository

```bash
cd /home/ethan/Work/lum.tools/cli/lrok

# Initialize git if not done
git init
git branch -M main

# Add all files
git add .

# Create .gitignore to exclude binaries
cat > .gitignore << 'EOF'
# Binaries
lrok
*.exe
*.dll
*.so
*.dylib
dist/

# Test
*.test
*.out

# Go
go.sum

# IDE
.vscode/
.idea/
*.swp

# OS
.DS_Store

# Temp
*.log

# Don't ignore embedded binaries
!internal/embed/bins/frpc_*
EOF

git add .gitignore

# Commit
git commit -m "lrok v1.0.0 - Cross-platform tunnel CLI

Features:
- Simple CLI: lrok 8000
- Random readable names (8100+ combinations)
- Embedded frpc binaries (all platforms)
- Multi-registry distribution (Homebrew, npm, PyPI)
- API key authentication
- Traffic tracking integration
"

# Add remote
git remote add origin https://github.com/lum-tools/lrok.git

# Push
git push -u origin main
```

- [ ] Repository pushed to GitHub
- [ ] All files visible in GitHub

### 8. Create Release

```bash
# Tag the release
git tag -a v1.0.0 -m "Release v1.0.0

First stable release of lrok CLI.

Installation:
- Homebrew: brew tap lum-tools/tap && brew install lrok
- npm: npm install -g lrok
- PyPI: pip install lrok
- Direct: Download from GitHub releases

Usage:
  export LUM_API_KEY='lum_your_key'
  lrok 8000
  lrok 8000 --name my-app
"

# Push tag (this triggers GitHub Actions)
git push origin v1.0.0
```

- [ ] Tag created
- [ ] Tag pushed to GitHub

### 9. Monitor GitHub Actions

Visit: https://github.com/lum-tools/lrok/actions

**Test Workflow** (runs on push to main):
- [ ] Build succeeds on Ubuntu
- [ ] Build succeeds on macOS
- [ ] Build succeeds on Windows
- [ ] All tests pass

**Release Workflow** (runs on tag push):
- [ ] frpc binaries downloaded
- [ ] GoReleaser builds all platforms
- [ ] GitHub Release created
- [ ] Binaries uploaded to release
- [ ] npm package published
- [ ] PyPI package published

### 10. Verify Release Artifacts

**GitHub Release:**
- [ ] Visit: https://github.com/lum-tools/lrok/releases/tag/v1.0.0
- [ ] Binaries present for all platforms:
  - [ ] `lrok_1.0.0_darwin_amd64.tar.gz`
  - [ ] `lrok_1.0.0_darwin_arm64.tar.gz`
  - [ ] `lrok_1.0.0_linux_amd64.tar.gz`
  - [ ] `lrok_1.0.0_linux_arm64.tar.gz`
  - [ ] `lrok_1.0.0_windows_amd64.zip`
  - [ ] `checksums.txt` present
  - [ ] Release notes look good

**npm:**
- [ ] Visit: https://www.npmjs.com/package/lrok
- [ ] Version 1.0.0 published
- [ ] README displays correctly
- [ ] Install count starts tracking

**PyPI:**
- [ ] Visit: https://pypi.org/project/lrok/
- [ ] Version 1.0.0 published
- [ ] README displays correctly
- [ ] Download stats visible

## Post-Release Testing

### 11. Test Installations

**npm (Cross-platform):**
```bash
npm install -g lrok
lrok version
# Should show: lrok version 1.0.0
```
- [ ] Installation works
- [ ] Binary downloads correctly for platform
- [ ] Command available in PATH

**PyPI (Cross-platform):**
```bash
pip install lrok
lrok version
# Should show: lrok version 1.0.0
```
- [ ] Installation works
- [ ] Binary downloads correctly for platform
- [ ] Command available in PATH

**Direct Download:**
```bash
curl -L https://github.com/lum-tools/lrok/releases/download/v1.0.0/lrok_1.0.0_linux_amd64.tar.gz | tar xz
./lrok version
```
- [ ] Download works
- [ ] Binary is executable
- [ ] Runs correctly

### 12. End-to-End Testing

```bash
export LUM_API_KEY='lum_your_real_key'

# Test 1: Simple usage
lrok 8000

# Test 2: Custom name
lrok 8000 --name my-test-app

# Test 3: HTTP command
lrok http 3000

# Verify at https://platform.lum.tools/tunnels
```

- [ ] Tunnel connects successfully
- [ ] URL is accessible
- [ ] Platform dashboard shows tunnel
- [ ] Traffic is tracked
- [ ] Graceful shutdown works (Ctrl+C)

## Documentation Updates

### 13. Update Platform Documentation

- [ ] Update `/home/ethan/Work/lum.tools/services/platform/app/templates/tunnels.html`
  - Change "Download Client" link to point to lrok releases
  - Update code examples to use `lrok 8000`
  
- [ ] Update `/home/ethan/Work/lum.tools/services/platform/app/templates/index.html`
  - Update Quick Start example to use lrok

- [ ] Create announcement for platform users

### 14. Create Announcement

**Blog Post / Announcement:**
```markdown
# Introducing lrok - The Easy Tunnel CLI

We're excited to announce lrok v1.0.0!

## What is lrok?

lrok is a simple CLI tool to expose your local services to the internet with readable URLs.

## Installation

Choose your preferred method:

```bash
# Homebrew (macOS/Linux)
brew tap lum-tools/tap
brew install lrok

# npm (Cross-platform)
npm install -g lrok

# PyPI (Cross-platform)
pip install lrok
```

## Quick Start

```bash
export LUM_API_KEY='your_key'
lrok 8000
# → https://happy-dolphin.t.lum.tools
```

## Features

- Simple CLI (just `lrok 8000`)
- Readable URLs (happy-dolphin, brave-tiger, etc.)
- Zero configuration
- Cross-platform (macOS, Linux, Windows)
- Built-in HTTPS
- Traffic tracking

Get started: https://github.com/lum-tools/lrok
```

- [ ] Announcement posted
- [ ] Shared on social media (if applicable)

## Troubleshooting

### Common Issues:

**npm install fails:**
- Check Node.js version (>= 14)
- Check npm registry access
- Try: `npm install -g lrok --verbose`

**PyPI install fails:**
- Check Python version (>= 3.7)
- Check pip version
- Try: `pip install lrok --verbose`

**Homebrew install fails:**
- Run: `brew update`
- Check tap is added: `brew tap lum-tools/tap`
- Try: `brew install lrok --verbose`

**Binary doesn't run:**
- Check platform/arch compatibility
- Verify executable permissions: `chmod +x lrok`
- Check PATH includes install location

## Success Criteria

All checkboxes above are complete:
- [ ] Repository setup ✓
- [ ] Credentials configured ✓
- [ ] Release created ✓
- [ ] GitHub Actions successful ✓
- [ ] All package managers working ✓
- [ ] End-to-end testing passed ✓
- [ ] Documentation updated ✓

🎉 **lrok v1.0.0 is live!**

