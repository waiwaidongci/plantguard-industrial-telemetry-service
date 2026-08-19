# Bug 是什么

查询不存在的设备型号时，错误链被断掉并统一映射成内部错误，接口返回 500 而不是 404。

# 如何触发

对不存在的 device model id 调用设备型号查询接口。

# 错误信息

`internal_error: an unexpected error occurred`
