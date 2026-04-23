import { createOpencode } from "@opencode-ai/sdk";

interface ServerUrl {
  url: string;
}

const opencode = await createOpencode({
  hostname: "0.0.0.0",
  port: 4096,
  config: {
    model: "opencode/big-pickle",
    default_agent: "dnd"
  } as never,
});

const serverUrl: string = (opencode.server as ServerUrl).url;
console.log(`Server running at ${serverUrl}`);
