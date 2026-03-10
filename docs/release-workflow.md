# Release Workflow

## 1. 语义化版本规范

### 1.1 版本格式

```
MAJOR.MINOR.PATCH[-PRERELEASE]

示例:
- 0.1.0-alpha    # 早期开发版本
- 0.5.0-beta     # 功能基本完成
- 1.0.0-rc1      # 发布候选
- 1.0.0          # 正式发布
- 1.1.0          # 新功能
- 1.0.1          # Bug 修复
```

### 1.2 版本规则

| 变更类型 | 版本增量 | 示例 |
|----------|----------|------|
| 不兼容的 API 变更 | MAJOR | 1.x.x → 2.0.0 |
| 向后兼容的新功能 | MINOR | 1.0.x → 1.1.0 |
| 向后兼容的 Bug 修复 | PATCH | 1.0.0 → 1.0.1 |
| 预发布版本 | PRERELEASE | 1.0.0-alpha.1 |

### 1.3 分支策略

```
main (stable)
  │
  ├── develop (integration)
  │     │
  │     ├── feature/phase-0-foundation
  │     ├── feature/phase-1-worktree
  │     ├── feature/phase-2-terminal
  │     └── ...
  │
  └── release/v1.0.0
```

---

## 2. GitHub Actions CI/CD

### 2.1 完整工作流

```yaml
# .github/workflows/release.yml
name: Build and Release

on:
  push:
    tags:
      - 'v*'
  workflow_dispatch:
    inputs:
      version:
        description: 'Version to release (e.g., v1.0.0)'
        required: true

env:
  GO_VERSION: '1.21'

jobs:
  build:
    strategy:
      matrix:
        include:
          - os: macos-latest
            platform: darwin
            arch: amd64
            ext: ''
          - os: macos-latest
            platform: darwin
            arch: arm64
            ext: ''
          - os: ubuntu-latest
            platform: linux
            arch: amd64
            ext: ''
          - os: windows-latest
            platform: windows
            arch: amd64
            ext: '.exe'
    
    runs-on: ${{ matrix.os }}
    
    steps:
      - uses: actions/checkout@v4
      
      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: ${{ env.GO_VERSION }}
      
      - name: Set up Node
        uses: actions/setup-node@v4
        with:
          node-version: '20'
          cache: 'npm'
          cache-dependency-path: frontend/package-lock.json
      
      - name: Install Wails
        run: go install github.com/wailsapp/wails/v2/cmd/wails@latest
      
      - name: Install frontend dependencies
        working-directory: frontend
        run: npm ci
      
      - name: Build
        run: wails build -platform ${{ matrix.platform }}/${{ matrix.arch }}
      
      - name: Package (macOS)
        if: matrix.platform == 'darwin'
        run: |
          cd build/bin
          zip -r agent-orch-${{ matrix.platform }}-${{ matrix.arch }}.zip "Agent Orchestrator.app"
      
      - name: Package (Linux)
        if: matrix.platform == 'linux'
        run: |
          cd build/bin
          tar -czvf agent-orch-${{ matrix.platform }}-${{ matrix.arch }}.tar.gz agent-orch
      
      - name: Package (Windows)
        if: matrix.platform == 'windows'
        run: |
          cd build/bin
          7z a agent-orch-${{ matrix.platform }}-${{ matrix.arch }}.zip agent-orch.exe
      
      - name: Upload artifact
        uses: actions/upload-artifact@v4
        with:
          name: agent-orch-${{ matrix.platform }}-${{ matrix.arch }}
          path: build/bin/agent-orch-${{ matrix.platform }}-${{ matrix.arch }}.*

  release:
    needs: build
    runs-on: ubuntu-latest
    
    steps:
      - uses: actions/checkout@v4
      
      - name: Download all artifacts
        uses: actions/download-artifact@v4
        with:
          path: artifacts
      
      - name: Get version
        id: version
        run: |
          if [ "${{ github.event_name }}" = "workflow_dispatch" ]; then
            echo "VERSION=${{ github.event.inputs.version }}" >> $GITHUB_OUTPUT
          else
            echo "VERSION=${GITHUB_REF#refs/tags/}" >> $GITHUB_OUTPUT
          fi
      
      - name: Generate changelog
        id: changelog
        run: |
          # 提取 CHANGELOG.md 中当前版本的内容
          VERSION=${{ steps.version.outputs.VERSION }}
          sed -n "/## ${VERSION#v}/,/## /p" CHANGELOG.md | head -n -1 > release_notes.md
          
          # 如果没有找到，使用最后提交
          if [ ! -s release_notes.md ]; then
            echo "Release for $VERSION" > release_notes.md
          fi
      
      - name: Create Release
        uses: softprops/action-gh-release@v1
        with:
          tag_name: ${{ steps.version.outputs.VERSION }}
          name: Agent Orchestrator ${{ steps.version.outputs.VERSION }}
          body_path: release_notes.md
          files: |
            artifacts/**/*
          draft: ${{ contains(steps.version.outputs.VERSION, '-') }}
          prerelease: ${{ contains(steps.version.outputs.VERSION, 'alpha') || contains(steps.version.outputs.VERSION, 'beta') }}
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

### 2.2 测试工作流

```yaml
# .github/workflows/ci.yml
name: CI

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]

