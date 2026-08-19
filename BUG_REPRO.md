# Bug 是什么

事件确认服务用无锁 map 缓存已确认事件，两个 goroutine 并发确认时发生数据竞争，可能重复确认。

# 如何触发

两个 goroutine 同时确认同一个事件。

# 错误信息

`WARNING: DATA RACE`，同一事件出现两次成功确认。
