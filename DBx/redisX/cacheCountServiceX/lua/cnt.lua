-- 具体业务，文章、帖子等
local key = KEYS[1]
-- 阅读数，点赞数还是收藏数
local cntKey = ARGV[1]
-- 增量
local delta = tonumber(ARGV[2])
-- key是否存在
local exist = redis.call('EXISTS', key)
if exist == 1 then
    redis.call('HINCRBY',key,cntKey,delta)
    return 1
else
    return 0
end