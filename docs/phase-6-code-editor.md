# Phase 6: Code Editor

> **周期**：3-4 周
> **目标**：内置代码编辑器
> **依赖**：Phase 0-5
> **交付**：v1.1.0

---

## 1. Feature List

### 1.1 基础编辑器

| Feature | 描述 | 优先级 |
|---------|------|--------|
| F6.1 | CodeMirror 6 集成 | P0 |
| F6.2 | 文件打开/保存 | P0 |
| F6.3 | 基础编辑功能 | P0 |
| F6.4 | 撤销/重做 | P0 |
| F6.5 | 查找/替换 | P1 |

### 1.2 语法高亮

| Feature | 描述 | 优先级 |
|---------|------|--------|
| F6.6 | Go 语法高亮 | P0 |
| F6.7 | TypeScript/JavaScript | P0 |
| F6.8 | Python | P0 |
| F6.9 | Rust | P0 |
| F6.10 | Markdown | P0 |
| F6.11 | JSON/YAML/TOML | P1 |

### 1.3 Markdown 支持

| Feature | 描述 | 优先级 |
|---------|------|--------|
| F6.12 | 实时预览 | P0 |
| F6.13 | GFM 支持 | P0 |
| F6.14 | 代码块高亮 | P1 |
| F6.15 | 分屏模式 | P1 |

### 1.4 代码导航

| Feature | 描述 | 优先级 |
|---------|------|--------|
| F6.16 | 正则符号匹配 (Layer 1) | P0 |
| F6.17 | 当前文件跳转 | P0 |
| F6.18 | Tree-sitter 索引 (Layer 2) | P2 |
| F6.19 | 跨文件跳转 (Layer 3) | P3 |

---

## 2. 技术选型

### 2.1 为什么不用 LSP

| 问题 | LSP | 我们的方案 |
|------|-----|-----------|
| 启动时间 | 1-3 秒 | <10ms |
| 内存占用 | 100MB+/语言 | ~10MB |
| 用户配置 | 需安装语言服务器 | 开箱即用 |
| 首次打开 | 需等待索引 | 立即可用 |

### 2.2 分层导航方案

```
┌─────────────────────────────────────────────────────────────┐
│                    Code Navigation Layers                   │
├─────────────────────────────────────────────────────────────┤
│ Layer 1: 即时响应（打开即用）<10ms                          │
│  - 正则符号匹配                                             │
│  - 当前文件跳转                                             │
│  - 无需任何初始化                                           │
├─────────────────────────────────────────────────────────────┤
│ Layer 2: 后台索引（异步）1-5秒后可用                        │
│  - Tree-sitter 解析：10-50ms/文件                           │
│  - 后台构建符号索引                                         │
│  - 不阻塞 UI                                                │
├─────────────────────────────────────────────────────────────┤
│ Layer 3: 高级导航（可选）索引完成后可用                     │
│  - Stack Graphs：跨文件引用追踪                             │
│  - GitHub 开源技术                                          │
└─────────────────────────────────────────────────────────────┘
```

---

## 3. 实现细节

### 3.1 Go 后端 - 文件操作

