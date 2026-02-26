# turbopool

一句话：**轻量 goroutine 池**，用更少的分配与 GC 压力跑更多任务。


**✨ 特性**

- 支持泛型任务 `Pool[T]` 与函数任务 `PoolWithFunc`
- 支持阻塞/非阻塞提交与最大阻塞数控制
- 支持空闲 worker 过期清理
- 支持 panic 处理器与自定义日志
- 提供运行指标：容量、空闲、运行中、等待数


**📦 安装**

```bash
go get github.com/gaohao-creator/turbopool
```


**🚀 示例 1：函数任务池（最简）**

```go
package main

import (
	"fmt"
	"time"

	"github.com/gaohao-creator/turbopool"
)

func main() {
	pool, _ := turbopool.NewPoolWithFuncDefaultHandler(
		5,
		turbopool.WithExpiryDuration(10*time.Second),
	)
	defer pool.Release()

	for i := 0; i < 10; i++ {
		j := i
		_ = pool.Submit(func() {
			fmt.Println(j)
		})
	}
}
```


**🚀 示例 2：泛型任务池（最简）**

```go
package main

import (
	"fmt"
	"time"

	"github.com/gaohao-creator/turbopool"
)

func main() {
	pool, _ := turbopool.NewPoolDefaultWorkers(
		5,
		func(task int) {
			fmt.Println(task)
		},
		turbopool.WithExpiryDuration(10*time.Second),
	)
	defer pool.Release()

	for i := 0; i < 10; i++ {
		_ = pool.Submit(i)
	}
}
```

```go
package main

import (
	"fmt"
	"time"

	"github.com/gaohao-creator/turbopool"
)

func main() {
	pool, _ := turbopool.NewPoolDefaultWorkers(
		5,
		func(task func()) {
			task()
		},
		turbopool.WithExpiryDuration(10*time.Second),
	)
	defer pool.Release()

	for i := 0; i < 10; i++ {
		_ = pool.Submit(func() {
			fmt.Println(i)
		})
	}
}
```


**🧭 API 速查**

- 构造（函数池）：`NewPoolWithFunc` / `NewPoolWithFuncDefaultWorkers` / `NewPoolWithFuncDefaultHandler`
- 构造（泛型池）：`NewPool` / `NewPoolDefaultWorkers` / `NewPoolDefaultHandler`
- 提交任务：`Submit`
- 释放资源：`Release` / `ReleaseWithWait` / `ReleaseWithTimeout`
- 等待任务完成：`Wait`
- 监控指标：`Cap` / `Free` / `Running` / `Waiting`
- 生命周期：`Open` / `Close` / `Opened` / `Closed`


**⚙️ 常用 Options**

- `WithNonblocking(bool)`：无空闲 worker 时直接失败
- `WithMaxBlockingTasks(int)`：阻塞提交的最大等待数
- `WithExpiryDuration(time.Duration)`：空闲 worker 过期清理
- `WithPanicHandler(func(any))`：自定义 panic 处理
- `WithLogger(Logger)`：自定义日志


**📊 性能对比**

同样的并发上限下，turbopool 的内存与分配次数明显更低，耗时与 Channel/ErrGroup 接近。

| 测试项 | 耗时 (ms/op) | 内存 (MB/op) | 分配次数 (allocs/op) |
|----------------|-----------:|------------:|---------------------:|
| Goroutine（直接） | 341.76 | 129.32 | 2,011,591 |
| Channel（信号量） | 561.45 | 144.15 | 2,001,581 |
| ErrGroup（SetLimit） | 562.36 | 152.77 | 3,008,049 |
| TurboPool | 600.07 | 37.74 | 1,081,410 |

- 基准命令：`go test -bench "^Benchmark(Goroutine|Channel|ErrGroup|TurboPool_Sleep)$" -benchmem -run "^$"`
- 任务模型：`sleep 10ms`，`RunTimes = 1e6`，`PoolCap = 5e4`，`Expiry = 10s`
- 环境：`windows/amd64`，`i5-13400`，`Go 1.24.12`
- 说明：`Goroutine（直接）` 为不限制并发，耗时仅供参考


**✅ 兼容性**

- 需要 Go 1.24+（详见 `go.mod`）
