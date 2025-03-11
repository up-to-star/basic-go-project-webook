local key1 = KEYS[1]
local key2 = KEYS[2]

local field1 = ARGV[1]
local field2 = ARGV[2]
local delta = tonumber(ARGV[3])
local exist1=redis.call("EXISTS", key1)
local exist2=redis.call("EXISTS", key2)
if exist1 == 1 and exist2 == 1 then
    redis.call("HINCRBY", key1, field1, delta)
    redis.call("HINCRBY", key2, field2, delta)
    return 1
else
    return 0
end