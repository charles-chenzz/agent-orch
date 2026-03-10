# Phase 4: API Manager ⭐ 核心差异化

> **周期**：3-4 周
> **目标**：LLM API 管理中心
> **依赖**：Phase 0-3
> **交付**：v0.5.0-beta

---

## 1. Feature List

### 1.1 API Profile 管理

| Feature | 描述 | 优先级 |
|---------|------|--------|
| F4.1 | Profile CRUD | P0 |
| F4.2 | 多 Provider 支持 (Anthropic/OpenAI/Gemini) | P0 |
| F4.3 | API Key 加密存储 | P0 |
| F4.4 | 一键切换 Profile | P0 |
| F4.5 | Profile 导入/导出 | P2 |

### 1.2 API Proxy 服务器

| Feature | 描述 | 优先级 |
|---------|------|--------|
| F4.6 | HTTP Proxy 启动/停止 | P0 |
| F4.7 | 请求拦截 | P0 |
| F4.8 | API Key 注入/替换 | P0 |
| F4.9 | 请求日志记录 | P0 |
| F4.10 | HTTPS 支持（证书生成） | P1 |

### 1.3 使用量追踪

| Feature | 描述 | 优先级 |
|---------|------|--------|
| F4.11 | Token 使用量记录 | P0 |
| F4.12 | 成本估算 | P0 |
| F4.13 | 按日期/Profile 聚合统计 | P0 |
| F4.14 | 使用量图表展示 | P1 |
| F4.15 | 导出 CSV | P2 |

### 1.4 环境变量注入

| Feature | 描述 | 优先级 |
|---------|------|--------|
| F4.16 | 终端环境变量设置 | P0 |
| F4.17 | 自动设置 HTTPS_PROXY | P0 |
| F4.18 | API Key 环境变量 | P1 |

### 1.5 前端组件

| Feature | 描述 | 优先级 |
|---------|------|--------|
| F4.19 | ProfileList 组件 | P0 |
| F4.20 | ProfileForm 组件 | P0 |
| F4.21 | UsageStats 组件 | P0 |
| F4.22 | UsageChart 组件 | P1 |
| F4.23 | ProxyStatus 组件 | P0 |

---

## 2. 系统架构

```
┌─────────────────────────────────────────────────────────────────────┐
│                        API Manager Architecture                      │
├─────────────────────────────────────────────────────────────────────┤
│                                                                      │
│  ┌─────────────────────────────────────────────────────────────┐    │
│  │                    Claude Code / Codex                       │    │
│  │                   (在终端中运行)                              │    │
│  └──────────────────────────┬──────────────────────────────────┘    │
│                             │ HTTP Request                          │
│                             │ (with original API key)               │
│                             ▼                                       │
│  ┌─────────────────────────────────────────────────────────────┐    │
│  │                    API Proxy Server                          │    │
│  │                    (localhost:8080)                          │    │
│  │  ┌─────────────────────────────────────────────────────────┐ │    │
│  │  │  1. 拦截请求                                            │ │    │
│  │  │  2. 替换 API Key (根据当前 Profile)                      │ │    │
│  │  │  3. 转发到真实 API 服务器                                │ │    │
│  │  │  4. 记录使用量 (tokens, cost)                            │ │    │
│  │  │  5. 返回响应                                            │ │    │
│  │  └─────────────────────────────────────────────────────────┘ │    │
│  └──────────────────────────┬──────────────────────────────────┘    │
│                             │                                       │
│                             ▼                                       │
│  ┌─────────────────────────────────────────────────────────────┐    │
│  │                 Real API Server                              │    │
│  │   api.anthropic.com / api.openai.com / ...                  │    │
│  └─────────────────────────────────────────────────────────────┘    │
│                                                                      │
└─────────────────────────────────────────────────────────────────────┘

数据流:
┌──────────┐     ┌──────────┐     ┌──────────┐     ┌──────────┐
│ Profile  │────▶│  Proxy   │────▶│   DB     │────▶│   UI     │
│ (config) │     │ Server   │     │ (usage)  │     │ (stats)  │
└──────────┘     └──────────┘     └──────────┘     └──────────┘
```

---

## 3. 实现细节

### 3.1 Go 后端 - 数据结构

