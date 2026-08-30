# FlyOS CLI Shell

FlyOS 是一个基于 Go 语言开发的轻量级交互式命令 REPL 环境。它支持内置命令管理、外部可执行文件动态扫描加载[cite: 5]、基于 TOML 的命令帮助文档与分类解析[cite: 1, 3]，以及配置文件热重载（Hot-Reload）功能。

---

## ✨ 核心特性

- **交互式 Shell 环境 (REPL)**：基于 `readline` 提供命令行交互，支持命令历史记录。
- **内置命令与扩展**：内置 `list`、`help`、`env`、`exit` 等基础系统命令。
- **动态外部命令扫描**：自动递归扫描配置目录下的可执行文件/脚本（支持 Linux/macOS 可执行权限、Shebang 脚本及 Windows 常见脚本）[cite: 5]。
- **结构化帮助文档**：支持解析 `desc.toml` 中的 TOML 层级结构，提供详细的命令说明、用法、参数、标志及返回值信息[cite: 1, 3]。
- **模糊搜索与分类管理**：支持按分类检索命令，支持对内置及外部命令名称与描述进行模糊搜索。
- **配置热重载 (FSNotify)**：实时监听配置与帮助文档修改，无需重启即可动态更新命令列表与帮助信息。

---

## 📁 目录结构与配置

项目运行时会在用户根目录下创建或读取配置文件夹 `~/.flyos/`（也可以通过环境变量 `FLYOS_HOME` 指定）：

```text
~/.flyos/
├── config.toml    # 主配置文件（扫描路径、环境变量、排除项）
└── desc.toml      # 外部命令帮助文档描述文件
```

1. config.toml 配置示例
```toml
# 命令扫描目录列表
commands_dirs = [
    "./tools",
    "/usr/local/networks"
]

# 排除扫描的文件名
excludes = ["ignore.sh"]

# 扩展环境变量（支持字符串或数组形态）
[env]
PATH = [
    "/usr/local/bin",
    "/usr/sbin",
    "/usr/bin"
]
DEBUG = "true"
```

2. desc.toml 配置示例
```toml
# 分类.命令名 格式的描述结构
[sys.hello]
desc = "hello.sh 远程连接 CPE 并显示过程"
usage = "hello.sh [SUBCOMMAND]"
subcommands = ["help      打印此帮助信息"]
args = ["<CPE_NAME>     # 要连接的 CPE 名称"]
flags = ["-h, --help    # 打印帮助信息"]
returns = ["连接过程的实时控制台输出"]
```

## 🛠️ 构建与运行
### 编译构建
```bash
# 执行构建（默认构建 Darwin/macOS 目标，可根据需求在 Makefile 中切换为 Linux）
make build
```

### 运行 FlyOS
```bash
# 运行 FlyOS CLI Shell
./flyos
```

## 💡 内置命令说明
启动 FlyOS REPL 后，可使用以下内置命令：
| 命令 | 分类 | 描述 | 用法 |
| :--- | :--- | :--- | :--- |
| `help` | `sys` | 显示命令帮助、分类列表，支持关键字模糊搜索 | `help [COMMAND\|CATEGORY\|KEYWORD]` |
| `list` | `sys` | 按分类打印所有已注册的内置与外部命令 | `list` |
| `env` | `sys` | 打印当前的系统及自定义环境变量 | `env [VAR...]` |
| `exit` | `sys` | 退出 FlyOS 交互式 Shell | `exit` |

### 使用示例
1. 查看命令帮助：
```bash
flyos> help             # 查看所有内置命令和外部命令分类
flyos> help sys         # 查看 sys 分类下的所有命令
flyos> help hello.sh    # 查看 hello.sh 命令的详细用法与参数
flyos> help connect     # 模糊搜索包含 connect 的命令
```
2. 查看当前环境变量：
```bash
flyos> env PATH DEBUG
```
3. 执行外部命令：
```bash
flyos> hello.sh my-cpe-01
```

## ⚙️ 配置路径与环境变量优先级
系统配置路径与环境变量的解析遵循以下逻辑：
1. FlyOS 配置目录：
- 优先读取 FLYOS_HOME 环境变量指定的路径。
- 其次使用当前用户主目录（~/.flyos）。
- 兜底使用当前工作目录（./.flyos）。
2. 环境注入：
- 系统自动注入默认环境变量 USER=fly 与 VERSION=1.0.0。
- config.toml 中配置的环境变量优先级高于系统默认环境变量。