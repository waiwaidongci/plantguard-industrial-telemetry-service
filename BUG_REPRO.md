# Bug 是什么

规则动作规范化原地复用输入切片，调用方保留的动作列表被污染。

# 如何触发

更新规则并传入动作列表，然后重新读取调用方保留的原始列表。

# 错误信息

`input actions were mutated: [log webhook invalid maintenance] -> [log webhook maintenance maintenance]`
