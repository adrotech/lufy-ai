import type { Plugin } from '@opencode-ai/plugin';

const touchedMemory = (event: unknown): boolean => {
  const body = JSON.stringify(event ?? {});
  return body.includes('.lufy/memory/');
};

const REQUIRED_DIAGNOSTICS = 'memory_provider_used context_graph_status context_graph_queries fallback_reason generic_discovery_before_graph';

const broadDiscovery = (body: string): boolean => {
  return /\b(glob|grep|find|rg)\b/.test(body) || body.includes('exploratory read') || body.includes('generic discovery');
};

const contextPreflight = (body: string): boolean => {
  return body.includes('lufy-ai context status') || body.includes('lufy-ai context query') || body.includes('context_graph_status');
};

const externalMemoryAsProject = (body: string): boolean => {
  return /(Engram|MCP|mem_search|engram_mem)/.test(body) && body.includes('project memory') && !body.includes('fallback_reason');
};

export const LufyMemoryContextPlugin: Plugin = async ({ $, directory, worktree }) => {
  const root = worktree || directory;
  let oriented = false;
  let sawContextPreflight = false;
  let warnedDiscovery = false;
  let warnedExternalMemory = false;

  const orient = async () => {
    if (oriented) return;
    oriented = true;
    await $`LUFY_PROJECT_ROOT=${root} bash ${root}/.opencode/hooks/memory-orient.sh`.quiet().nothrow();
  };

  const validateMemory = async () => {
    await $`LUFY_PROJECT_ROOT=${root} bash ${root}/.opencode/hooks/memory-validate.sh`.quiet().nothrow();
  };

  return {
    event: async ({ event }) => {
      const body = JSON.stringify(event ?? {});
      if (event?.type === 'session.created') {
        await orient();
        return;
      }
      if (contextPreflight(body)) {
        sawContextPreflight = true;
      }
      if (!sawContextPreflight && !warnedDiscovery && broadDiscovery(body)) {
        warnedDiscovery = true;
        console.warn(`[lufy] context_graph.enabled=true requires lufy-ai context status/query before broad generic discovery, unless direct-path or evidenced fallback applies. Record ${REQUIRED_DIAGNOSTICS}.`);
      }
      if (!warnedExternalMemory && externalMemoryAsProject(body)) {
        warnedExternalMemory = true;
        console.warn(`[lufy] memory.provider=obsidian requires Obsidian project memory first; Engram/MCP must be fallback or non-project memory with ${REQUIRED_DIAGNOSTICS}.`);
      }
      if (event?.type === 'file.edited' && touchedMemory(event)) {
        await validateMemory();
      }
    },
  };
};
