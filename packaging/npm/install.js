#!/usr/bin/env node

const https = require('https');
const fs = require('fs');
const path = require('path');
const { exec } = require('child_process');
const { promisify } = require('util');

const execAsync = promisify(exec);

// Get version from package.json
const packageJson = require('./package.json');
const VERSION = packageJson.version;

// Determine platform and architecture
const PLATFORM_MAP = {
  darwin: 'darwin',
  linux: 'linux',
  win32: 'windows'
};

const ARCH_MAP = {
  x64: 'amd64',
  arm64: 'arm64'
};

const platform = PLATFORM_MAP[process.platform];
const arch = ARCH_MAP[process.arch];

if (!platform || !arch) {
  console.error(`Unsupported platform: ${process.platform}/${process.arch}`);
  process.exit(1);
}

const ext = platform === 'windows' ? '.exe' : '';
const archiveName = `lrok_${VERSION}_${platform}_${arch}.tar.gz`;
const url = `https://github.com/lum-tools/lrok/releases/download/v${VERSION}/${archiveName}`;

console.log(`📦 Installing lrok v${VERSION} for ${platform}/${arch}...`);

// Create bin directory
const binDir = path.join(__dirname, 'bin');
if (!fs.existsSync(binDir)) {
  fs.mkdirSync(binDir, { recursive: true });
}

const binPath = path.join(binDir, `lrok${ext}`);

// Download and extract
function download(url, dest) {
  return new Promise((resolve, reject) => {
    const file = fs.createWriteStream(dest);
    https.get(url, (response) => {
      if (response.statusCode === 302 || response.statusCode === 301) {
        // Follow redirect
        return download(response.headers.location, dest).then(resolve).catch(reject);
      }
      
      if (response.statusCode !== 200) {
        reject(new Error(`Failed to download: ${response.statusCode}`));
        return;
      }
      
      response.pipe(file);
      file.on('finish', () => {
        file.close(resolve);
      });
    }).on('error', (err) => {
      fs.unlink(dest, () => {});
      reject(err);
    });
  });
}

async function install() {
  try {
    const tmpFile = path.join(binDir, archiveName);
    
    console.log(`  → Downloading from GitHub releases...`);
    await download(url, tmpFile);
    
    console.log(`  → Extracting binary...`);
    
    if (platform === 'windows') {
      // For Windows, use PowerShell to extract
      await execAsync(`powershell -command "Expand-Archive -Path '${tmpFile}' -DestinationPath '${binDir}' -Force"`);
      fs.renameSync(path.join(binDir, `lrok${ext}`), binPath);
    } else {
      // For Unix-like systems, use tar
      await execAsync(`tar -xzf "${tmpFile}" -C "${binDir}"`);
    }
    
    // Clean up archive
    fs.unlinkSync(tmpFile);
    
    // Make executable (Unix only)
    if (platform !== 'windows') {
      fs.chmodSync(binPath, 0o755);
    }
    
    console.log(`✅ lrok installed successfully!`);
    console.log(`\nRun: lrok version`);
    
  } catch (error) {
    console.error(`❌ Installation failed: ${error.message}`);
    console.error(`\nTry manual installation from: https://github.com/lum-tools/lrok/releases`);
    process.exit(1);
  }
}

install();

