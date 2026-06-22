import type { Plugin } from '@opencode-ai/plugin';

const touchedMemory = (event: unknown): boolean => {
  const body = JSON.stringify(event ?? {});
  return body.includes('.lufy/memory/');
};

export const LufyMemoryContextPlugin: Plugin = async ({ $, directory, worktree }) => {
  const root = worktree || directory;
  let oriented = false;

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
      if (event?.type === 'session.created') {
        await orient();
        return;
      }
      if (event?.type === 'file.edited' && touchedMemory(event)) {
        await validateMemory();
      }
    },
  };
};
