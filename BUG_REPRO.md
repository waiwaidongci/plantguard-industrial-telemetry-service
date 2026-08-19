# Bug 是什么

worker 聚合函数在读取遥测摘要时原地写回输入切片，并发聚合同一份切片发生数据竞争。

# 如何触发

两个 goroutine 同时聚合同一份遥测摘要切片。

# 错误信息

`WARNING: DATA RACE` at `internal/worker.aggregateSummaries()`
