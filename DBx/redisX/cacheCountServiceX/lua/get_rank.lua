-- 接收 KEYS[1] 为排行榜键，ARGV[1] 和 ARGV[2] 为排名范围。返回一个数组。
local rank_key = KEYS[1]
local start_idx = tonumber(ARGV[1])
local end_idx = tonumber(ARGV[2])

if not start_idx or not end_idx then
    return {}
end
-- 获取排行榜
local items = redis.call('ZREVRANGE', rank_key, start_idx, end_idx, 'WITHSCORES')
return items