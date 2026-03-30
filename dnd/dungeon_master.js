import { createOpencode } from "@opencode-ai/sdk"

const opencode = await createOpencode({
  hostname: "127.0.0.1",
  port: 4096,
  config: {
        model: "opencode/big-pickle",
        agent: 'dnd'
  },
})

console.log(`Server running at ${opencode.server.url}`)
