# Bug 是什么

遥测批量写入在调用方取消 context 后仍继续写库，取消没有向下游传播。

# 如何触发

取消上游 context 后调用遥测批量写入接口。

# 错误信息

`expected context.Canceled, got <nil>`
