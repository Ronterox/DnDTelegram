import { createOpencode } from "@opencode-ai/sdk";

const opencode = await createOpencode({
  hostname: "0.0.0.0",
  port: 4096,
  pure: true, // Run without external plugins that might interfere
  config: {
    model: "opencode/big-pickle",
  } as any,
});

console.log(`Server running at http://0.0.0.0:4096`);
