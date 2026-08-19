# Bug 是什么

HTTP 中间件链在两个位置丢失 context deadline，下游 handler 看不到超时边界。

# 如何触发

经过 Timeout 或 RequestMetrics 后，在下游 handler 检查 deadline。

# 错误信息

`downstream context has no deadline` / `request context deadline was lost`