```go
// internal/proxy/types.go
package proxy

import "time"

type Profile struct {
    ID        uint      `gorm:"primaryKey" json:"id"`
    Name      string    `gorm:"uniqueIndex;size:64" json:"name"`
    Provider  string    `gorm:"size:32" json:"provider"`  // anthropic, openai, gemini
    APIKey    string    `gorm:"size:256" json:"-"`        // 加密存储，不返回给前端
    APIKeyHint string   `gorm:"size:16" json:"apiKeyHint"` // 显示前4位
    BaseURL   string    `gorm:"size:256" json:"baseUrl"`
    Active    bool      `json:"active"`
    CreatedAt time.Time `json:"createdAt"`
    UpdatedAt time.Time `json:"updatedAt"`
}

type UsageRecord struct {
    ID           uint      `gorm:"primaryKey" json:"id"`
    ProfileID    uint      `gorm:"index" json:"profileId"`
    ProfileName  string    `json:"profileName"`
    Provider     string    `json:"provider"`
    Model        string    `json:"model"`
    Endpoint     string    `json:"endpoint"`        // /v1/messages
    InputTokens  int       `json:"inputTokens"`
    OutputTokens int       `json:"outputTokens"`
    TotalTokens  int       `json:"totalTokens"`
    Cost         float64   `json:"cost"`            // USD
    Duration     int64     `json:"duration"`        // ms
    StatusCode   int       `json:"statusCode"`
    Timestamp    time.Time `gorm:"index" json:"timestamp"`
    RequestID    string    `json:"requestId"`
}

type UsageStats struct {
    TotalRequests  int     `json:"totalRequests"`
    TotalTokens    int     `json:"totalTokens"`
    TotalCost      float64 `json:"totalCost"`
    InputTokens    int     `json:"inputTokens"`
    OutputTokens   int     `json:"outputTokens"`
    ByProvider     map[string]ProviderStats `json:"byProvider"`
    ByDate         []DateStats              `json:"byDate"`
}

type ProviderStats struct {
    Requests int     `json:"requests"`
    Tokens   int     `json:"tokens"`
    Cost     float64 `json:"cost"`
}

type DateStats struct {
    Date     string  `json:"date"`
    Requests int     `json:"requests"`
    Tokens   int     `json:"tokens"`
    Cost     float64 `json:"cost"`
}

type ProxyStatus struct {
    Running   bool   `json:"running"`
    Port      int    `json:"port"`
    ActiveProfile *Profile `json:"activeProfile,omitempty"`
}
```

### 3.2 Go 后端 - Profile Manager

```go
// internal/proxy/profile.go
package proxy

import (
    "crypto/aes"
    "crypto/cipher"
    "crypto/rand"
    "encoding/hex"
    "errors"
    "io"
    
    "gorm.io/gorm"
)

type ProfileManager struct {
    db       *gorm.DB
    encKey   []byte // 加密密钥
}

func NewProfileManager(db *gorm.DB, encKey string) *ProfileManager {
    // 自动迁移
    db.AutoMigrate(&Profile{})
    
    return &ProfileManager{
        db:     db,
        encKey: []byte(encKey),
    }
}

// encryptAPIKey 加密 API Key
func (m *ProfileManager) encryptAPIKey(key string) (string, error) {
    block, err := aes.NewCipher(m.encKey)
    if err != nil {
        return "", err
    }
    
    plaintext := []byte(key)
    ciphertext := make([]byte, aes.BlockSize+len(plaintext))
    iv := ciphertext[:aes.BlockSize]
    
    if _, err := io.ReadFull(rand.Reader, iv); err != nil {
        return "", err
    }
    
    stream := cipher.NewCFBEncrypter(block, iv)
    stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)
    
    return hex.EncodeToString(ciphertext), nil
}

// decryptAPIKey 解密 API Key
func (m *ProfileManager) decryptAPIKey(encrypted string) (string, error) {
    block, err := aes.NewCipher(m.encKey)
    if err != nil {
        return "", err
    }
    
    ciphertext, err := hex.DecodeString(encrypted)
    if err != nil {
        return "", err
    }
    
    if len(ciphertext) < aes.BlockSize {
        return "", errors.New("ciphertext too short")
    }
    
    iv := ciphertext[:aes.BlockSize]
    ciphertext = ciphertext[aes.BlockSize:]
    
    stream := cipher.NewCFBDecrypter(block, iv)
    stream.XORKeyStream(ciphertext, ciphertext)
    
    return string(ciphertext), nil
}

// CreateProfile 创建新 Profile
func (m *ProfileManager) CreateProfile(name, provider, apiKey, baseURL string) (*Profile, error) {
    // 加密 API Key
    encryptedKey, err := m.encryptAPIKey(apiKey)
    if err != nil {
        return nil, err
    }
    
    profile := &Profile{
        Name:       name,
        Provider:   provider,
        APIKey:     encryptedKey,
        APIKeyHint: apiKey[:4] + "..." + apiKey[len(apiKey)-4:],
        BaseURL:    baseURL,
    }
    
    if err := m.db.Create(profile).Error; err != nil {
        return nil, err
    }
    
    return profile, nil
}

// GetProfile 获取 Profile（包含解密后的 API Key）
func (m *ProfileManager) GetProfile(id uint) (*Profile, error) {
    var profile Profile
    if err := m.db.First(&profile, id).Error; err != nil {
        return nil, err
    }
    
    // 解密 API Key
    decryptedKey, err := m.decryptAPIKey(profile.APIKey)
    if err != nil {
        return nil, err
    }
    
    profile.APIKey = decryptedKey
    return &profile, nil
}

// GetActiveProfile 获取当前活跃的 Profile
func (m *ProfileManager) GetActiveProfile() (*Profile, error) {
    var profile Profile
    if err := m.db.Where("active = ?", true).First(&profile).Error; err != nil {
        return nil, err
    }
    
    decryptedKey, err := m.decryptAPIKey(profile.APIKey)
    if err != nil {
        return nil, err
    }
    
    profile.APIKey = decryptedKey
    return &profile, nil
}

// SetActiveProfile 设置活跃 Profile
func (m *ProfileManager) SetActiveProfile(id uint) error {
    // 先取消所有活跃
    m.db.Model(&Profile{}).Where("active = ?", true).Update("active", false)
    
    // 设置新的活跃
    return m.db.Model(&Profile{}).Where("id = ?", id).Update("active", true).Error
}

// ListProfiles 列出所有 Profile
func (m *ProfileManager) ListProfiles() ([]Profile, error) {
    var profiles []Profile
    err := m.db.Order("created_at DESC").Find(&profiles).Error
    return profiles, err
}

// DeleteProfile 删除 Profile
func (m *ProfileManager) DeleteProfile(id uint) error {
    return m.db.Delete(&Profile{}, id).Error
}
```

