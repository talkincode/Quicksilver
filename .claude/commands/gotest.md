---
description: "Go 项目：补齐测试、自动执行与迭代修正，直到全部通过并满足覆盖率目标"
argument-hint: "可选参数：TARGET=./pkg/foo COV_TARGET=85 MAX_ITERS=5"
allowed-tools: ["Read", "Edit", "Run"]
model: claude-4-5-sonnet
---

## 目标

针对 Go 项目自动化测试改进：

- 补齐关键测试用例；
- 运行测试；
- 若失败，最小化修正；
- 直到所有测试通过并达到目标覆盖率。

## 默认参数

- `TARGET=${TARGET:=./...}`
- `COV_TARGET=${COV_TARGET:=80}`
- `MAX_ITERS=${MAX_ITERS:=5}`
- `TEST_CMD=${TEST_CMD:=go test ${TARGET} -count=1 -race -covermode=atomic -coverprofile=coverage.out}`

## 工作流程

**每轮执行以下步骤：**

1️⃣ **检测缺口**

- 若存在 `coverage.out`，执行：  
  `!go tool cover -func=coverage.out`  
  Claude 识别覆盖率最低的函数和文件，确定补测优先级。
- 若无覆盖报告，则从源码结构与 `_test.go` 文件中推断缺测区域。

2️⃣ **补充测试**

- 采用表格驱动测试（table-driven test）。
- 优先覆盖：导出函数、错误/边界分支、并发或异常路径。
- 尽量不改动实现，必要时添加轻量 mock/stub。

3️⃣ **运行测试**

- 执行：`!bash -lc "$TEST_CMD"`
- 若失败，进入第 4 步修复逻辑。

4️⃣ **修复并重跑**

- 优先修正测试断言或依赖；
- 若确为实现缺陷，最小化修改并补测试；
- 回到第 3 步重新运行。

5️⃣ **覆盖率检查**

- 执行：`!go tool cover -func=coverage.out`
- 读取总覆盖率（total），若低于 `$COV_TARGET`，继续迭代，否则结束。

## 输出

- 新增或改进的测试点列表
- 当前总覆盖率
- 测试修正摘要
- 后续补测建议（未覆盖的关键区域）