```go
// internal/editor/file.go
package editor

import (
    "io/ioutil"
    "os"
    "path/filepath"
    "strings"
)

type FileInfo struct {
    Path     string `json:"path"`
    Name     string `json:"name"`
    IsDir    bool   `json:"isDir"`
    Size     int64  `json:"size"`
    Modified string `json:"modified"`
}

type FileContent struct {
    Path    string `json:"path"`
    Content string `json:"content"`
    Language string `json:"language"`
}

// ListFiles 列出目录下的文件
func ListFiles(dirPath string) ([]FileInfo, error) {
    entries, err := os.ReadDir(dirPath)
    if err != nil {
        return nil, err
    }
    
    var files []FileInfo
    for _, entry := range entries {
        // 跳过隐藏文件
        if strings.HasPrefix(entry.Name(), ".") {
            continue
        }
        
        info, err := entry.Info()
        if err != nil {
            continue
        }
        
        files = append(files, FileInfo{
            Path:     filepath.Join(dirPath, entry.Name()),
            Name:     entry.Name(),
            IsDir:    entry.IsDir(),
            Size:     info.Size(),
            Modified: info.ModTime().Format("2006-01-02 15:04"),
        })
    }
    
    return files, nil
}

// ReadFile 读取文件内容
func ReadFile(path string) (*FileContent, error) {
    data, err := ioutil.ReadFile(path)
    if err != nil {
        return nil, err
    }
    
    return &FileContent{
        Path:     path,
        Content:  string(data),
        Language: detectLanguage(path),
    }, nil
}

// WriteFile 保存文件
func WriteFile(path, content string) error {
    return ioutil.WriteFile(path, []byte(content), 0644)
}

// detectLanguage 检测文件语言
func detectLanguage(path string) string {
    ext := strings.ToLower(filepath.Ext(path))
    
    langMap := map[string]string{
        ".go":   "go",
        ".ts":   "typescript",
        ".tsx":  "typescript",
        ".js":   "javascript",
        ".jsx":  "javascript",
        ".py":   "python",
        ".rs":   "rust",
        ".md":   "markdown",
        ".json": "json",
        ".yaml": "yaml",
        ".yml":  "yaml",
        ".toml": "toml",
        ".css":  "css",
        ".html": "html",
        ".sh":   "shell",
    }
    
    return langMap[ext]
}
```

### 3.2 Go 后端 - 正则符号匹配

```go
// internal/editor/navigator.go
package editor

import (
    "regexp"
    "strings"
)

type Symbol struct {
    Name     string `json:"name"`
    Type     string `json:"type"`     // function, variable, type, import
    Line     int    `json:"line"`
    Column   int    `json:"column"`
    EndLine  int    `json:"endLine"`
}

type LanguagePatterns struct {
    FunctionDef *regexp.Regexp
    VarDef      *regexp.Regexp
    TypeDef     *regexp.Regexp
    Import      *regexp.Regexp
}

var patterns = map[string]*LanguagePatterns{
    "go": {
        FunctionDef: regexp.MustCompile(`func\s+(?:\([^)]+\)\s*)?(\w+)\s*\(`),
        VarDef:      regexp.MustCompile(`(?:var|const)\s+(\w+)\s*(?:=|,)`),
        TypeDef:     regexp.MustCompile(`type\s+(\w+)\s+(?:struct|interface)`),
        Import:      regexp.MustCompile(`import\s+(?:\(([^)]+)\)|"([^"]+)")`),
    },
    "typescript": {
        FunctionDef: regexp.MustCompile(`(?:function|const|let|var)\s+(\w+)\s*[=\(]`),
        VarDef:      regexp.MustCompile(`(?:const|let|var)\s+(\w+)\s*=`),
        TypeDef:     regexp.MustCompile(`(?:interface|type|class)\s+(\w+)`),
        Import:      regexp.MustCompile(`import\s+.*?from\s+['"]([^'"]+)['"]`),
    },
    "python": {
        FunctionDef: regexp.MustCompile(`def\s+(\w+)\s*\(`),
        VarDef:      regexp.MustCompile(`(\w+)\s*=\s*(?:[^=]|$)`),
        TypeDef:     regexp.MustCompile(`class\s+(\w+)`),
        Import:      regexp.MustCompile(`(?:import|from)\s+(\w+)`),
    },
    "rust": {
        FunctionDef: regexp.MustCompile(`fn\s+(\w+)\s*[<\(]`),
        VarDef:      regexp.MustCompile(`let\s+(?:mut\s+)?(\w+)`),
        TypeDef:     regexp.MustCompile(`(?:struct|enum|trait)\s+(\w+)`),
        Import:      regexp.MustCompile(`use\s+([^;]+)`),
    },
}

