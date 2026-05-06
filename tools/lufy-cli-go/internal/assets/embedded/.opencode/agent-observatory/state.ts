export type UsageTokens = {
  total?: number;
  input?: number;
  output?: number;
  reasoning?: number;
  cache?: {
    read?: number;
    write?: number;
  };
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
  providerID?: string;
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

type PartTime = { start?: number; end?: number; created?: number };

export type PartLike = {
  id: string;
  sessionID: string;
  messageID: string;
  type: string;
  text?: string;
  prompt?: string;
  description?: string;
  agent?: string;
  command?: string;
  model?: { providerID?: string; modelID?: string };
  callID?: string;
  tool?: string;
  state?: {
    status?: 'pending' | 'running' | 'completed' | 'error';
    title?: string;
    input?: Record<string, unknown>;
    output?: string;
    error?: string;
    metadata?: Record<string, unknown>;
    time?: { start?: number; end?: number; compacted?: number };
  };
  cost?: number;
  tokens?: UsageTokens;
  reason?: string;
  time?: PartTime;
};

export type ToolActivity = {
  id: string;
  sessionID: string;
  messageID: string;
  callID: string;
  tool: string;
  status: 'pending' | 'running' | 'completed' | 'error';
  title?: string;
  input?: string;
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
  objective?: string;
  currentAction?: string;
  lastActivity?: string;
  state: AgentRuntimeState;
  errorReason?: string;
  tokens: NormalizedTokens;
  cost?: number;
  tools: ToolActivity[];
  runningTools: ToolActivity[];
  completedToolCount: number;
  modelID?: string;
  providerID?: string;
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
  parts: Record<string, PartLike>;
  partOrder: string[];
  statuses: Record<string, { type: string } | undefined>;
  errors: Record<string, string | undefined>;
  compacted: Record<string, boolean>;
};

export type ObservatoryEvent =
  | { type: 'session.created' | 'session.updated'; session: SessionLike }
  | { type: 'session.deleted'; sessionID: string }
  | { type: 'session.status'; sessionID: string; status: { type: string } }
  | { type: 'session.compacted'; sessionID: string }
  | { type: 'message.updated'; message: MessageLike }
  | { type: 'message.part.updated'; part: PartLike }
  | { type: 'message.part.removed'; partID: string };

const ZERO_TOKENS: NormalizedTokens = { total: 0, input: 0, output: 0, reasoning: 0, cacheRead: 0, cacheWrite: 0 };

export function createObservatoryState(): ObservatoryState {
  return { sessions: {}, messages: {}, messageOrder: [], parts: {}, partOrder: [], statuses: {}, errors: {}, compacted: {} };
}

export function reduceObservatoryState(state: ObservatoryState, event: ObservatoryEvent): ObservatoryState {
  switch (event.type) {
    case 'session.created':
    case 'session.updated':
      return { ...state, sessions: { ...state.sessions, [event.session.id]: event.session } };
    case 'session.deleted': {
      const { [event.sessionID]: _removed, ...sessions } = state.sessions;
      return { ...state, sessions };
    }
    case 'session.status':
      return { ...state, statuses: { ...state.statuses, [event.sessionID]: event.status } };
    case 'session.compacted':
      return { ...state, compacted: { ...state.compacted, [event.sessionID]: true } };
    case 'message.updated':
      return {
        ...state,
        messages: { ...state.messages, [event.message.id]: event.message },
        messageOrder: state.messages[event.message.id] ? state.messageOrder : [...state.messageOrder, event.message.id],
      };
    case 'message.part.updated':
      return {
        ...state,
        parts: { ...state.parts, [event.part.id]: event.part },
        partOrder: state.parts[event.part.id] ? state.partOrder : [...state.partOrder, event.part.id],
      };
    case 'message.part.removed': {
      const { [event.partID]: _removed, ...parts } = state.parts;
      return { ...state, parts, partOrder: state.partOrder.filter(id => id !== event.partID) };
    }
  }
  return state;
}

export function selectObservatorySnapshot(
  state: ObservatoryState,
  sessionID: string,
  options: { availableAgents?: readonly AgentCatalogEntry[]; defaultAgent?: string; maxAgentTools?: number; maxTools?: number }
): ObservatorySnapshot {
  const root = buildUsage({
    sessionID,
    source: 'current-session',
    name: options.defaultAgent || 'Main',
    session: state.sessions[sessionID],
    messages: messagesForSession(state, sessionID),
    parts: partsForSession(state, sessionID),
    status: state.statuses[sessionID],
    compacted: state.compacted[sessionID],
  });

  const agents = Object.values(state.sessions)
    .filter(s => s.parentID === sessionID)
    .sort((a, b) => lastSessionTime(b) - lastSessionTime(a))
    .map((session): AgentUsage => buildUsage({
      sessionID: session.id,
      source: 'child-session',
      name: session.title || 'subagent',
      title: session.title,
      session,
      messages: messagesForSession(state, session.id),
      parts: partsForSession(state, session.id),
      status: state.statuses[session.id],
      compacted: state.compacted[session.id],
    }));

  const totalTokens = [root, ...agents].reduce((total, agent) => addTokens(total, agent.tokens), ZERO_TOKENS);
  const totalCost = [root, ...agents].reduce((total, agent) => total + (agent.cost || 0), 0);
  const allTools = [root, ...agents].flatMap(agent => agent.tools);

  return {
    sessionID,
    generatedAt: Date.now(),
    root,
    availableAgents: (options.availableAgents || []).map(a => ({ ...a, active: a.name === root.name })),
    agents,
    activeAgents: agents.filter(a => a.state === 'busy').length,
    totalTokens,
    totalCost: totalCost || undefined,
    tools: allTools.slice(0, options.maxTools ?? allTools.length),
    compacted: Boolean(state.compacted[sessionID]),
  };
}

function buildUsage(input: {
  sessionID: string;
  source: AgentUsage['source'];
  name: string;
  title?: string;
  session?: SessionLike;
  messages: MessageLike[];
  parts: PartLike[];
  status?: { type: string };
  compacted?: boolean;
}): AgentUsage {
  const assistantMessages = input.messages.filter(message => message.role === 'assistant');
  const allMessages = assistantMessages.length ? assistantMessages : input.messages;
  const messageTokens = allMessages.reduce((total, message) => addTokens(total, normalizeTokens(message.tokens)), ZERO_TOKENS);
  const stepTokens = input.parts
    .filter(part => part.type === 'step-finish')
    .reduce((total, part) => addTokens(total, normalizeTokens(part.tokens)), ZERO_TOKENS);
  const tokens = messageTokens.total > 0 ? messageTokens : stepTokens;
  const messageCost = allMessages.reduce((total, message) => total + (typeof message.cost === 'number' ? message.cost : 0), 0);
  const stepCost = input.parts.filter(part => part.type === 'step-finish').reduce((total, part) => total + (typeof part.cost === 'number' ? part.cost : 0), 0);
  const cost = messageCost || stepCost;
  const lastMessage = [...allMessages].reverse().find(message => message.modelID || message.agent);
  const startedAt = input.session?.time?.created ?? firstTimestamp(input.messages, input.parts);
  const updatedAt = input.session?.time?.updated ?? lastTimestamp(input.messages, input.parts);
  const errorMessage = allMessages.find(message => message.error);
  const tools = input.parts
    .filter((part): part is PartLike & { type: 'tool'; callID: string; tool: string } => part.type === 'tool' && Boolean(part.callID && part.tool))
    .map(toolFromPart)
    .sort((a, b) => (b.startedAt || 0) - (a.startedAt || 0));
  const runningTools = tools.filter(tool => tool.status === 'running' || tool.status === 'pending');
  const completedToolCount = tools.filter(tool => tool.status === 'completed').length;
  const state = resolveState(input.status, input.source, Boolean(errorMessage), runningTools.length > 0);

  return {
    sessionID: input.sessionID,
    source: input.source,
    name: lastMessage?.agent || input.name,
    title: input.title,
    objective: inferObjective(input.title, input.parts, input.messages),
    currentAction: inferCurrentAction(runningTools, input.parts),
    lastActivity: inferLastActivity(input.parts),
    state,
    errorReason: errorMessage?.error ? String(errorMessage.error) : undefined,
    tokens,
    cost: cost || undefined,
    tools,
    runningTools,
    completedToolCount,
    modelID: lastMessage?.modelID || lastModelFromParts(input.parts),
    providerID: lastMessage?.providerID,
    startedAt,
    updatedAt,
    durationMs: startedAt && updatedAt && updatedAt >= startedAt ? updatedAt - startedAt : undefined,
    compacted: Boolean(input.compacted),
  };
}

function toolFromPart(part: PartLike & { callID: string; tool: string }): ToolActivity {
  const status = part.state?.status || 'pending';
  const startedAt = part.state?.time?.start;
  const endedAt = part.state?.time?.end;
  return {
    id: part.id,
    sessionID: part.sessionID,
    messageID: part.messageID,
    callID: part.callID,
    tool: part.tool,
    status,
    title: part.state?.title,
    input: summarizeInput(part.state?.input),
    error: part.state?.error,
    startedAt,
    endedAt,
    durationMs: startedAt && endedAt && endedAt >= startedAt ? endedAt - startedAt : undefined,
  };
}

function resolveState(status: { type: string } | undefined, source: AgentUsage['source'], hasError: boolean, hasRunningTool: boolean): AgentRuntimeState {
  if (hasError || status?.type === 'error') return 'error';
  if (status?.type === 'busy' || hasRunningTool) return 'busy';
  if (source === 'current-session') return 'idle';
  return 'done';
}

function inferObjective(title: string | undefined, parts: PartLike[], messages: MessageLike[]): string | undefined {
  const subtask = parts.find(part => part.type === 'subtask');
  const prompt = subtask?.description || subtask?.prompt;
  const titleCandidate = title && title !== 'subagent' ? title : undefined;
  return truncateMiddle(cleanText(prompt || titleCandidate || firstUserText(parts, messages)), 92);
}

function inferCurrentAction(runningTools: ToolActivity[], parts: PartLike[]): string | undefined {
  const running = runningTools[0];
  if (running) return truncateMiddle(running.title || [running.tool, running.input].filter(Boolean).join(' '), 86);
  const reasoning = [...parts].reverse().find(part => part.type === 'reasoning' && part.text);
  if (reasoning?.text) return truncateMiddle(cleanText(reasoning.text), 86);
  const text = [...parts].reverse().find(part => part.type === 'text' && part.text);
  if (text?.text) return truncateMiddle(cleanText(text.text), 86);
  return undefined;
}

function inferLastActivity(parts: PartLike[]): string | undefined {
  const latestTool = [...parts].reverse().find(part => part.type === 'tool' && part.tool);
  if (latestTool?.tool) {
    const state = latestTool.state?.status || 'pending';
    return truncateMiddle([latestTool.tool, state].join(' '), 64);
  }
  const latestFinish = [...parts].reverse().find(part => part.type === 'step-finish');
  if (latestFinish?.reason) return truncateMiddle(latestFinish.reason, 64);
  return undefined;
}

function firstUserText(parts: PartLike[], messages: MessageLike[]): string | undefined {
  const firstUserMessage = messages.find(message => message.role === 'user');
  if (!firstUserMessage) return undefined;
  return parts.find(part => part.messageID === firstUserMessage.id && part.type === 'text')?.text;
}

function lastModelFromParts(parts: PartLike[]): string | undefined {
  return [...parts].reverse().find(part => part.model?.modelID)?.model?.modelID;
}

function summarizeInput(input: Record<string, unknown> | undefined): string | undefined {
  if (!input) return undefined;
  const preferred = ['file', 'path', 'command', 'cmd', 'pattern', 'query', 'description'];
  for (const key of preferred) {
    const value = input[key];
    if (typeof value === 'string' && value.trim()) return truncateMiddle(value.trim(), 48);
  }
  const keys = Object.keys(input);
  return keys.length ? keys.slice(0, 3).join(', ') : undefined;
}

function messagesForSession(state: ObservatoryState, sessionID: string): MessageLike[] {
  return state.messageOrder
    .map(id => state.messages[id])
    .filter((message): message is MessageLike => Boolean(message && message.sessionID === sessionID));
}

function partsForSession(state: ObservatoryState, sessionID: string): PartLike[] {
  return state.partOrder
    .map(id => state.parts[id])
    .filter((part): part is PartLike => Boolean(part && part.sessionID === sessionID));
}

function firstTimestamp(messages: MessageLike[], parts: PartLike[]): number | undefined {
  for (const timestamp of [
    ...messages.map(message => message.time?.created),
    ...parts.map(part => part.time?.created ?? part.time?.start ?? part.state?.time?.start),
  ]) {
    if (typeof timestamp === 'number') return timestamp;
  }
  return undefined;
}

function lastTimestamp(messages: MessageLike[], parts: PartLike[]): number | undefined {
  const timestamps = [
    ...messages.map(message => message.time?.completed ?? message.time?.created),
    ...parts.map(part => part.time?.end ?? part.time?.created ?? part.time?.start ?? part.state?.time?.end ?? part.state?.time?.start),
  ].filter((value): value is number => typeof value === 'number');
  return timestamps.length ? Math.max(...timestamps) : undefined;
}

function lastSessionTime(session: SessionLike): number {
  return session.time?.updated ?? session.time?.created ?? 0;
}

function addTokens(left: NormalizedTokens, right: NormalizedTokens): NormalizedTokens {
  return {
    total: left.total + right.total,
    input: left.input + right.input,
    output: left.output + right.output,
    reasoning: left.reasoning + right.reasoning,
    cacheRead: left.cacheRead + right.cacheRead,
    cacheWrite: left.cacheWrite + right.cacheWrite,
  };
}

export function normalizeTokens(tokens?: UsageTokens): NormalizedTokens {
  const input = tokens?.input || 0;
  const output = tokens?.output || 0;
  const reasoning = tokens?.reasoning || 0;
  const cacheRead = tokens?.cache?.read || 0;
  const cacheWrite = tokens?.cache?.write || 0;
  return {
    total: tokens?.total || input + output + reasoning + cacheRead + cacheWrite,
    input,
    output,
    reasoning,
    cacheRead,
    cacheWrite,
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

export function formatTokenBreakdown(tokens: NormalizedTokens): string {
  const parts = [
    `in ${formatTokenCount(tokens.input)}`,
    `out ${formatTokenCount(tokens.output)}`,
  ];
  if (tokens.reasoning > 0) parts.push(`reason ${formatTokenCount(tokens.reasoning)}`);
  if (tokens.cacheRead > 0 || tokens.cacheWrite > 0) parts.push(`cache ${formatTokenCount(tokens.cacheRead + tokens.cacheWrite)}`);
  return parts.join(' · ');
}

function cleanText(value: string | undefined): string | undefined {
  return value?.replace(/\s+/g, ' ').trim();
}

function truncateMiddle(value: string | undefined, max: number): string | undefined {
  if (!value) return undefined;
  if (value.length <= max) return value;
  const head = Math.max(8, Math.floor(max * 0.62));
  const tail = Math.max(6, max - head - 3);
  return `${value.slice(0, head)}...${value.slice(-tail)}`;
}
