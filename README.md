# phototidy

一个用于整理照片和视频的小工具，可以根据文件的拍摄日期将其分类到相应的年月目录中。

## 功能特性

- 自动识别图片和视频文件的拍摄日期
  - 图片：通过 EXIF 数据获取拍摄日期
  - 视频：通过 MP4/MOV 元数据获取创建时间
  - 其他：使用文件修改时间作为备选
- 按照年月创建目录结构（如：2025-01）
- 自动重命名文件为标准格式（IMG_YYYYMMDD_HHMMSS.ext 或 VID_YYYYMMDD_HHMMSS.ext）
- 处理同一秒拍摄的多个文件（添加序号后缀）
- 避免重复处理已分类的文件
- 生成处理日志文件
- 保留原始文件的时间戳

## 支持的格式

### 图片格式
- JPG/JPEG
- PNG
- TIFF/TIF
- HEIC

### 视频格式
- MP4
- MOV
- AVI

## 安装

```bash
# 克隆项目
git clone https://github.com/chuanglibs/phototidy.git

# 进入项目目录
cd phototidy

# 构建二进制文件
go build -ldflags="-s -w" -o phototidy .
```

或者直接从 [Releases](https://github.com/chuanglibs/phototidy/releases) 页面下载预编译的二进制文件。

## 使用方法

### 基本命令

```bash
# 整理当前目录下的照片和视频
./phototidy date

# 指定要处理的目录
./phototidy date -d /path/to/photos

# 启用详细输出
./phototidy date -d /path/to/photos -v
```

### 参数说明

- `-d, --dir`: 指定要处理的目录，默认为当前目录（`.`）
- `-v, --verbose`: 启用详细输出，显示处理过程中的每一步操作

### 查看帮助

```bash
./phototidy --help
./phototidy date --help
```

### Mac系统注意事项
使用MacOS时，可能会遇到 `com.apple.quarantine` 属性问题，导致程序无法正常运行，在使用前可以先用命令移除该属性。

```zsh
xattr -d com.apple.quarantine phototidy
```

### 文件命名规则

## 文件命名规则

- 图片文件：`IMG_YYYYMMDD_HHMMSS.ext`
- 视频文件：`VID_YYYYMMDD_HHMMSS.ext`
- 重复文件：`IMG_YYYYMMDD_HHMMSS_001.ext`

其中 YYYY 是年份，MM 是月份，DD 是日期，HHMMSS 是时分秒。

## 目录结构

处理完成后，文件会被移动到以年月命名的子目录中，例如：

```
目标目录/
├── 2023-01/
│   ├── IMG_20230115_123045.jpg
│   ├── IMG_20230115_123045_001.jpg
│   └── VID_20230120_142230.mp4
├── 2023-02/
│   ├── IMG_20230210_091522.jpg
│   └── VID_20230212_183000.mov
└── 2023-01-15_12_30_45.log  # 处理日志文件
```

## 注意事项

- 程序会保留原始文件的时间戳
- 已经在年月目录中的文件不会被重复处理
- 如果遇到同名文件，会自动添加序号后缀
- 程序会在处理目录中生成日志文件，记录处理过程
- 支持嵌套目录扫描，但不会处理已在年月目录结构中的文件

## 自动发布

项目配置了 GitHub Actions 自动发布功能，当创建新的 Git 标签（如 `v1.0.0`）时，会自动编译并发布到 Releases 页面。

要发布编译后的可执行文件，请确保：

- 自动发布：当打上 `v*.*.*` 格式的标签并推送时，自动构建多平台二进制文件并发布到 GitHub Releases
  - 针对 macOS 平台，自动移除 `com.apple.quarantine` 属性，防止系统弹出安全警告

- 正确使用 `go build` 命令进行编译
- 通过环境变量（GOOS/GOARCH）进行交叉编译以支持多平台（Linux/amd64, Linux/arm64, Windows/amd64, macOS/amd64, macOS/arm64）
- 使用 `-ldflags="-s -w"` 优化编译，减小二进制文件大小
- 构建后验证产物为可执行文件而非源码
- 将编译后的可执行文件打包成 ZIP 格式发布

发布的文件包含：

- 各平台的编译后可执行文件
- LICENSE 许可证文件
- README.md 说明文档

发布流程包含详细的调试信息，包括环境变量设置、Go环境信息、构建状态检查、可执行文件类型检查以及压缩包内容验证，以确保发布的是编译后的程序而非源代码。

## 开发

本项目使用 Go 1.24.4 和 Cobra CLI 库开发。

```bash
# 安装依赖
go mod tidy

# 运行项目
go run main.go date -d /path/to/photos

# 构建项目
go build -o phototidy .
```

## 许可证

本项目遵循 [LICENSE](./LICENSE) 许可证。