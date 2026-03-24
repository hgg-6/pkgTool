# 贡献指南

感谢您对 pkg_tool 项目的兴趣！我们欢迎各种形式的贡献，包括但不限于代码提交、功能请求、问题反馈和文档改进。

## 行为准则

- 保持友好和尊重的沟通态度
- 对不同的观点和意见保持开放态度
- 聚焦于对项目最有益的方向

## 开发环境设置

### 前置条件

- Go 1.21 或更高版本
- Git

### 克隆与依赖安装

```bash
git clone https://gitee.com/hgg_test/pkg_tool.git
cd pkg_tool
go mod tidy
```

### 运行测试

```bash
# 运行所有测试
go test ./...

# 运行测试并查看覆盖率
go test -cover ./...

# 运行特定包的测试
go test -v -run TestFunctionName ./path/to/package
```

## 分支管理

- `master`: 稳定分支，包含已发布的代码
- 功能开发请创建新的分支，命名规范：`feature/功能名称` 或 `fix/问题描述`

```bash
git checkout -b feature/新功能
```

## 提交规范

### Commit Message 格式

```
<type>: <subject>

<body>
```

**Type 类型：**

| 类型 | 说明 |
|------|------|
| feat | 新功能 |
| fix | 错误修复 |
| docs | 文档变更 |
| style | 代码格式（不影响功能） |
| refactor | 重构 |
| test | 测试相关 |
| chore | 构建/工具变更 |

**示例：**

```
feat: 添加 Redis ZSet 排行榜服务

- 实现基于 Redis ZSet 的实时排行榜
- 支持分数排名和成员查询
- 添加批量操作接口
```

## 代码规范

### Go 代码风格

- 遵循 Go 官方代码格式化工具 `go fmt`
- 变量命名应具有描述性，避免缩写
- 公开接口需要有完整的文档注释
- 错误处理：优先使用有意义的错误信息

```go
// 正确示例
func (s *Service) GetUser(ctx context.Context, id string) (*User, error) {
    user, err := s.repo.FindByID(ctx, id)
    if err != nil {
        return nil, fmt.Errorf("GetUser: %w", err)
    }
    return user, nil
}

// 避免
func (s *Service) GetUser(ctx context.Context, id string) (*User, error) {
    return s.repo.FindByID(ctx, id)
}
```

### 接口定义

新模块应定义接口并在 `types.go` 中声明，示例：

```go
// types.go
type ServiceInterface interface {
    Method1(param ParamType) (ResultType, error)
    Method2(ctx context.Context, param ParamType) (ResultType, error)
}
```

### Mock 生成

如需为接口生成 Mock 用于测试：

```bash
go generate ./...
```

## 模块开发规范

### 目录结构

每个功能模块应包含：

```
moduleX/
├── types.go          # 类型定义和接口
├── module.go         # 主要实现
├── module_test.go    # 单元测试
├── mock_*.go         # Mock 实现（由 mockgen 生成）
└── help_doc.txt      # 辅助文档（可选）
```

### 新模块检查清单

- [ ] 在 `types.go` 中定义核心接口
- [ ] 实现接口并添加构造函数
- [ ] 编写单元测试，覆盖率建议 >60%
- [ ] 更新 README.md 添加使用示例
- [ ] 运行 `go fmt` 和 `go vet` 确保代码质量

## Pull Request 流程

1. **Fork** 本仓库，创建功能分支
2. **开发** 并确保所有测试通过
3. **提交** 代码，遵循 Commit Message 规范
4. **推送** 到您的 Fork 仓库
5. **创建 Pull Request**，描述变更内容

### PR 描述模板

```markdown
## 变更描述
简要说明本次变更的内容和目的。

## 变更类型
- [ ] 新功能 (feat)
- [ ] 错误修复 (fix)
- [ ] 文档更新 (docs)
- [ ] 代码重构 (refactor)
- [ ] 测试相关 (test)

## 影响范围
说明本次变更影响哪些模块或功能。

## 测试情况
- [ ] 已添加单元测试
- [ ] 手动测试验证通过
- [ ] 不需要测试（仅文档变更）
```

## 问题反馈

提交 Issue 时请包含：

- 问题描述和复现步骤
- 您的环境信息（Go 版本、操作系统等）
- 相关的错误日志或截图
- 您认为可能的解决方案

## 许可证

通过贡献代码，您同意将您的作品按照项目的 MIT 许可证发布。