### 3.3 Go 后端 - Proxy Server

```go
// internal/proxy/server.go
package proxy

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "strings"
    "sync"
    "time"
    
    "gorm.io/gorm"
)

type Server struct {
    port      int
    server    *http.Server
    profile   *ProfileManager
    db        *gorm.DB
    running   bool
    mu        sync.RWMutex
    active    *Profile
}

func NewServer(port int, db *gorm.DB, profile *ProfileManager) *Server {
    return &Server{
        port:    port,
        profile: profile,
        db:      db,
    }
}

// Start 启动代理服务器
func (s *Server) Start() error {
    s.mu.Lock()
    defer s.mu.Unlock()
    
    if s.running {
        return fmt.Errorf("server already running")
    }
    
    // 获取活跃 Profile
    active, err := s.profile.GetActiveProfile()
    if err != nil {
        // 如果没有活跃 Profile，返回错误
        return fmt.Errorf("no active profile")
    }
    s.active = active
    
    mux := http.NewServeMux()
    mux.HandleFunc("/", s.handleProxy)
    
    s.server = &http.Server{
        Addr:    fmt.Sprintf(":%d", s.port),
        Handler: mux,
    }
    
    go func() {
        if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            // log error
        }
    }()
    
    s.running = true
    return nil
}

// Stop 停止代理服务器
func (s *Server) Stop() error {
    s.mu.Lock()
    defer s.mu.Unlock()
    
    if !s.running {
        return nil
    }
    
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    if err := s.server.Shutdown(ctx); err != nil {
        return err
    }
    
    s.running = false
    return nil
}

// handleProxy 处理代理请求
func (s *Server) handleProxy(w http.ResponseWriter, r *http.Request) {
    start := time.Now()
    
    // 获取当前活跃 Profile
    s.mu.RLock()
    profile := s.active
    s.mu.RUnlock()
    
    if profile == nil {
        http.Error(w, "No active profile", http.StatusServiceUnavailable)
        return
    }
    
    // 读取请求体
    body, err := io.ReadAll(r.Body)
    r.Body.Close()
    if err != nil {
        http.Error(w, "Failed to read request", http.StatusBadRequest)
        return
    }
    
    // 解析请求以提取 token 信息（用于记录）
    var reqData map[string]interface{}
    json.Unmarshal(body, &reqData)
    
    // 创建转发请求
    targetURL := profile.BaseURL + r.URL.Path
    if r.URL.RawQuery != "" {
        targetURL += "?" + r.URL.RawQuery
    }
    
    proxyReq, err := http.NewRequest(r.Method, targetURL, bytes.NewReader(body))
    if err != nil {
        http.Error(w, "Failed to create request", http.StatusInternalServerError)
        return
    }
    
    // 复制请求头
    for key, values := range r.Header {
        // 跳过 Authorization 头，使用我们的 API Key
        if strings.ToLower(key) == "authorization" {
            continue
        }
        for _, value := range values {
            proxyReq.Header.Add(key, value)
        }
    }
    
    // 设置我们的 API Key
    switch profile.Provider {
    case "anthropic":
        proxyReq.Header.Set("x-api-key", profile.APIKey)
        proxyReq.Header.Set("anthropic-version", "2023-06-01")
    case "openai":
        proxyReq.Header.Set("Authorization", "Bearer "+profile.APIKey)
    case "gemini":
        // Gemini 使用 URL 参数
        proxyReq.URL.RawQuery += "&key=" + profile.APIKey
    }
    
    // 发送请求
    client := &http.Client{Timeout: 120 * time.Second}
    resp, err := client.Do(proxyReq)
    if err != nil {
        http.Error(w, "Failed to send request", http.StatusBadGateway)
        return
    }
    defer resp.Body.Close()
    
    // 读取响应
    respBody, err := io.ReadAll(resp.Body)
    if err != nil {
        http.Error(w, "Failed to read response", http.StatusInternalServerError)
        return
    }
    
    // 复制响应头
    for key, values := range resp.Header {
        for _, value := range values {
            w.Header().Add(key, value)
        }
    }
    
    // 写入响应
    w.WriteHeader(resp.StatusCode)
    w.Write(respBody)
    
    // 记录使用量
    go s.recordUsage(profile, r.URL.Path, reqData, respBody, resp.StatusCode, time.Since(start))
}

// recordUsage 记录使用量
func (s *Server) recordUsage(profile *Profile, endpoint string, reqData map[string]interface{}, respBody []byte, statusCode int, duration time.Duration) {
    var inputTokens, outputTokens int
    var model string
    
    // 从请求中提取 input tokens
    if content, ok := reqData["messages"].([]interface{}); ok {
        // 简单估算：每个字符约 0.25 token
        for _, msg := range content {
            if m, ok := msg.(map[string]interface{}); ok {
                if text, ok := m["content"].(string); ok {
                    inputTokens += len(text) / 4
                }
            }
        }
    }
    if cnt, ok := reqData["system"].(string); ok {
        inputTokens += len(cnt) / 4
    }
    
    // 从响应中提取 output tokens 和 model
    var respData map[string]interface{}
    if err := json.Unmarshal(respBody, &respData); err == nil {
        if usage, ok := respData["usage"].(map[string]interface{}); ok {
            if in, ok := usage["input_tokens"].(float64); ok {
                inputTokens = int(in)
            }
            if out, ok := usage["output_tokens"].(float64); ok {
                outputTokens = int(out)
            }
        }
        if m, ok := respData["model"].(string); ok {
            model = m
        }
    }
    
    // 计算成本（简化版本）
    cost := s.calculateCost(profile.Provider, model, inputTokens, outputTokens)
    
    // 保存记录
    record := &UsageRecord{
        ProfileID:    profile.ID,
        ProfileName:  profile.Name,
        Provider:     profile.Provider,
        Model:        model,
        Endpoint:     endpoint,
        InputTokens:  inputTokens,
        OutputTokens: outputTokens,
        TotalTokens:  inputTokens + outputTokens,
        Cost:         cost,
        Duration:     duration.Milliseconds(),
        StatusCode:   statusCode,
        Timestamp:    time.Now(),
    }
    
    s.db.Create(record)
}

// calculateCost 计算成本（简化版本）
func (s *Server) calculateCost(provider, model string, inputTokens, outputTokens int) float64 {
    // 价格表（USD per 1M tokens）
    prices := map[string]struct{ input, output float64 }{
        "claude-3-opus-20240229":     {15, 75},
        "claude-3-sonnet-20240229":   {3, 15},
        "claude-3-haiku-20240307":    {0.25, 1.25},
        "claude-3-5-sonnet-20241022": {3, 15},
        "gpt-4-turbo":                {10, 30},
        "gpt-4o":                     {5, 15},
        "gpt-3.5-turbo":              {0.5, 1.5},
    }
    
    price, ok := prices[model]
    if !ok {
        // 默认价格
        price = prices["claude-3-sonnet-20240229"]
    }
    
    inputCost := float64(inputTokens) * price.input / 1_000_000
    outputCost := float64(outputTokens) * price.output / 1_000_000
    
    return inputCost + outputCost
}

// GetStatus 获取服务器状态
func (s *Server) GetStatus() ProxyStatus {
    s.mu.RLock()
    defer s.mu.RUnlock()
    
    var active *Profile
    if s.active != nil {
        // 不返回 API Key
        active = &Profile{
            ID:         s.active.ID,
            Name:       s.active.Name,
            Provider:   s.active.Provider,
            APIKeyHint: s.active.APIKeyHint,
            BaseURL:    s.active.BaseURL,
            Active:     s.active.Active,
        }
    }
    
    return ProxyStatus{
        Running:      s.running,
        Port:         s.port,
        ActiveProfile: active,
    }
}

// GetUsageStats 获取使用统计
func (s *Server) GetUsageStats(startDate, endDate time.Time) (*UsageStats, error) {
    var records []UsageRecord
    err := s.db.Where("timestamp BETWEEN ? AND ?", startDate, endDate).Find(&records).Error
    if err != nil {
        return nil, err
    }
    
    stats := &UsageStats{
        ByProvider: make(map[string]ProviderStats),
    }
    
    dateMap := make(map[string]*DateStats)
    
    for _, r := range records {
        stats.TotalRequests++
        stats.TotalTokens += r.TotalTokens
        stats.TotalCost += r.Cost
        stats.InputTokens += r.InputTokens
        stats.OutputTokens += r.OutputTokens
        
        // 按 Provider 聚合
        ps := stats.ByProvider[r.Provider]
        ps.Requests++
        ps.Tokens += r.TotalTokens
        ps.Cost += r.Cost
        stats.ByProvider[r.Provider] = ps
        
        // 按日期聚合
        dateStr := r.Timestamp.Format("2006-01-02")
        ds, ok := dateMap[dateStr]
        if !ok {
            ds = &DateStats{Date: dateStr}
            dateMap[dateStr] = ds
        }
        ds.Requests++
        ds.Tokens += r.TotalTokens
        ds.Cost += r.Cost
    }
    
    // 转换 dateMap 为 slice 并排序
    for _, ds := range dateMap {
        stats.ByDate = append(stats.ByDate, *ds)
    }
    
    return stats, nil
}
```

