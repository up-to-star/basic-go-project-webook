-- redis 存的key
local key = KEYS[1]
-- 验证次数，记录还可以验证几次
local cntKey = key..":cnt"
-- 验证码
local val = ARGV[1]
-- 过期时间
local ttl = tonumber(redis.call("ttl", key))

if ttl == -1 then
    -- key 存在，但没有过期时间, 就是系统错误
    return -2
    -- ttl = -2 是 key 不存在，< 540 是已经过了一分钟了
elseif ttl == -2 or ttl < 540  then
    redis.call("set", key, val)
    redis.call("expire", key, 600)
    redis.call("set", cntKey, 3)
    redis.call("expire", cntKey, 600)
    -- 代表正常
    return 0
else
    -- 发送太频繁
    return -1
end