// FindSymbols 查找文件中的所有符号
func FindSymbols(content, language string) []Symbol {
    p, ok := patterns[language]
    if !ok {
        return nil
    }
    
    var symbols []Symbol
    lines := strings.Split(content, "\n")
    
    for i, line := range lines {
        // 函数定义
        if p.FunctionDef != nil {
            if matches := p.FunctionDef.FindStringSubmatch(line); len(matches) > 1 {
                symbols = append(symbols, Symbol{
                    Name:   matches[1],
                    Type:   "function",
                    Line:   i + 1,
                    Column: strings.Index(line, matches[1]),
                })
            }
        }
        
        // 变量定义
        if p.VarDef != nil {
            if matches := p.VarDef.FindStringSubmatch(line); len(matches) > 1 {
                symbols = append(symbols, Symbol{
                    Name:   matches[1],
                    Type:   "variable",
                    Line:   i + 1,
                    Column: strings.Index(line, matches[1]),
                })
            }
        }
        
        // 类型定义
        if p.TypeDef != nil {
            if matches := p.TypeDef.FindStringSubmatch(line); len(matches) > 1 {
                symbols = append(symbols, Symbol{
                    Name:   matches[1],
                    Type:   "type",
                    Line:   i + 1,
                    Column: strings.Index(line, matches[1]),
                })
            }
        }
    }
    
    return symbols
}

// FindDefinition 查找符号定义
func FindDefinition(content, language, symbolName string) *Symbol {
    symbols := FindSymbols(content, language)
    
    for _, sym := range symbols {
        if sym.Name == symbolName {
            return &sym
        }
    }
    
    return nil
}
```

### 3.3 前端 - 编辑器组件

```tsx
// frontend/src/components/Editor/CodeEditor.tsx
import { useEffect, useRef, useState } from 'react'
import { EditorState } from '@codemirror/state'
import { EditorView, keymap, lineNumbers, highlightActiveLine } from '@codemirror/view'
import { defaultKeymap, history, historyKeymap } from '@codemirror/commands'
import { syntaxHighlighting, defaultHighlightStyle } from '@codemirror/language'
import { go } from '@codemirror/lang-go'
import { javascript } from '@codemirror/lang-javascript'
import { python } from '@codemirror/lang-python'
import { rust } from '@codemirror/lang-rust'
import { markdown } from '@codemirror/lang-markdown'
import { json } from '@codemirror/lang-json'
import { yaml } from '@codemirror/lang-yaml'
import { oneDark } from '@codemirror/theme-one-dark'

import { ReadFile, WriteFile } from '../../wailsjs/go/main/App'

interface CodeEditorProps {
  filePath: string
  onSave?: () => void
}

