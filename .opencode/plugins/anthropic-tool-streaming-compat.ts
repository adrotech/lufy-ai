import type { Plugin } from "@opencode-ai/plugin";

const ANTHROPIC_COMPATIBLE_SDKS = new Set([
  "@ai-sdk/anthropic",
  "@ai-sdk/google-vertex/anthropic",
]);

export const AnthropicToolStreamingCompat: Plugin = async () => ({
  "chat.params": async (input, output) => {
    if (!ANTHROPIC_COMPATIBLE_SDKS.has(input.model.api.npm)) return;

    output.options.toolStreaming = false;
  },
});
