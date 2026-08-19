# Bug 是什么

维护任务状态机缺少 retrying 到 completed/cancelled 的合法转换，重试中的任务无法完成。

# 如何触发

将任务置于 retrying 状态后尝试转换到 completed。

# 错误信息

`invalid_task_transition: maintenance task cannot transition to the requested status`
