# Bug 是什么

幂等键重复写入时，存储层用 `%v` 包装唯一约束错误，errors.Is 无法识别为冲突。

# 如何触发

用同一个幂等键第二次调用 Put。

# 错误信息

`insert idempotency key: constraint failed: UNIQUE constraint failed: idempotency_keys.tenant_id, idempotency_keys.key`
