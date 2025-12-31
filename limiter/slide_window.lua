-- Redis滑动窗口限流Lua脚本
-- 修复版本：解决member唯一性和统计逻辑问题

-- KEYS[1]: 限流key
-- ARGV[1]: 窗口大小（毫秒）
-- ARGV[2]: 阈值（窗口内允许的最大请求数）
-- ARGV[3]: 当前时间戳（毫秒）
-- ARGV[4]: 唯一请求ID（可选，用于生成唯一member）

local key = KEYS[1]
local window = tonumber(ARGV[1])          -- 窗口大小（毫秒）
local threshold = tonumber(ARGV[2])       -- 阈值
local now = tonumber(ARGV[3])             -- 当前时间戳（毫秒）
local requestId = ARGV[4]                 -- 唯一请求ID（可选）

-- 计算窗口开始时间戳
local windowStart = now - window

-- 1. 清理窗口外的过期数据
-- 删除所有score小于窗口开始时间的数据
redis.call('ZREMRANGEBYSCORE', key, '-inf', windowStart)

-- 2. 统计当前窗口内的请求数量
-- 注意：使用windowStart到'+inf'的范围，只统计窗口内的数据
local cnt = redis.call('ZCOUNT', key, windowStart, '+inf')

-- 3. 检查是否超过阈值
if cnt >= threshold then
    -- 触发限流，返回1
    return 1
end

-- 4. 生成唯一的member标识
-- 如果提供了requestId，则使用它；否则生成一个基于时间的唯一标识
local member
if requestId and requestId ~= '' then
    member = requestId
else
    -- 使用时间戳+微秒时间生成相对唯一的member
    -- 这里使用redis的TIME命令获取当前时间的秒和微秒部分
    local timeInfo = redis.call('TIME')
    local seconds = tonumber(timeInfo[1])
    local microseconds = tonumber(timeInfo[2])
    member = string.format("%s:%s:%s", now, seconds, microseconds)
end

-- 5. 添加当前请求到滑动窗口
-- score使用当前时间戳，member使用唯一标识
redis.call('ZADD', key, now, member)

-- 6. 设置key的过期时间，避免内存泄漏
-- 过期时间设置为窗口大小的2倍，确保即使没有新请求也能自动清理
redis.call('PEXPIRE', key, window * 2)

-- 7. 返回0表示通过限流
return 0
