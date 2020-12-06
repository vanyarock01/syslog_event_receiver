box.cfg{
    vinyl_memory = 512 * 1024 * 1024,
    vinyl_dir = '/var/lib/tarantool',
    wal_dir = '/var/lib/tarantool',
    listen = 3301,
    username = 'gouser',
}

local s = box.schema.space.create('syslog', {
    engine = 'vinyl',
    format = {
        {name = 'id',        type = 'unsigned'},
        {name = 'timestamp', type = 'number'},
        {name = 'message',   type = 'string'},
        {name = 'hostname',  type = 'string', is_nullable = true},
        {name = 'priority',  type = 'number', is_nullable = true},
        {name = 'program',   type = 'string', is_nullable = true},
        {name = 'pid',       type = 'number', is_nullable = true},
        {name = 'sequence',  type = 'number', is_nullable = true},
    },
    if_not_exists = true,
})

local seq = box.schema.sequence.create('id', {if_not_exists = true})

s:create_index('primary', {
    sequence = seq,
    type = 'TREE',
    unique = true,
    parts = {
        {field = 1, type = 'unsigned'},
    },
    if_not_exists = true,
})

s:create_index('timescan', {
    type = 'TREE',
    unique = false,
    parts = {
        {field = 2, type = 'number'},
    },
    if_not_exists = true,
})

box.schema.user.create('gouser', {password = 'secret', if_not_exists = true})
box.schema.user.grant('gouser', 'read,write,execute', 'universe', nil, {if_not_exists = true})
box.schema.user.revoke('guest', 'write,execute', 'universe', nil, {if_exist = true})