### 3.4 App.go 绑定

```go
// app.go (新增)
// === Profile 管理 ===
func (a *App) ListProfiles() ([]proxy.Profile, error) {
    return a.profileManager.ListProfiles()
}

func (a *App) CreateProfile(name, provider, apiKey, baseURL string) error {
    _, err := a.profileManager.CreateProfile(name, provider, apiKey, baseURL)
    return err
}

func (a *App) SetActiveProfile(id uint) error {
    err := a.profileManager.SetActiveProfile(id)
    if err != nil {
        return err
    }
    
    // 更新 proxy server 的活跃 profile
    if a.proxyServer.GetStatus().Running {
        active, _ := a.profileManager.GetActiveProfile()
        a.proxyServer.SetActiveProfile(active)
    }
    
    return nil
}

func (a *App) DeleteProfile(id uint) error {
    return a.profileManager.DeleteProfile(id)
}

// === Proxy 管理 ===
func (a *App) StartProxy() error {
    return a.proxyServer.Start()
}

func (a *App) StopProxy() error {
    return a.proxyServer.Stop()
}

func (a *App) GetProxyStatus() proxy.ProxyStatus {
    return a.proxyServer.GetStatus()
}

// === 使用量统计 ===
func (a *App) GetUsageStats(startDate, endDate string) (*proxy.UsageStats, error) {
    start, _ := time.Parse("2006-01-02", startDate)
    end, _ := time.Parse("2006-01-02", endDate)
    return a.proxyServer.GetUsageStats(start, end)
}
```

