export type UsageTokens = {
  total?: number;
  input?: number;
  output?: number;
  reasoning?: number;
};

export type NormalizedTokens = {
  total: number;
  input: number;
  output: number;
  reasoning: number;
  cacheRead: number;
  cacheWrite: number;
};

export type MessageLike = {
  id: string;
  sessionID: string;
  role: 'user' | 'assistant';
  time?: { created?: number; completed?: number };
  agent?: string;
  modelID?: string;
  cost?: number;
  tokens?: UsageTokens;
  error?: unknown;
};

export type SessionLike = {
  id: string;
  parentID?: string;
  title?: string;
  time?: { created?: number; updated?: number };
};

export type ToolActivity = {
  id: string;
  sessionID: string;
  messageID: string;
  callID: string;
  tool: string;
  status: 'pending' | 'running' | 'completed' | 'error';
  title?: string;
  error?: string;
  startedAt?: number;
  endedAt?: number;
  durationMs?: number;
};

export type AgentRuntimeState = 'busy' | 'idle' | 'error' | 'done';

export type AgentCatalogEntry = {
  name: string;
  description?: string;
  mode?: 'subagent' | 'primary' | 'all';
  color?: string;
  hidden?: boolean;
};

export type AvailableAgent = AgentCatalogEntry & { active: boolean };

export type AgentUsage = {
  sessionID: string;
  source: 'current-session' | 'child-session';
  name: string;
  title?: string;
  state: AgentRuntimeState;
  errorReason?: string;
  tokens: NormalizedTokens;
  cost?: number;
  tools: ToolActivity[];
  modelID?: string;
  startedAt?: number;
  updatedAt?: number;
  durationMs?: number;
  compacted: boolean;
};

export type ObservatorySnapshot = {
  sessionID: string;
  generatedAt: number;
  root: AgentUsage;
  availableAgents: AvailableAgent[];
  agents: AgentUsage[];
  activeAgents: number;
  totalTokens: NormalizedTokens;
  totalCost?: number;
  tools: ToolActivity[];
  compacted: boolean;
};

export type ObservatoryState = {
  sessions: Record<string, SessionLike>;
  messages: Record<string, MessageLike>;
  messageOrder: string[];
  statuses: Record<string, { type: string } | undefined>;
  errors: Record<string, string | undefined>;
  compacted: Record<string, boolean>;
};

export type ObservatoryEvent =
  | { type: 'session.created' | 'session.updated'; session: SessionLike }
  | { type: 'session.deleted'; sessionID: string }
  | { type: 'session.status'; sessionID: string; status: { type: string } }
  | { type: 'message.updated'; message: MessageLike };

const ZERO_TOKENS: NormalizedTokens = { total: 0, input: 0, output: 0, reasoning: 0, cacheRead: 0, cacheWrite: 0 };

export function createObservatoryState(): ObservatoryState {
  return { sessions: {}, messages: {}, messageOrder: [], statuses: {}, errors: {}, compacted: {} };
}

export function reduceObservatoryState(state: ObservatoryState, event: ObservatoryEvent): ObservatoryState {
  switch (event.type) {
    case 'session.created':
    case 'session.updated':
      return { ...state, sessions: { ...state.sessions, [event.session.id]: event.session } };
    case 'session.status':
      return { ...state, statuses: { ...state.statuses, [event.sessionID]: event.status } };
    case 'message.updated':
      return {
        ...state,
        messages: { ...state.messages, [event.message.id]: event.message },
        messageOrder: state.messages[event.message.id] ? state.messageOrder : [...state.messageOrder, event.message.id],
      };
  }
  return state;
}

export function selectObservatorySnapshot(
  state: ObservatoryState,
  sessionID: string,
  options: { availableAgents?: readonly AgentCatalogEntry[]; defaultAgent?: string; maxAgentTools?: number; maxTools?: number }
): ObservatorySnapshot {
  const root: AgentUsage = {
    sessionID,
    source: 'current-session',
    name: options.defaultAgent || 'Main',
    state: 'idle',
    tokens: ZERO_TOKENS,
    tools: [],
    compacted: false,
  };

  const agents = Object.values(state.sessions)
    .filter(s => s.parentID === sessionID)
    .map(s => ({
      sessionID: s.id,
      source: 'child-session' as const,
      name: s.title || 'subagent',
      state: 'done' as const,
      tokens: ZERO_TOKENS,
      tools: [],
      compacted: false,
    }));

  return {
    sessionID,
    generatedAt: Date.now(),
    root,
    availableAgents: (options.availableAgents || []).map(a => ({ ...a, active: a.name === root.name })),
    agents,
    activeAgents: agents.filter(a => a.state === 'busy').length,
    totalTokens: ZERO_TOKENS,
    tools: [],
    compacted: false,
  };
}

export function normalizeTokens(tokens?: UsageTokens): NormalizedTokens {
  return {
    total: tokens?.total || 0,
    input: tokens?.input || 0,
    output: tokens?.output || 0,
    reasoning: tokens?.reasoning || 0,
    cacheRead: 0,
    cacheWrite: 0,
  };
}

export function formatTokenCount(tokens: number): string {
  if (tokens >= 1_000_000) return `${(tokens / 1_000_000).toFixed(1)}M`;
  if (tokens >= 1_000) return `${(tokens / 1_000).toFixed(1)}k`;
  return `${Math.round(tokens)}`;
}

export function formatCost(cost: number | undefined): string {
  if (typeof cost !== 'number') return '--';
  return cost === 0 ? '$0' : `$${cost < 0.01 ? cost.toFixed(4) : cost.toFixed(2)}`;
}

export function formatDuration(durationMs: number | undefined): string {
  if (typeof durationMs !== 'number' || durationMs < 0) return '--';
  if (durationMs < 1000) return `${Math.round(durationMs)}ms`;
  const seconds = Math.round(durationMs / 1000);
  if (seconds < 60) return `${seconds}s`;
  const minutes = Math.floor(seconds / 60);
  return `${minutes}m ${seconds % 60}s`;
}