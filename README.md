# 🛰️ FlyOS 架构设计文档

FlyOS 是一个集网络与安全的操作系统，支持多种控制通道（REPL / REST / MCP）通过 IPC 与守护进程通信，统一调度网络模块执行操作。

---

## 🔹 总体架构

```mermaid
graph LR

    %% ======================
    %% Clients Layer
    %% ======================
    subgraph Clients["Clients (外部控制入口)"]
        direction TB
        REPL[REPL 客户端<br/>(cmd/repl - 使用 DSL)]
        CLI[CLI/Tools<br/>(其它命令行工具)]
        RESTClient[REST Clients<br/>(前端/Agent)]
        MCPClient[MCP Clients<br/>(AI Agents/LLM)]
    end

    %% 对接 Daemon
    REPL -->|IPC: Unix Socket| DAEMON[flyos-daemon]
    CLI  -->|IPC: Unix Socket| DAEMON
    RESTClient -->|HTTP/JSON| REST
    MCPClient  -->|WebSocket/JSON-RPC| MCP

    %% ======================
    %% Daemon 外部节点
    %% ======================
    DAEMON --> REST
    DAEMON --> MCP
    DAEMON --> Runtime


    %% ======================
    %% Daemon 内部结构
    %% ======================
    subgraph Daemon["flyos-daemon (核心执行进程)"]
        direction LR

        REST[REST Server<br/>HTTP/JSON]
        MCP[MCP Server<br/>WebSocket/JSON-RPC]
        Runtime[runtime.Manager<br/>(统一命令调度)]
        Modules[Modules<br/>modules/*<br/>(路由/ACL/NIC/...)]
    end

    Runtime --> Modules

    %% 运行机制说明
    note right of REST
        REST/MCP 在 daemon 内部监听外部请求，
        通过 runtime.Manager.Exec() 统一调度模块。
    end
    note right of REPL
        REPL 客户端解析 DSL 命令，
        通过 IPC 调用 daemon 的 ExecDSL() 方法。
    end
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
        Converter
        Modules/Managers
    end

    %% DSL路径
    REPL -->|输入DSL文本| DSLParser
    DSLParser -->|生成Command对象| Runtime
    Runtime -->|ExecuteDSLCommand| Converter
    Converter -->|ConvertFromDSL| ModuleObject
    ModuleObject -->|Execute/Manager方法| Manager
    Manager -->|系统调用/操作| Kernel/Net/Config

    %% REST/MCP路径
    REST -->|JSON/Args| Runtime
    MCP -->|JSON/Args| Runtime
    Runtime -->|ExecuteFromJSON| Converter
    Converter -->|ConvertFromJSON| ModuleObject
    ModuleObject -->|Execute/Manager方法| Manager
    Manager -->|系统调用/操作| Kernel/Net/Config

```
流程说明
1. DSL（REPL）
 - 用户输入 DSL 文本（如 route add static { prefix 10.0.0.0/24; via 192.168.1.1 }）。
 - DSL Parser 解析成 Command 对象。
 - Runtime 的 ExecuteDSLCommand 接收 Command 对象。
2. Runtime
 - 根据 Command.Kind 调用对应 Converter。
 - Converter 将 DSL Command 转成模块对象（如 Route/BGP/OSPF）。
 - 模块对象内部有 Execute(verb string) 方法，封装具体的 Manager 调用。
3. REST/MCP
 - 直接传 JSON/Args 给 Runtime。
 - Runtime 使用 ExecuteFromJSON。
 - Converter 将 JSON 转成模块对象。
 - 模块对象调用 Manager 执行。
4. 模块/Manager
 - 负责真正系统操作，如：
  - routing.CLIManager 调用系统命令。
  - routing.NetlinkManager 调用 netlink。
  - acl.Manager 管理防火墙规则。
 - Manager 可以复用同一套接口，实现统一调用。
5. 最终系统效果
- 所有路径（DSL、REST、MCP）都通过 Runtime + Converter + Module/Manager 执行。
- 可以统一权限检查、事件发布、日志等。

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