### 3.5 前端 - Store

```typescript
// frontend/src/stores/apiProxyStore.ts
import { create } from 'zustand'
import { Profile, ProxyStatus, UsageStats } from '../types/apiProxy'
import {
  ListProfiles,
  CreateProfile,
  SetActiveProfile,
  DeleteProfile,
  StartProxy,
  StopProxy,
  GetProxyStatus,
  GetUsageStats,
} from '../../wailsjs/go/main/App'

interface ApiProxyState {
  profiles: Profile[]
  status: ProxyStatus | null
  stats: UsageStats | null
  loading: boolean
  error: string | null
  
  // Actions
  fetchProfiles: () => Promise<void>
  createProfile: (name: string, provider: string, apiKey: string, baseURL: string) => Promise<void>
  setActiveProfile: (id: number) => Promise<void>
  deleteProfile: (id: number) => Promise<void>
  startProxy: () => Promise<void>
  stopProxy: () => Promise<void>
  fetchStatus: () => Promise<void>
  fetchStats: (startDate: string, endDate: string) => Promise<void>
}

export const useApiProxyStore = create<ApiProxyState>((set, get) => ({
  profiles: [],
  status: null,
  stats: null,
  loading: false,
  error: null,
  
  fetchProfiles: async () => {
    set({ loading: true, error: null })
    try {
      const profiles = await ListProfiles()
      set({ profiles, loading: false })
    } catch (err) {
      set({ error: String(err), loading: false })
    }
  },
  
  createProfile: async (name, provider, apiKey, baseURL) => {
    set({ loading: true, error: null })
    try {
      await CreateProfile(name, provider, apiKey, baseURL)
      await get().fetchProfiles()
    } catch (err) {
      set({ error: String(err), loading: false })
    }
  },
  
  setActiveProfile: async (id) => {
    set({ loading: true, error: null })
    try {
      await SetActiveProfile(id)
      await get().fetchProfiles()
      await get().fetchStatus()
    } catch (err) {
      set({ error: String(err), loading: false })
    }
  },
  
  deleteProfile: async (id) => {
    try {
      await DeleteProfile(id)
      await get().fetchProfiles()
    } catch (err) {
      set({ error: String(err) })
    }
  },
  
  startProxy: async () => {
    set({ loading: true, error: null })
    try {
      await StartProxy()
      await get().fetchStatus()
    } catch (err) {
      set({ error: String(err), loading: false })
    }
  },
  
  stopProxy: async () => {
    try {
      await StopProxy()
      await get().fetchStatus()
    } catch (err) {
      set({ error: String(err) })
    }
  },
  
  fetchStatus: async () => {
    try {
      const status = await GetProxyStatus()
      set({ status })
    } catch (err) {
      set({ error: String(err) })
    }
  },
  
  fetchStats: async (startDate, endDate) => {
    try {
      const stats = await GetUsageStats(startDate, endDate)
      set({ stats })
    } catch (err) {
      set({ error: String(err) })
    }
  },
}))
```

