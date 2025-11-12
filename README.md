# 🛰️ FlyOS 架构设计文档

FlyOS 是一个轻量级网络操作系统，支持多种控制通道（REPL / REST / MCP）通过 IPC 与守护进程通信，统一调度网络模块执行操作。

---

## 🔹 总体架构

```mermaid
graph LR
    REPL[REPL (DSL)] -->|IPC| Daemon[flyos-daemon]
    REST[REST Server (JSON)] -->|IPC| Daemon
    MCP[MCP Server (JSON-RPC)] -->|IPC| Daemon
    Daemon --> Runtime[runtime.Manager]
    Runtime --> Modules[modules.*]
```

说明：
- REPL 输入 DSL → ExecDSL()
- REST / MCP → Exec()
- Runtime 调度模块执行实际业务逻辑


## 🔹 REPL DSL 执行时序图

```mermaid
sequenceDiagram
    participant User as 用户
    participant REPL as flyos-client (REPL)
    participant Daemon as flyos-daemon
    participant Module as 模块组 (route/acl/nic)

    User->>REPL: 输入 DSL 命令
    REPL->>REPL: DSL Parser 解析
    REPL->>Daemon: ExecDSL(input)
    Daemon->>Module: 调度对应模块执行
    Module-->>Daemon: 返回执行结果
    Daemon-->>REPL: stdout / stderr
    REPL-->>User: 显示结果
```

## 🔹 REST / MCP 执行时序图
```mermaid
sequenceDiagram
    participant Client as REST / MCP
    participant Daemon as flyos-daemon
    participant Runtime as runtime.Manager
    participant Module as 模块组

    Client->>Daemon: JSON 请求 (REST / MCP)
    Daemon->>Runtime: Exec(cmd, args)
    Runtime->>Module: 调用对应模块
    Module-->>Runtime: 返回执行结果
    Runtime-->>Daemon: stdout / stderr
    Daemon-->>Client: JSON / JSON-RPC 响应
```
## 🔹 模块注册流程
```mermaid
graph TD
    Daemon --> Runtime
    Runtime --> ModuleRegistry[Module Registry]
    ModuleRegistry --> RouteModule[modules/routing]
    ModuleRegistry --> ACLModule[modules/acl]
    ModuleRegistry --> NICModule[modules/nic]
```

## 🔹 数据流总览
```mermaid
flowchart LR
    subgraph Client Layer
        REPL
        REST
        MCP
    end

    subgraph Daemon Layer
        Runtime
        Modules
    end

    REPL -->|DSL| Runtime
    REST -->|Command/Args| Runtime
    MCP -->|Command/Args| Runtime
    Runtime --> Modules
```

## 典型目录结构
```mermaid
graph TD
    FlyOS[flyos/]
    FlyOS --> cmd
    FlyOS --> pkg
    FlyOS --> modules
    cmd --> repl
    cmd --> client
    cmd --> daemon
    pkg --> dsl
    pkg --> ipc
    pkg --> runtime
    pkg --> module
    pkg --> auth
    modules --> acl
    modules --> routing
    modules --> nic
    modules --> nat
    modules --> tunnel
    modules --> vrf
```

