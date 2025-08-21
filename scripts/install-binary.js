#!/usr/bin/env node

const fs = require('fs');
const path = require('path');
const https = require('https');
const { execSync } = require('child_process');

// 🎯 Detect platform and architecture
function getPlatform() {
  const platform = process.platform;
  const arch = process.arch;
  
  let goos, goarch;
  
  switch (platform) {
    case 'darwin':
      goos = 'darwin';
      break;
    case 'linux':
      goos = 'linux';
      break;
    case 'win32':
      goos = 'windows';
      break;
    default:
      throw new Error(`Unsupported platform: ${platform}`);
  }
  
  switch (arch) {
    case 'x64':
      goarch = 'amd64';
      break;
    case 'arm64':
      goarch = 'arm64';
      break;
    default:
      throw new Error(`Unsupported architecture: ${arch}`);
  }
  
  return { goos, goarch };
}

// 🚀 Download and install binary
async function installBinary() {
  console.log('🚀 Installing glab-tui binary...');
  
  try {
    const { goos, goarch } = getPlatform();
    const binDir = path.join(__dirname, '..', 'bin');
    const binaryName = goos === 'windows' ? 'glab-tui.exe' : 'glab-tui';
    const binaryPath = path.join(binDir, binaryName);
    
    // Create bin directory
    if (!fs.existsSync(binDir)) {
      fs.mkdirSync(binDir, { recursive: true });
    }
    
    // For now, build from source (later we'll download from releases)
    console.log('🏗️  Building glab-tui from source...');
    
    // Check if Go is available
    try {
      execSync('go version', { stdio: 'ignore' });
    } catch (error) {
      console.error('❌ Go is required to build glab-tui');
      console.error('💡 Install Go from https://golang.org/dl/');
      process.exit(1);
    }
    
    // Build the binary
    const buildEnv = {
      ...process.env,
      GOOS: goos,
      GOARCH: goarch,
      CGO_ENABLED: '0'
    };
    
    execSync(`go build -ldflags="-s -w" -o "${binaryPath}" .`, {
      cwd: path.join(__dirname, '..'),
      env: buildEnv,
      stdio: 'inherit'
    });
    
    // Make executable on Unix systems
    if (goos !== 'windows') {
      fs.chmodSync(binaryPath, '755');
    }
    
    console.log('✅ glab-tui installed successfully!');
    console.log('🎯 Try: npx glab-tui help');
    
  } catch (error) {
    console.error('❌ Failed to install glab-tui:', error.message);
    process.exit(1);
  }
}

// 🎯 Run installation
if (require.main === module) {
  installBinary();
}

module.exports = { installBinary, getPlatform };