### 3.6 前端 - 组件

```tsx
// frontend/src/components/APIManager/ProfileList.tsx
import { useEffect, useState } from 'react'
import { useApiProxyStore } from '../../stores/apiProxyStore'
import ProfileForm from './ProfileForm'

export default function ProfileList() {
  const { profiles, loading, error, fetchProfiles, setActiveProfile, deleteProfile } = useApiProxyStore()
  const [showForm, setShowForm] = useState(false)
  
  useEffect(() => {
    fetchProfiles()
  }, [])
  
  return (
    <div className="flex flex-col h-full">
      <div className="p-4 border-b border-gray-700">
        <div className="flex items-center justify-between">
          <h3 className="text-sm font-semibold text-gray-400 uppercase">API Profiles</h3>
          <button
            onClick={() => setShowForm(true)}
            className="px-3 py-1 bg-blue-600 rounded text-sm hover:bg-blue-700"
          >
            + Add
          </button>
        </div>
      </div>
      
      {error && (
        <div className="p-2 bg-red-900/50 text-red-300 text-sm">{error}</div>
      )}
      
      <div className="flex-1 overflow-auto p-2">
        {profiles.length === 0 ? (
          <div className="text-center text-gray-500 py-8">
            No profiles yet. Add one to get started.
          </div>
        ) : (
          <div className="space-y-2">
            {profiles.map(profile => (
              <div
                key={profile.id}
                className={`
                  p-3 rounded border cursor-pointer transition-colors
                  ${profile.active 
                    ? 'border-blue-500 bg-blue-900/20' 
                    : 'border-gray-700 hover:border-gray-600'}
                `}
                onClick={() => !profile.active && setActiveProfile(profile.id)}
              >
                <div className="flex items-center justify-between">
                  <div>
                    <div className="font-medium">{profile.name}</div>
                    <div className="text-sm text-gray-500">
                      {profile.provider} • {profile.apiKeyHint}
                    </div>
                  </div>
                  
                  <div className="flex items-center gap-2">
                    {profile.active && (
                      <span className="px-2 py-0.5 bg-green-900 text-green-300 text-xs rounded">
                        Active
                      </span>
                    )}
                    <button
                      onClick={(e) => {
                        e.stopPropagation()
                        if (confirm('Delete this profile?')) {
                          deleteProfile(profile.id)
                        }
                      }}
                      className="p-1 text-gray-500 hover:text-red-400"
                    >
                      <TrashIcon className="w-4 h-4" />
                    </button>
                  </div>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
      
      {showForm && <ProfileForm onClose={() => setShowForm(false)} />}
    </div>
  )
}
```