jobs:
  test:
    runs-on: ${{ matrix.os }}
    strategy:
      matrix:
        os: [macos-latest, ubuntu-latest, windows-latest]
    
    steps:
      - uses: actions/checkout@v4
      
      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.21'
      
      - name: Set up Node
        uses: actions/setup-node@v4
        with:
          node-version: '20'
      
      - name: Install dependencies
        run: |
          go mod download
          cd frontend && npm ci
      
      - name: Run Go tests
        run: go test ./... -v -race -coverprofile=coverage.out
      
      - name: Run frontend tests
        working-directory: frontend
        run: npm test
      
      - name: Lint Go
        uses: golangci/golangci-lint-action@v3
      
      - name: Lint frontend
        working-directory: frontend
        run: npm run lint
      
      - name: Type check
        working-directory: frontend
        run: npm run typecheck
      
      - name: Build check
        run: |
          go install github.com/wailsapp/wails/v2/cmd/wails@latest
          wails build

  coverage:
    runs-on: ubuntu-latest
    needs: test
    
    steps:
      - uses: actions/checkout@v4
      
      - name: Download coverage
        uses: actions/download-artifact@v4
        with:
          name: coverage
      
      - name: Upload to Codecov
        uses: codecov/codecov-action@v3
        with:
          files: coverage.out
```

---

## 3. 发布检查清单

### 3.1 代码质量

- [ ] 所有测试通过
- [ ] 代码覆盖率达标
- [ ] golangci-lint 通过
- [ ] ESLint 通过
- [ ] TypeScript 检查通过

### 3.2 文档更新

- [ ] CHANGELOG.md 已更新
- [ ] README.md 已更新（如有需要）
- [ ] 迁移指南已编写（如有破坏性变更）

### 3.3 版本信息

- [ ] 版本号符合语义化规范
- [ ] Git tag 已创建
- [ ] Release notes 已准备

### 3.4 构建验证

- [ ] macOS 构建成功
- [ ] Linux 构建成功
- [ ] Windows 构建成功
- [ ] 安装包测试通过

---

## 4. CHANGELOG 格式

```markdown
# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- ...

### Changed
- ...

### Fixed
- ...

## [1.0.0] - 2024-03-10

### Added
- Git worktree management (create, list, delete)
- Embedded terminal with tmux support
- API proxy for LLM API management
- Usage tracking and cost estimation
- Agent process detection
- GitHub PR integration
- Diff viewer
- Code editor with syntax highlighting

### Changed
- Initial stable release

### Fixed
- Various bug fixes

## [0.5.0-beta] - 2024-02-15

### Added
- API Manager feature (core differentiator)
- Profile management
- API Key encryption

## [0.4.0-alpha] - 2024-02-01

### Added
- Terminal stability improvements
- TUI application compatibility
- Session persistence

## [0.3.0-alpha] - 2024-01-15

### Added
- Basic terminal functionality
- tmux integration
- xterm.js integration

## [0.2.0-alpha] - 2024-01-01

### Added
- Git worktree CRUD operations
- Branch status indicators

## [0.1.0-alpha] - 2023-12-15

### Added
- Initial project scaffold
- Basic three-column layout
- Configuration system

