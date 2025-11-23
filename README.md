# Celery Go Client

一个基于 **Celery 2.0 协议**的 Golang 客户端包,用于向 Celery 分布式任务队列发送任务。

## 特性

- ✅ 完全兼容 Celery 协议版本 1 (Protocol v1)
- ✅ 支持 Redis 和 RabbitMQ (AMQP) 作为消息代理
- ✅ 支持位置参数和关键字参数
- ✅ 支持任务调度 (ETA) 和过期时间
- ✅ 支持自定义队列和交换机
- ✅ 简洁易用的 API 设计
- ✅ 完整的单元测试覆盖

## 安装

```bash
go get github.com/celery-go-client
```

## 快速开始

### 使用 Redis 作为 Broker

```go
package main

import (
    "context"
    "log"
    
    celery "github.com/celery-go-client"
)

func main() {
    // 创建 Redis broker
    broker, err := celery.NewRedisBroker(celery.RedisBrokerConfig{
        Addr:     "localhost:6379",
        Password: "",
        DB:       0,
        Queue:    "celery",
    })
    if err != nil {
        log.Fatal(err)
    }
    defer broker.Close()

    // 创建 Celery 客户端
    client := celery.NewClient(celery.ClientConfig{
        Broker: broker,
        Queue:  "celery",
    })
    defer client.Close()

    // 发送任务
    ctx := context.Background()
    taskID, err := client.SendTaskWithArgs(ctx, "tasks.add", 10, 20)
    if err != nil {
        log.Fatal(err)
    }
    
    log.Printf("Task sent! ID: %s", taskID)
}
```

### 使用 RabbitMQ (AMQP) 作为 Broker (默认 Base64 编码)

```go
package main

import (
    "context"
    "log"
    
    celery "github.com/celery-go-client"
)

func main() {
    // 创建 AMQP broker
    broker, err := celery.NewAMQPBroker(celery.AMQPBrokerConfig{
        URL:      "amqp://guest:guest@localhost:5672/",
        Exchange: "celery",
        Queue:    "celery",
    })
    if err != nil {
        log.Fatal(err)
    }
    defer broker.Close()

    // 创建 Celery 客户端
    client := celery.NewClient(celery.ClientConfig{
        Broker:   broker,
        Queue:    "celery",
        Exchange: "celery",
    })
    
    // 如果需要发送原始 JSON 消息体 (非 Base64 编码),请使用以下配置:
    /*
    client := celery.NewClient(celery.ClientConfig{
        Broker:   broker,
        Queue:    "celery",
        Exchange: "celery",
        UseRawJSONBody: true, // 启用原始 JSON 消息体
    })
    */
    defer client.Close()

    // 发送任务
    ctx := context.Background()
    taskID, err := client.SendTaskWithArgs(ctx, "tasks.multiply", 5, 8)
    if err != nil {
        log.Fatal(err)
    }
    
    log.Printf("Task sent! ID: %s", taskID)
}
```

## 使用示例

### 1. 发送带位置参数的任务 (默认 Base64 编码)

```go
taskID, err := client.SendTaskWithArgs(ctx, "tasks.add", 10, 20)
```

### 2. 发送带关键字参数的任务 (默认 Base64 编码)

```go
taskID, err := client.SendTaskWithKwargs(ctx, "tasks.process_data", map[string]interface{}{
    "name":   "John Doe",
    "age":    30,
    "active": true,
})
```

### 3. 发送带位置参数和关键字参数的任务 (默认 Base64 编码)

```go
taskID, err := client.SendTask(ctx, "tasks.complex_task", &celery.TaskOptions{
    Args: []interface{}{"arg1", "arg2"},
    Kwargs: map[string]interface{}{
        "key1": "value1",
        "key2": 42,
    },
})
```

### 4. 发送定时任务 (ETA) (默认 Base64 编码)

```go
eta := time.Now().Add(5 * time.Minute)
taskID, err := client.SendTask(ctx, "tasks.scheduled_task", &celery.TaskOptions{
    Args: []interface{}{"scheduled"},
    ETA:  &eta,
})
```

### 5. 发送带过期时间的任务 (默认 Base64 编码)

```go
expires := time.Now().Add(10 * time.Minute)
taskID, err := client.SendTask(ctx, "tasks.temporary_task", &celery.TaskOptions{
    Args:    []interface{}{"data"},
    Expires: &expires,
})
```

### 6. 发送到指定队列 (默认 Base64 编码)

```go
taskID, err := client.SendTaskToQueue(ctx, "tasks.priority_task", "high_priority",
    []interface{}{"urgent"},
    map[string]interface{}{"priority": "high"},
)
```

## Python Worker 配置

为了与此 Go 客户端兼容,Python Celery worker 必须配置为使用**协议版本 1**和 **JSON 序列化**:

```python
from celery import Celery

app = Celery(
    'tasks',
    broker='redis://localhost:6379/0',
    backend='redis://localhost:6379/0'
)

# 重要配置
app.conf.update(
    task_serializer='json',
    accept_content=['json'],
    result_serializer='json',
    enable_utc=True,
    task_protocol=1,  # 必须设置为协议版本 1
)

@app.task(name='tasks.add')
def add(x, y):
    return x + y
```

运行 worker:

```bash
celery -A tasks worker --loglevel=info
```

## 项目结构