```tsx
// frontend/src/components/APIManager/ProfileForm.tsx
import { useState } from 'react'
import { useApiProxyStore } from '../../stores/apiProxyStore'

interface ProfileFormProps {
  onClose: () => void
}

const PROVIDERS = [
  { id: 'anthropic', name: 'Anthropic (Claude)', defaultURL: 'https://api.anthropic.com' },
  { id: 'openai', name: 'OpenAI (GPT)', defaultURL: 'https://api.openai.com' },
  { id: 'gemini', name: 'Google (Gemini)', defaultURL: 'https://generativelanguage.googleapis.com' },
  { id: 'custom', name: 'Custom', defaultURL: '' },
]

export default function ProfileForm({ onClose }: ProfileFormProps) {
  const { createProfile, loading, error } = useApiProxyStore()
  
  const [name, setName] = useState('')
  const [provider, setProvider] = useState('anthropic')
  const [apiKey, setApiKey] = useState('')
  const [baseURL, setBaseURL] = useState(PROVIDERS[0].defaultURL)
  
  const handleProviderChange = (id: string) => {
    setProvider(id)
    const p = PROVIDERS.find(p => p.id === id)
    if (p) setBaseURL(p.defaultURL)
  }
  
  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    
    await createProfile(name, provider, apiKey, baseURL)
    
    if (!error) {
      onClose()
    }
  }
  
  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
      <div className="bg-gray-800 rounded-lg p-6 w-96">
        <h3 className="text-lg font-semibold mb-4">Add API Profile</h3>
        
        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="block text-sm text-gray-400 mb-1">Name</label>
            <input
              type="text"
              value={name}
              onChange={e => setName(e.target.value)}
              className="w-full px-3 py-2 bg-gray-700 rounded border border-gray-600 focus:border-blue-500 focus:outline-none"
              placeholder="My API Profile"
              required
            />
          </div>
          
          <div>
            <label className="block text-sm text-gray-400 mb-1">Provider</label>
            <select
              value={provider}
              onChange={e => handleProviderChange(e.target.value)}
              className="w-full px-3 py-2 bg-gray-700 rounded border border-gray-600 focus:border-blue-500 focus:outline-none"
            >
              {PROVIDERS.map(p => (
                <option key={p.id} value={p.id}>{p.name}</option>
              ))}
            </select>
          </div>
          
          <div>
            <label className="block text-sm text-gray-400 mb-1">API Key</label>
            <input
              type="password"
              value={apiKey}
              onChange={e => setApiKey(e.target.value)}
              className="w-full px-3 py-2 bg-gray-700 rounded border border-gray-600 focus:border-blue-500 focus:outline-none"
              placeholder="sk-..."
              required
            />
          </div>
          
          <div>
            <label className="block text-sm text-gray-400 mb-1">Base URL</label>
            <input
              type="url"
              value={baseURL}
              onChange={e => setBaseURL(e.target.value)}
              className="w-full px-3 py-2 bg-gray-700 rounded border border-gray-600 focus:border-blue-500 focus:outline-none"
              placeholder="https://api.example.com"
              required
            />
          </div>
          
          {error && (
            <div className="p-2 bg-red-900/50 text-red-300 text-sm rounded">
              {error}
            </div>
          )}
          
          <div className="flex justify-end gap-2 pt-2">
            <button
              type="button"
              onClick={onClose}
              className="px-4 py-2 text-gray-400 hover:text-white"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={loading}
              className="px-4 py-2 bg-blue-600 rounded hover:bg-blue-700 disabled:opacity-50"
            >
              {loading ? 'Saving...' : 'Save'}
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}
```

```tsx
// frontend/src/components/APIManager/UsageStats.tsx
import { useEffect, useState } from 'react'
import { useApiProxyStore } from '../../stores/apiProxyStore'

export default function UsageStats() {
  const { stats, fetchStats } = useApiProxyStore()
  const [period, setPeriod] = useState<'7d' | '30d' | 'all'>('7d')
  
  useEffect(() => {
    const end = new Date()
    let start = new Date()
    
    switch (period) {
      case '7d':
        start.setDate(start.getDate() - 7)
        break
      case '30d':
        start.setDate(start.getDate() - 30)
        break
      case 'all':
        start = new Date('2020-01-01')
        break
    }
    
    fetchStats(start.toISOString().split('T')[0], end.toISOString().split('T')[0])
  }, [period])
  
  if (!stats) {
    return <div className="p-4 text-gray-500">Loading...</div>
  }
  
  return (
    <div className="p-4">
      <div className="flex items-center justify-between mb-4">
        <h3 className="text-sm font-semibold text-gray-400 uppercase">Usage</h3>
        <select
          value={period}
          onChange={e => setPeriod(e.target.value as any)}
          className="px-2 py-1 bg-gray-700 rounded text-sm"
        >
          <option value="7d">Last 7 days</option>
          <option value="30d">Last 30 days</option>
          <option value="all">All time</option>
        </select>
      </div>
      
      {/* 总览卡片 */}
      <div className="grid grid-cols-2 gap-3 mb-4">
        <div className="p-3 bg-gray-800 rounded">
          <div className="text-2xl font-bold">{stats.totalRequests}</div>
          <div className="text-sm text-gray-500">Requests</div>
        </div>
        <div className="p-3 bg-gray-800 rounded">
          <div className="text-2xl font-bold">{formatNumber(stats.totalTokens)}</div>
          <div className="text-sm text-gray-500">Tokens</div>
        </div>
        <div className="p-3 bg-gray-800 rounded">
          <div className="text-2xl font-bold">${stats.totalCost.toFixed(2)}</div>
          <div className="text-sm text-gray-500">Cost</div>
        </div>
        <div className="p-3 bg-gray-800 rounded">
          <div className="text-2xl font-bold">{formatNumber(stats.inputTokens)}</div>
          <div className="text-sm text-gray-500">Input</div>
        </div>
      </div>
      
      {/* 按日期图表（简化版） */}
      <div className="border-t border-gray-700 pt-4">
        <h4 className="text-sm text-gray-400 mb-2">By Date</h4>
        <div className="h-32 flex items-end gap-1">
          {stats.byDate.slice(-14).map((d, i) => (
            <div
              key={i}
              className="flex-1 bg-blue-600 rounded-t"
              style={{ height: `${(d.cost / maxCost(stats.byDate)) * 100}%` }}
              title={`${d.date}: $${d.cost.toFixed(2)}`}
            />
          ))}
        </div>
      </div>
    </div>
  )
}

function formatNumber(n: number): string {
  if (n >= 1_000_000) return (n / 1_000_000).toFixed(1) + 'M'
  if (n >= 1_000) return (n / 1_000).toFixed(1) + 'K'
  return n.toString()
}

function maxCost(data: { cost: number }[]): number {
  if (data.length === 0) return 1
  let max = 0
  for (const d of data) {
    if (d.cost > max) max = d.cost
  }
  return max || 1
}
```

