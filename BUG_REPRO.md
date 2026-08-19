# Bug 是什么

通知发送服务用命名返回值和 defer 吞掉 MarkSent 错误，接口误报成功。

# 如何触发

让通知仓储的 MarkSent 返回错误，再调用发送接口。

# 错误信息

`expected MarkSent error to be preserved, got <nil>`