export default function CodeEditor({ filePath, onSave }: CodeEditorProps) {
  const containerRef = useRef<HTMLDivElement>(null)
  const viewRef = useRef<EditorView | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  
  // 获取语言支持
  const getLanguageExtension = (lang: string) => {
    switch (lang) {
      case 'go':
        return go()
      case 'typescript':
      case 'javascript':
        return javascript({ typescript: lang === 'typescript' })
      case 'python':
        return python()
      case 'rust':
        return rust()
      case 'markdown':
        return markdown()
      case 'json':
        return json()
      case 'yaml':
        return yaml()
      default:
        return []
    }
  }
  
  useEffect(() => {
    if (!containerRef.current || !filePath) return
    
    loadFile()
    
    return () => {
      if (viewRef.current) {
        viewRef.current.destroy()
      }
    }
  }, [filePath])
  
  const loadFile = async () => {
    setLoading(true)
    setError(null)
    
    try {
      const result = await ReadFile(filePath)
      
      // 创建编辑器状态
      const state = EditorState.create({
        doc: result.content,
        extensions: [
          lineNumbers(),
          highlightActiveLine(),
          history(),
          keymap.of([
            ...defaultKeymap,
            ...historyKeymap,
            // Ctrl+S 保存
            {
              key: 'Mod-s',
              run: (view) => {
                saveFile(view.state.doc.toString())
                return true
              },
            },
          ]),
          getLanguageExtension(result.language),
          syntaxHighlighting(defaultHighlightStyle),
          oneDark,
          EditorView.theme({
            '&': { height: '100%' },
            '.cm-scroller': { overflow: 'auto' },
          }),
          EditorView.updateListener.of((update) => {
            if (update.docChanged) {
              // 可以在这里实现自动保存
            }
          }),
        ],
      })
      
      // 销毁旧视图
      if (viewRef.current) {
        viewRef.current.destroy()
      }
      
      // 创建新视图
      const view = new EditorView({
        state,
        parent: containerRef.current,
      })
      
      viewRef.current = view
      setLoading(false)
      
    } catch (err) {
      setError(String(err))
      setLoading(false)
    }
  }
  
  const saveFile = async (content: string) => {
    try {
      await WriteFile(filePath, content)
      onSave?.()
    } catch (err) {
      console.error('Failed to save file:', err)
    }
  }
  
  if (loading) {
    return <div className="flex items-center justify-center h-full text-gray-500">Loading...</div>
  }
  
  if (error) {
    return <div className="p-4 text-red-500">{error}</div>
  }
  
  return (
    <div 
      ref={containerRef} 
      className="h-full w-full overflow-hidden"
    />
  )
}
```

### 3.4 前端 - Markdown 预览

```tsx
// frontend/src/components/Editor/MarkdownEditor.tsx
import { useState, useEffect } from 'react'
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter'
import { oneDark as prismOneDark } from 'react-syntax-highlighter/dist/esm/styles/prism'
import CodeEditor from './CodeEditor'

interface MarkdownEditorProps {
  filePath: string
}

export default function MarkdownEditor({ filePath }: MarkdownEditorProps) {
  const [content, setContent] = useState('')
  const [mode, setMode] = useState<'edit' | 'preview' | 'split'>('split')
  
  useEffect(() => {
    loadContent()
  }, [filePath])
  
  const loadContent = async () => {
    const result = await ReadFile(filePath)
    setContent(result.content)
  }
  
  const handleContentChange = (newContent: string) => {
    setContent(newContent)
  }
  
  return (
    <div className="h-full flex flex-col">
      {/* 工具栏 */}
      <div className="flex items-center gap-2 p-2 border-b border-gray-700">
        <button
          className={`px-3 py-1 rounded ${mode === 'edit' ? 'bg-gray-700' : ''}`}
          onClick={() => setMode('edit')}
        >
          Edit
        </button>
        <button
          className={`px-3 py-1 rounded ${mode === 'preview' ? 'bg-gray-700' : ''}`}
          onClick={() => setMode('preview')}
        >
          Preview
        </button>
        <button
          className={`px-3 py-1 rounded ${mode === 'split' ? 'bg-gray-700' : ''}`}
          onClick={() => setMode('split')}
        >
          Split
        </button>
      </div>
      
      {/* 内容区域 */}
      <div className="flex-1 flex overflow-hidden">
        {/* 编辑器 */}
        {(mode === 'edit' || mode === 'split') && (
          <div className={`${mode === 'split' ? 'w-1/2' : 'w-full'} h-full`}>
            <CodeEditor filePath={filePath} />
          </div>
        )}
        
        {/* 预览 */}
        {(mode === 'preview' || mode === 'split') && (
          <div className={`${mode === 'split' ? 'w-1/2 border-l border-gray-700' : 'w-full'} h-full overflow-auto p-4`}>
            <ReactMarkdown
              remarkPlugins={[remarkGfm]}
              components={{
                code({ node, className, children, ...props }) {
                  const match = /language-(\w+)/.exec(className || '')
                  const isInline = !match
                  
                  if (isInline) {
                    return <code className="bg-gray-800 px-1 rounded" {...props}>{children}</code>
                  }
                  
                  return (
                    <SyntaxHighlighter
                      style={prismOneDark}
                      language={match[1]}
                      PreTag="div"
                      {...props}
                    >
                      {String(children).replace(/\n$/, '')}
                    </SyntaxHighlighter>
                  )
                },
              }}
            >
              {content}
            </ReactMarkdown>
          </div>
        )}
      </div>
    </div>
  )
}
```

### 3.5 前端 - 文件树

```tsx
// frontend/src/components/Editor/FileTree.tsx
import { useState, useEffect } from 'react'
import { FileInfo } from '../../types/editor'
import { ListFiles } from '../../wailsjs/go/main/App'