---

## 4. 测试计划

### 4.1 单元测试

```go
// internal/proxy/profile_test.go
func TestEncryptDecryptAPIKey(t *testing.T) {
    pm := &ProfileManager{encKey: []byte("0123456789abcdef")} // 16 bytes
    
    original := "sk-ant-api03-xxxxx"
    
    encrypted, err := pm.encryptAPIKey(original)
    require.NoError(t, err)
    
    decrypted, err := pm.decryptAPIKey(encrypted)
    require.NoError(t, err)
    
    assert.Equal(t, original, decrypted)
}

func TestCreateProfile(t *testing.T) {
    // 使用内存数据库测试
    db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
    db.AutoMigrate(&Profile{})
    
    pm := NewProfileManager(db, "0123456789abcdef")
    
    profile, err := pm.CreateProfile("test", "anthropic", "sk-test-123", "https://api.anthropic.com")
    require.NoError(t, err)
    
    assert.Equal(t, "test", profile.Name)
    assert.Equal(t, "anthropic", profile.Provider)
    assert.Contains(t, profile.APIKeyHint, "sk-t")
}
```

### 4.2 集成测试

```go
// internal/proxy/server_test.go
func TestProxyServer(t *testing.T) {
    // 创建测试数据库
    db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
    db.AutoMigrate(&Profile{}, &UsageRecord{})
    
    pm := NewProfileManager(db, "0123456789abcdef")
    pm.CreateProfile("test", "anthropic", "sk-test", "https://api.anthropic.com")
    pm.SetActiveProfile(1)
    
    server := NewServer(18080, db, pm)
    
    err := server.Start()
    require.NoError(t, err)
    defer server.Stop()
    
    // 发送测试请求
    // ...
}
```

### 4.3 手动测试清单

- [ ] 创建 Profile，保存后能看到 API Key hint
- [ ] 切换 Profile，状态正确更新
- [ ] 启动 Proxy，状态显示 Running
- [ ] 在终端设置 `HTTPS_PROXY=http://localhost:8080`
- [ ] 运行 Claude Code，请求通过 Proxy
- [ ] 查看使用量统计，数据正确

---

## 5. 验收标准

| 标准 | 描述 |
|------|------|
| Profile 管理 | CRUD 功能完整 |
| Proxy 运行 | 启动/停止正常 |
| API Key 注入 | 请求正确替换 Key |
| 使用量追踪 | Token 和成本正确记录 |
| 终端集成 | 环境变量自动设置 |

---

## 6. 发布检查清单

- [ ] 所有单元测试通过
- [ ] 集成测试通过
- [ ] 手动测试清单完成
- [ ] API Key 加密验证
- [ ] 更新 CHANGELOG.md
- [ ] 创建 Git tag: v0.5.0-beta
- [ ] GitHub Release 构建

---

## 7. 时间估算

| 任务 | 时间 |
|------|------|
| Profile Manager 实现 | 3 天 |
| Proxy Server 实现 | 4 天 |
| 使用量追踪 | 2 天 |
| 前端组件 | 3 天 |
| 测试 | 2 天 |
| Bug 修复和优化 | 2 天 |
| 文档和发布 | 1 天 |
| **总计** | **17 天 (3-4 周)** |
