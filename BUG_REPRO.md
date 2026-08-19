# Bug 是什么

创建站点时如果名称是空字符串，校验逻辑向 nil map 写入，导致进程 panic。

# 如何触发

提交 `{"name":"   "}` 的站点创建请求。

# 错误信息

`panic: assignment to entry in nil map`