interface FileTreeProps {
  rootPath: string
  onFileSelect: (path: string) => void
  selectedFile?: string
}

export default function FileTree({ rootPath, onFileSelect, selectedFile }: FileTreeProps) {
  const [files, setFiles] = useState<FileInfo[]>([])
  const [expanded, setExpanded] = useState<Set<string>>(new Set())
  
  useEffect(() => {
    loadFiles(rootPath)
  }, [rootPath])
  
  const loadFiles = async (path: string) => {
    const result = await ListFiles(path)
    setFiles(result || [])
  }
  
  const toggleExpand = (path: string, isDir: boolean) => {
    if (isDir) {
      const newExpanded = new Set(expanded)
      if (newExpanded.has(path)) {
        newExpanded.delete(path)
      } else {
        newExpanded.add(path)
      }
      setExpanded(newExpanded)
    } else {
      onFileSelect(path)
    }
  }
  
  const renderTree = (items: FileInfo[], depth: number = 0) => {
    return items.map(item => (
      <div key={item.path}>
        <div
          className={`
            flex items-center gap-2 px-2 py-1 cursor-pointer hover:bg-gray-700
            ${selectedFile === item.path ? 'bg-gray-700' : ''}
          `}
          style={{ paddingLeft: `${depth * 16 + 8}px` }}
          onClick={() => toggleExpand(item.path, item.isDir)}
        >
          {/* 图标 */}
          {item.isDir ? (
            expanded.has(item.path) ? (
              <FolderOpenIcon className="w-4 h-4 text-yellow-500" />
            ) : (
              <FolderIcon className="w-4 h-4 text-yellow-500" />
            )
          ) : (
            <FileIcon className="w-4 h-4 text-gray-500" />
          )}
          
          {/* 名称 */}
          <span className="truncate text-sm">{item.name}</span>
        </div>
        
        {/* 子目录 */}
        {item.isDir && expanded.has(item.path) && (
          <SubTree path={item.path} depth={depth + 1} onSelect={onFileSelect} selectedFile={selectedFile} />
        )}
      </div>
    ))
  }
  
  return (
    <div className="h-full overflow-auto">
      <div className="p-2 border-b border-gray-700 text-sm text-gray-400 uppercase">
        Files
      </div>
      {renderTree(files)}
    </div>
  )
}

