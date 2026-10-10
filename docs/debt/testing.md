# 债务登记：测试纪律

> 本文件是 `docs/architecture-rules.md` 第 8 节债务登记表的「测试纪律」分册。
> 通用机制、条目字段含义与维护流程见 [`README.md`](./README.md)。

## 1. 测试替身内嵌空接口的运行时脆性

- **状态**：存量约定，三处实证。
- **基线**：service 层 fake 以内嵌接口继承全部方法，接口新增方法被既有测试路径调用时以 nil panic 暴露而非编译错误；已实证三例：`recordedTokenService.GenerateTokenPair`、`loginUserRepo.Update`、`lockoutUserRepo.GetByUsername`。
- **验证命令**：`go test ./internal/... -count=1`。
- **红线**：接口新增方法被既有测试路径触达时，必须为受影响 fake 显式覆写，禁止依赖内嵌空接口的静默兼容。
