# lrok Release Process

## 🚀 Automated Release Flow

The `lrok` CLI uses a fully automated release process that publishes to multiple package registries after successful tests.

### 📋 Release Triggers

Releases are triggered by **Git tags** following semantic versioning:

```bash
# Create a new release
git tag -a v0.1.0 -m "Release v0.1.0: Enhanced tunnel types"
git push origin v0.1.0
```

### 🔄 Release Workflow

When a tag is pushed, the following happens automatically:

1. **GitHub Release Creation** (via GoReleaser)
   - Builds binaries for Linux (AMD64/ARM64), macOS, Windows
   - Creates GitHub release with changelog
   - Uploads `.tar.gz`, `.deb`, `.rpm` packages
   - Generates checksums for verification

2. **npm Publication**
   - Publishes to npm registry as `lrok`
   - Updates package.json version automatically
   - Includes install script for automatic binary download
   - Verifies publication success

3. **PyPI Publication**
   - Publishes to PyPI as `lrok`
   - Updates setup.py version automatically
   - Includes Python wrapper for binary download
   - Verifies publication success

### 📦 Package Registries

#### npm (Node.js)
- **Package**: `lrok`
- **Install**: `npm install -g lrok`
- **Registry**: https://www.npmjs.com/package/lrok
- **Auto-download**: Downloads correct binary for user's platform

#### PyPI (Python)
- **Package**: `lrok`
- **Install**: `pip install lrok`
- **Registry**: https://pypi.org/project/lrok/
- **Auto-download**: Downloads correct binary for user's platform

### 🔧 Required Secrets

The following GitHub secrets must be configured:

- `NPM_TOKEN`: npm automation token with publish permissions
- `PYPI_TOKEN`: PyPI API token with upload permissions

### 📊 Release Status

You can monitor release progress at:
- **GitHub Actions**: https://github.com/lum-tools/lrok/actions
- **Releases**: https://github.com/lum-tools/lrok/releases

### 🎯 Release Checklist

Before creating a release:

- [ ] All tests pass (`go test -v ./...`)
- [ ] Version number follows semantic versioning
- [ ] Changelog updated with new features/fixes
- [ ] Documentation updated if needed
- [ ] Local build works (`go build ./cmd/lrok`)

### 🚨 Troubleshooting

#### Release Fails
1. Check GitHub Actions logs for specific error
2. Verify secrets are correctly configured
3. Check if package already exists (scripts handle this automatically)

#### Package Already Exists
- Scripts automatically detect existing packages and skip publication
- This prevents duplicate publication errors

#### Authentication Issues
- Verify `NPM_TOKEN` has publish permissions
- Verify `PYPI_TOKEN` has upload permissions
- Check token expiration dates

### 📈 Release History

| Version | Date | Features |
|---------|------|----------|
| v0.1.0 | 2025-10-28 | Enhanced tunnel types, UDP removal, CI fixes |
| v0.0.6 | Previous | Basic HTTP tunnels |

### 🔗 Installation Commands

After release, users can install via:

```bash
# npm (Node.js)
npm install -g lrok

# pip (Python)
pip install lrok

# Direct download
curl -L https://github.com/lum-tools/lrok/releases/download/v0.1.0/lrok_0.1.0_linux_amd64.tar.gz | tar -xz
sudo mv lrok /usr/local/bin/
```

### 📝 Release Notes Template

When creating a release, use this template:

```markdown
## What's New in v0.1.0

### ✨ New Features
- Added TCP tunnel support
- Added STCP secure tunnel support  
- Added XTCP P2P tunnel support
- Added visitor command for STCP/XTCP

### 🔧 Improvements
- Enhanced configuration generation
- Improved error handling
- Better documentation

### 🐛 Bug Fixes
- Fixed CI test compatibility
- Resolved cross-platform build issues

### 📦 Installation
```bash
npm install -g lrok
# or
pip install lrok
```

### 🚀 Usage
```bash
lrok tcp 8080 --remote-port 15000
lrok stcp 22 --secret-key my-secret
lrok xtcp 8080 --secret-key p2p-key
```
```

## 🎉 Success Criteria

A successful release means:
- ✅ GitHub release created with binaries
- ✅ npm package published and verified
- ✅ PyPI package published and verified
- ✅ All installation methods work
- ✅ Users can install and use new features