// 子树组件（延迟加载）
function SubTree({ path, depth, onSelect, selectedFile }: { 
  path: string; 
  depth: number; 
  onSelect: (path: string) => void;
  selectedFile?: string;
}) {
  const [files, setFiles] = useState<FileInfo[]>([])
  
  useEffect(() => {
    ListFiles(path).then(setFiles)
  }, [path])
  
  if (files.length === 0) return null
  
  return (
    <>
      {files.map(item => (
        <div key={item.path}>
          <div
            className={`
              flex items-center gap-2 px-2 py-1 cursor-pointer hover:bg-gray-700
              ${selectedFile === item.path ? 'bg-gray-700' : ''}
            `}
            style={{ paddingLeft: `${depth * 16 + 8}px` }}
            onClick={() => {
              if (item.isDir) {
                // 目录点击逻辑在外层处理
              } else {
                onSelect(item.path)
              }
            }}
          >
            {item.isDir ? (
              <FolderIcon className="w-4 h-4 text-yellow-500" />
            ) : (
              <FileIcon className="w-4 h-4 text-gray-500" />
            )}
            <span className="truncate text-sm">{item.name}</span>
          </div>
        </div>
      ))}
    </>
  )
}
```

---

## 4. 测试计划

### 4.1 单元测试

```go
// internal/editor/navigator_test.go
func TestFindSymbols_Go(t *testing.T) {
    content := `package main

func hello() string {
    return "hello"
}

type Person struct {
    Name string
}

var globalVar = 1
`
    
    symbols := FindSymbols(content, "go")
    
    assert.Len(t, symbols, 3)
    assert.Equal(t, "hello", symbols[0].Name)
    assert.Equal(t, "function", symbols[0].Type)
    assert.Equal(t, "Person", symbols[1].Name)
    assert.Equal(t, "type", symbols[1].Type)
    assert.Equal(t, "globalVar", symbols[2].Name)
    assert.Equal(t, "variable", symbols[2].Type)
}

func TestFindDefinition(t *testing.T) {
    content := `package main
func target() {}
func other() {}
`
    
    sym := FindDefinition(content, "go", "target")
    require.NotNil(t, sym)
    assert.Equal(t, 2, sym.Line)
}
```

### 4.2 性能测试

```go
func BenchmarkFindSymbols(b *testing.B) {
    // 生成 1000 行代码
    content := generateTestCode(1000)
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        FindSymbols(content, "go")
    }
}

// 预期: < 5ms
```

### 4.3 手动测试清单

- [ ] 打开 .go 文件，语法高亮正确
- [ ] 打开 .ts 文件，语法高亮正确
- [ ] 打开 .md 文件，预览正确
- [ ] 编辑文件，Ctrl+S 保存成功
- [ ] 打开大文件 (1MB+)，性能可接受
- [ ] 符号跳转功能正常

---

## 5. 验收标准

| 标准 | 描述 |
|------|------|
| 语法高亮 | 支持 Go/TS/Python/Rust/Markdown |
| 文件操作 | 打开/保存正常 |
| Markdown 预览 | 实时预览正确 |
| 性能 | 打开 1MB 文件 < 100ms |
| 符号跳转 | 正则匹配 < 5ms |

---

## 6. 发布检查清单

- [ ] 所有测试通过
- [ ] 性能测试通过
- [ ] 手动测试完成
- [ ] 更新 CHANGELOG.md
- [ ] 创建 Git tag: v1.1.0
- [ ] GitHub Release 构建

---

## 7. 依赖

### Go 依赖

无新增

### 前端依赖

```json
{
  "dependencies": {
    "@codemirror/state": "^6.4",
    "@codemirror/view": "^6.26",
    "@codemirror/commands": "^6.6",
    "@codemirror/language": "^6.10",
    "@codemirror/lang-go": "^6.0",
    "@codemirror/lang-javascript": "^6.2",
    "@codemirror/lang-python": "^6.1",
    "@codemirror/lang-rust": "^6.0",
    "@codemirror/lang-markdown": "^6.2",
    "@codemirror/lang-json": "^6.0",
    "@codemirror/lang-yaml": "^6.0",
    "@codemirror/theme-one-dark": "^6.1",
    "react-markdown": "^9.0",
    "remark-gfm": "^4.0",
    "react-syntax-highlighter": "^15.5"
  }
}
```

---

## 8. 时间估算

| 任务 | 时间 |
|------|------|
| CodeMirror 集成 | 3 天 |
| 文件操作后端 | 1 天 |
| 语法高亮配置 | 1 天 |
| Markdown 预览 | 2 天 |
| 符号导航 | 2 天 |
| 文件树组件 | 2 天 |
| 测试 | 2 天 |
| 优化和 Bug 修复 | 2 天 |
| 文档和发布 | 1 天 |
| **总计** | **16 天 (3-4 周)** |
