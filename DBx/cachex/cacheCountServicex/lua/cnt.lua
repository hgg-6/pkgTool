---- 接收 KEYS[1] 为计数器键，KEYS[2] 为排行榜键，ARGV[1] 为变化量。返回新的计数值。
--local cnt_key = KEYS[1]
--local rank_key = KEYS[2]
--local delta = ARGV[1]
--
---- 更新计数
--local new_cnt = redis.call('INCRBY', cnt_key, delta)
--if new_cnt < 0 then
--    new_cnt = 0
--    redis.call('SET', cnt_key, 0)
--end
--
---- 更新排行榜
--if delta ~= '0' then
--    redis.call('ZADD', rank_key, new_cnt, string.sub(cnt_key, -string.find(string.reverse(cnt_key), ':') + 1))
--end
--
---- 设置过期时间（如果需要）
---- 计数过期时间短，5分钟
--redis.call('EXPIRE', cnt_key, 300)
---- 排行榜过期时间更长，1小时，避免缓存击穿
--redis.call('EXPIRE', rank_key, 3600)
--
--return new_cnt





-- KEYS[1] = 计数器 key (e.g., "user:123:cnt")
-- KEYS[2] = 排行榜 key (e.g., "rank:daily")
-- ARGV[1] = delta (string, e.g., "1")
-- ARGV[2] = member (e.g., "123") ← 更安全高效！

local cnt_key = KEYS[1]
local rank_key = KEYS[2]
local delta = tonumber(ARGV[1])
local member = ARGV[2]

-- 更新计数
local new_cnt = redis.call('INCRBY', cnt_key, delta)
if new_cnt < 0 then
    new_cnt = 0
    redis.call('SET', cnt_key, 0)
end

-- 更新排行榜（仅当变化量非零）
if delta ~= 0 then
    redis.call('ZADD', rank_key, new_cnt, member)
end

-- 设置过期时间（滑动窗口场景下合理）
redis.call('EXPIRE', cnt_key, 660)     -- 11分钟
redis.call('EXPIRE', rank_key, 86460)   -- 1天1分

return new_cnt