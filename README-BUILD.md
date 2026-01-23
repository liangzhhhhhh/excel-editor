# 环境配置说明

## 环境区分

项目支持开发和生产两种环境，通过环境变量或 build tags 来区分。

## 开发环境

### 方式一：使用环境变量（推荐，最简单）

```bash
# Windows
set WAILS_ENV=dev
wails dev

# Linux/Mac
export WAILS_ENV=dev
wails dev
```

### 方式二：使用 API 环境变量

```bash
# Windows
set API_BASE_URL=http://127.0.0.1:3000
wails dev

# Linux/Mac
export API_BASE_URL=http://127.0.0.1:3000
wails dev
```

### 方式三：使用 build tags

```bash
# 需要修改构建命令，添加 -tags dev
go build -tags dev
```

## 生产环境

```bash
# 直接构建，不设置环境变量
wails build
```

## 配置说明

- **开发环境**: 
  - BaseURL: `http://127.0.0.1:3000`
  - AuthURL: `http://wechat.aaagame.com`
  - OaURL: `https://fatcat-admin-test.54030.com`
  
- **生产环境**:
  - BaseURL: `http://192.168.1.6:3001`
  - AuthURL: `http://wechat.aaagame.com`
  - OaURL: `https://fatcat-admin-test.54030.com`

## 环境变量优先级

1. **最高优先级**: `API_BASE_URL`, `API_AUTH_URL`, `API_OA_URL` - 直接指定 URL
2. **次优先级**: `WAILS_ENV=dev` - 指定为开发环境
3. **默认**: 根据 build tags 或运行模式自动判断

## 快速开始

开发环境：
```bash
# Windows
set WAILS_ENV=dev && wails dev

# Linux/Mac  
export WAILS_ENV=dev && wails dev
```