```
celery-go-client/
├── broker.go           # Broker 接口定义
├── client.go           # Celery 客户端实现
├── message.go          # 消息结构定义
├── redis_broker.go     # Redis Broker 实现
├── amqp_broker.go      # AMQP Broker 实现
├── message_test.go     # 单元测试
├── cmd/                # 示例程序
│   ├── redis_example/
│   │   └── main.go
│   ├── amqp_example/
│   │   └── main.go
│   └── python_worker.py
├── go.mod
├── go.sum
├── Makefile
├── LICENSE
└── README.md
```

## API 文档

### Client

#### `NewClient(config ClientConfig) *Client`

创建新的 Celery 客户端。

**参数:**
- `config.Broker`: Broker 实现 (必需)
- `config.Queue`: 默认队列名称 (默认: "celery")
- `config.Exchange`: 默认交换机名称 (默认: "celery")

#### `SendTask(ctx context.Context, taskName string, options *TaskOptions) (string, error)`

发送任务到 Celery worker。

**参数:**
- `ctx`: 上下文
- `taskName`: 任务名称
- `options`: 任务选项

**返回:** 任务 ID 和错误

#### `SendTaskWithArgs(ctx context.Context, taskName string, args ...interface{}) (string, error)`

便捷方法,发送带位置参数的任务。

#### `SendTaskWithKwargs(ctx context.Context, taskName string, kwargs map[string]interface{}) (string, error)`

便捷方法,发送带关键字参数的任务。

#### `SendTaskToQueue(ctx context.Context, taskName, queue string, args []interface{}, kwargs map[string]interface{}) (string, error)`

发送任务到指定队列。

#### `Close() error`

关闭客户端和 broker 连接。

### TaskOptions

```go
type TaskOptions struct {
    Queue    string                 // 覆盖默认队列
    Exchange string                 // 覆盖默认交换机
    ETA      *time.Time             // 预计执行时间
    Expires  *time.Time             // 任务过期时间
    Args     []interface{}          // 位置参数
    Kwargs   map[string]interface{} // 关键字参数
}
```

### Redis Broker

#### `NewRedisBroker(config RedisBrokerConfig) (*RedisBroker, error)`

创建 Redis broker。

**配置:**
- `Addr`: Redis 服务器地址 (例如: "localhost:6379")
- `Password`: Redis 密码 (可选)
- `DB`: Redis 数据库编号 (默认: 0)
- `Queue`: 默认队列名称 (默认: "celery")

### AMQP Broker

#### `NewAMQPBroker(config AMQPBrokerConfig) (*AMQPBroker, error)`

创建 AMQP (RabbitMQ) broker。

**配置:**
- `URL`: AMQP 连接 URL (例如: "amqp://guest:guest@localhost:5672/")
- `Exchange`: 交换机名称 (默认: "celery")
- `Queue`: 队列名称 (默认: "celery")

## 协议说明

默认情况下,客户端使用 Base64 编码的 JSON 消息体,以确保与标准 Celery worker 的兼容性。

### 原始 JSON 消息体 (Raw JSON Body)

当 `ClientConfig.UseRawJSONBody` 设置为 `true` 时,客户端将直接发送 TaskMessage 的 JSON 字符串作为消息体,此时消息头将变为:

- **CeleryMessage.Body**: TaskMessage 的原始 JSON 字符串
- **CeleryMessage.ContentType**: `application/json`
- **CeleryMessage.Properties.BodyEncoding**: `utf-8` (表示消息体是 utf-8 编码的原始 JSON)

这种模式通常用于 worker 端配置为接受原始 JSON 消息体的场景。

### Base64 编码消息格式

本客户端实现了 **Celery 协议版本 1**,消息格式如下:

本客户端实现了 **Celery 协议版本 1**,消息格式如下:

### 外层消息 (CeleryMessage, Base64 编码)

```json
{
  "body": "<base64-encoded-task-message>",
  "content-type": "application/json",
  "content-encoding": "utf-8",
  "properties": {
    "body_encoding": "base64",
    "correlation_id": "<uuid>",
    "reply_to": "<uuid>",
    "delivery_info": {
      "priority": 0,
      "routing_key": "celery",
      "exchange": "celery"
    },
    "delivery_mode": 2,
    "delivery_tag": "<uuid>"
  }
}
```

### 内层消息 (TaskMessage, Base64 解码后)

```json
{
  "id": "<task-uuid>",
  "task": "tasks.add",
  "args": [10, 20],
  "kwargs": {},
  "retries": 0,
  "eta": "2024-01-01T12:00:00Z"
}
```

## 测试

运行单元测试:

```bash
go test -v
```

运行测试覆盖率:

```bash
go test -cover
```

## 依赖

- `github.com/google/uuid` - UUID 生成
- `github.com/redis/go-redis/v9` - Redis 客户端
- `github.com/rabbitmq/amqp091-go` - AMQP 客户端

## 注意事项

1. **协议版本**: 本客户端仅支持 Celery 协议版本 1。Python Celery 4.0+ 默认使用版本 2,必须显式设置 `task_protocol=1`。

2. **序列化格式**: 仅支持 JSON 序列化,不支持 pickle。

3. **结果后端**: 本客户端仅实现任务发送功能,不包含结果获取功能。如需获取任务结果,请使用 Python 客户端或直接从结果后端读取。

4. **连接管理**: 建议在应用生命周期内复用 Client 实例,避免频繁创建和销毁连接。

## 许可证

MIT License

## 贡献

欢迎提交 Issue 和 Pull Request!

## 参考资料

- [Celery 官方文档](https://docs.celeryproject.org/)
- [Celery 协议规范](https://docs.celeryproject.org/en/stable/internals/protocol.html)
- [AMQP 协议](https://www.rabbitmq.com/tutorials/amqp-concepts.html)