[Unreleased]: https://github.com/user/agent-orch/compare/v1.0.0...HEAD
[1.0.0]: https://github.com/user/agent-orch/compare/v0.5.0-beta...v1.0.0
[0.5.0-beta]: https://github.com/user/agent-orch/compare/v0.4.0-alpha...v0.5.0-beta
[0.4.0-alpha]: https://github.com/user/agent-orch/compare/v0.3.0-alpha...v0.4.0-alpha
[0.3.0-alpha]: https://github.com/user/agent-orch/compare/v0.2.0-alpha...v0.3.0-alpha
[0.2.0-alpha]: https://github.com/user/agent-orch/compare/v0.1.0-alpha...v0.2.0-alpha
[0.1.0-alpha]: https://github.com/user/agent-orch/releases/tag/v0.1.0-alpha
```

---

## 5. 分发渠道

### 5.1 GitHub Releases

主要分发渠道，所有版本都在这里发布。

### 5.2 Homebrew (macOS)

```ruby
# Formula/agent-orch.rb
class AgentOrch < Formula
  desc "AI Agent Worktree Manager"
  homepage "https://github.com/user/agent-orch"
  version "1.0.0"
  license "MIT"
  
  on_macos do
    on_intel do
      url "https://github.com/user/agent-orch/releases/download/v#{version}/agent-orch-darwin-amd64.zip"
      sha256 "..."
    end
    on_arm do
      url "https://github.com/user/agent-orch/releases/download/v#{version}/agent-orch-darwin-arm64.zip"
      sha256 "..."
    end
  end
  
  def install
    bin.install "agent-orch"
  end
end
```

安装命令：
```bash
brew tap user/agent-orch
brew install agent-orch
```

### 5.3 Scoop (Windows)

```json
// bucket/agent-orch.json
{
  "version": "1.0.0",
  "description": "AI Agent Worktree Manager",
  "homepage": "https://github.com/user/agent-orch",
  "license": "MIT",
  "architecture": {
    "64bit": {
      "url": "https://github.com/user/agent-orch/releases/download/v1.0.0/agent-orch-windows-amd64.zip",
      "hash": "..."
    }
  },
  "bin": "agent-orch.exe",
  "shortcuts": [
    ["agent-orch.exe", "Agent Orchestrator"]
  ]
}
```

安装命令：
```powershell
scoop bucket add agent-orch https://github.com/user/scoop-bucket
scoop install agent-orch
```

### 5.4 AUR (Arch Linux)

```bash
# yay -S agent-orch-bin
```

---

## 6. 发布流程

### 6.1 自动发布（推荐）

```bash
# 1. 更新 CHANGELOG.md
# 2. 创建 tag
git tag v1.0.0
git push origin v1.0.0

# 3. GitHub Actions 自动构建和发布
```

### 6.2 手动发布

```bash
# 1. 确保 main 分支是最新的
git checkout main
git pull

# 2. 运行测试
go test ./...
cd frontend && npm test

# 3. 构建
wails build -platform darwin/amd64,darwin/arm64,linux/amd64,windows/amd64

# 4. 打包
cd build/bin
# macOS
zip -r agent-orch-darwin-amd64.zip "Agent Orchestrator.app"
# Linux
tar -czvf agent-orch-linux-amd64.tar.gz agent-orch
# Windows
7z a agent-orch-windows-amd64.zip agent-orch.exe

# 5. 创建 GitHub Release
gh release create v1.0.0 \
  --title "Agent Orchestrator v1.0.0" \
  --notes-file release_notes.md \
  agent-orch-*.zip agent-orch-*.tar.gz
```

---

## 7. 版本规划

| Phase | 版本 | 预计时间 | 里程碑 |
|-------|------|----------|--------|
| 0 | v0.1.0-alpha | Week 1 | 项目脚手架 |
| 1 | v0.2.0-alpha | Week 3 | Worktree 管理 |
| 2 | v0.3.0-alpha | Week 5 | 终端 MVP |
| 3 | v0.4.0-alpha | Week 7 | 终端稳定性 |
| 4 | v0.5.0-beta | Week 11 | API 管理中心 |
| 5 | v1.0.0-rc1 | Week 13 | Agent 监控 |
| 5 | v1.0.0 | Week 14 | 正式发布 |
| 6 | v1.1.0 | Week 18 | 代码编辑器 |

---

## 8. 发布后检查

- [ ] GitHub Release 页面正确显示
- [ ] 下载链接可用
- [ ] macOS 应用可正常打开
- [ ] Windows 应用可正常打开
- [ ] Linux 二进制可执行
- [ ] Homebrew formula 更新
- [ ] Scoop manifest 更新
- [ ] 社交媒体公告
