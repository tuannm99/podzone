local typedefs = require "kong.db.schema.typedefs"

return {
    name = "tenant-identifier",
    fields = {
        { consumer = typedefs.no_consumer },
        {
            config = {
                type = "record",
                fields = {
                    { required = { type = "boolean", default = true } },
                },
            }
        },
    },
}
