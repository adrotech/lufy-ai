/** @jsxImportSource @opentui/solid */
import type { TuiPluginApi, TuiPluginModule } from '@opencode-ai/plugin/tui';
import { createEffect, createMemo, createSignal, ErrorBoundary, For, Show } from 'solid-js';

import {
  createObservatoryState,
  formatCost,
  formatDuration,
  formatTokenBreakdown,
  formatTokenCount,
  reduceObservatoryState,
  selectObservatorySnapshot,
  type AgentCatalogEntry,
  type AgentUsage,
  type AvailableAgent,
  type MessageLike,
  type ObservatoryEvent,
  type ObservatorySnapshot,
  type ObservatoryState,
  type PartLike,
  type SessionLike,
  type ToolActivity,
} from '../agent-observatory/state';

const PLUGIN_ID = 'lufy-ai.observatory';
const MAX_CHILDREN_TO_HYDRATE = 12;
const REFRESH_DEBOUNCE_MS = 250;
const MAX_TOOLS = 50;
const VISIBLE_TOOLS = 5;
const VISIBLE_PRIMARY_AGENTS = 4;
const VISIBLE_SUBAGENTS = 8;
const VISIBLE_AGENT_TOOLS = 3;
const DETAIL_MAX = 36;
const TOOL_DETAIL_MAX = 22;
const AGENT_NAME_MAX = 18;
const TOOL_NAME_MAX = 10;

type TogglePreference =
  | 'observatory.enabled'
  | 'observatory.availableAgentsExpanded'
  | 'observatory.subAgentsExpanded'
  | 'observatory.toolsExpanded'
  | 'observatory.showCost'
  | 'observatory.showTools';
type TuiColor = string;

const COLORS = {
  title: '#d6dcff',
  bullet: '#b6f09c',
  accent: '#ffd580',
  success: '#b6f09c',
  text: '#f4f7ff',
  muted: '#9aa3c7',
  faint: '#6f789c',
  error: '#ff6b8a',
  busy: '#8bd5ff',
};

type BooleanSignal = {
  value: () => boolean;
  set: (value: boolean) => void;
  toggle: () => void;
};

const moduleDefinition: TuiPluginModule = {
  id: PLUGIN_ID,
  async tui(api) {
    let observatoryState = createObservatoryState();
    let refreshTimer: ReturnType<typeof setTimeout> | undefined;
    let agentCatalogHydrated = false;
    const [revision, setRevision] = createSignal(0);
    const [agentCatalog, setAgentCatalog] = createSignal<AgentCatalogEntry[]>([]);
    const [unavailable, setUnavailable] = createSignal<string | undefined>(undefined);
    const hydratingSessions = new Set<string>();
    const preferences = {
      enabled: createBooleanPreference(api, 'observatory.enabled', true),
      availableAgentsExpanded: createBooleanPreference(api, 'observatory.availableAgentsExpanded', true),
      agentsExpanded: createBooleanPreference(api, 'observatory.subAgentsExpanded', true),
      toolsExpanded: createBooleanPreference(api, 'observatory.toolsExpanded', true),
      showCost: createBooleanPreference(api, 'observatory.showCost', true),
      showTools: createBooleanPreference(api, 'observatory.showTools', true),
    };

    const scheduleRefresh = () => {
      if (refreshTimer) clearTimeout(refreshTimer);
      refreshTimer = setTimeout(() => {
        refreshTimer = undefined;
        setRevision((value) => value + 1);
      }, REFRESH_DEBOUNCE_MS);
    };

    const hydrateAgentCatalog = async () => {
      if (agentCatalogHydrated) return;
      agentCatalogHydrated = true;
      try {
        const response = await api.client.app.agents({ directory: api.state.path.directory }) as { data?: AgentCatalogEntry[] };
        setAgentCatalog((response.data ?? []).filter((agent) => !agent.hidden));
        scheduleRefresh();
      } catch {
        agentCatalogHydrated = false;
      }
    };

    const commit = (next: ObservatoryState) => {
      observatoryState = next;
      scheduleRefresh();
    };

    const ingest = (event: ObservatoryEvent) => {
      commit(reduceObservatoryState(observatoryState, event));
    };

    const hydrateSession = async (sessionID: string) => {
      if (!sessionID || hydratingSessions.has(sessionID)) return;
      hydratingSessions.add(sessionID);
      try {
        setUnavailable(undefined);
        hydrateSessionFromTuiState(api, sessionID, ingest);
        const children = await api.client.session.children({ sessionID }) as { data?: SessionLike[] };
        for (const child of (children.data ?? []).slice(0, MAX_CHILDREN_TO_HYDRATE)) {
          ingest({ type: 'session.created', session: child });
          await hydrateChildSession(api, child.id, ingest);
        }
      } catch {
        setUnavailable('Observatory unavailable');
      } finally {
        hydratingSessions.delete(sessionID);
      }
    };

    const getSnapshot = (sessionID: string) => {
      revision();
      return selectObservatorySnapshot(observatoryState, sessionID, {
        availableAgents: agentCatalog(),
        defaultAgent: api.state.config.default_agent,
        maxAgentTools: VISIBLE_TOOLS,
        maxTools: MAX_TOOLS,
      });
    };

    registerEvents(api, ingest);
    registerCommands(api, preferences);
    void hydrateAgentCatalog();

    api.lifecycle.onDispose(() => {
      if (refreshTimer) clearTimeout(refreshTimer);
    });

    api.slots.register({
      order: 650,
      slots: {
        sidebar_content: (_ctx, props) => (
          <ErrorBoundary fallback={() => <box flexDirection="column" paddingTop={1}><text fg={COLORS.error}>Observatory unavailable</text></box>}>
            <AgentObservatoryPanel
              agentsExpanded={preferences.agentsExpanded.value}
              availableAgentsExpanded={preferences.availableAgentsExpanded.value}
              enabled={preferences.enabled.value}
              getSnapshot={getSnapshot}
              hydrateSession={hydrateSession}
              sessionID={props.session_id}
              showCost={preferences.showCost.value}
              showTools={preferences.showTools.value}
              toolsExpanded={preferences.toolsExpanded.value}
              toggleAgents={preferences.agentsExpanded.toggle}
              toggleAvailableAgents={preferences.availableAgentsExpanded.toggle}
              unavailable={unavailable}
            />
          </ErrorBoundary>
        ),
      },
    });
  },
};

export default moduleDefinition;

function AgentObservatoryPanel(props: {
  enabled: () => boolean;
  availableAgentsExpanded: () => boolean;
  agentsExpanded: () => boolean;
  toolsExpanded: () => boolean;
  showCost: () => boolean;
  showTools: () => boolean;
  toggleAvailableAgents: () => void;
  toggleAgents: () => void;
  unavailable: () => string | undefined;
  hydrateSession: (sessionID: string) => Promise<void>;
  getSnapshot: (sessionID: string) => ObservatorySnapshot;
  sessionID: string;
}) {
  let lastHydratedSessionID: string | undefined;
  createEffect(() => {
    const sessionID = props.sessionID;
    if (!props.enabled() || !sessionID || sessionID === lastHydratedSessionID) return;
    lastHydratedSessionID = sessionID;
    void props.hydrateSession(sessionID);
  });
  const snapshot = createMemo(() => props.getSnapshot(props.sessionID));

  return (
    <Show when={props.enabled()}>
      <box flexDirection="column" paddingTop={0} gap={0}>
        <Show when={!props.unavailable()} fallback={<text fg={COLORS.error}>Observatory unavailable</text>}>
          <AvailableAgentsSection agents={snapshot().availableAgents} expanded={props.availableAgentsExpanded} onToggle={props.toggleAvailableAgents} />
          <SubAgentsSection agents={snapshot().agents} expanded={props.agentsExpanded} onToggle={props.toggleAgents} showCost={props.showCost} showTools={() => props.showTools() && props.toolsExpanded()} totalCost={snapshot().totalCost} totalTokens={snapshot().totalTokens} />
        </Show>
      </box>
    </Show>
  );
}

function AvailableAgentsSection(props: {
  agents: AvailableAgent[];
  expanded: () => boolean;
  onToggle: () => void;
}) {
  const primaryAgents = createMemo(() => props.agents.filter(agent => agent.mode !== 'subagent'));
  const subagentCatalog = createMemo(() => props.agents.filter(agent => agent.mode === 'subagent'));
  const visiblePrimary = createMemo(() => primaryAgents().slice(0, VISIBLE_PRIMARY_AGENTS));
  const hiddenPrimary = createMemo(() => Math.max(0, primaryAgents().length - visiblePrimary().length));

  return (
    <box flexDirection="column" paddingTop={0} gap={0}>
      <box flexDirection="row" gap={1} onMouseDown={props.onToggle}>
        <DisclosureArrow expanded={props.expanded} color={props.expanded() ? COLORS.title : COLORS.faint} />
        <text fg={COLORS.title}><b>Agents</b></text>
        <text fg={COLORS.faint}>· {primaryAgents().length} primary</text>
      </box>
      <Show when={props.expanded()}>
        <For each={visiblePrimary()}>{(agent) => (
          <box flexDirection="row" gap={1} paddingLeft={2}>
            <Dot color={agent.active ? COLORS.success : (agent.color || COLORS.faint)} dim={!agent.active} />
            <text fg={agent.active ? COLORS.text : COLORS.muted}>{agent.name}</text>
            <Show when={agent.active}><text fg={COLORS.success}>active</text></Show>
          </box>
        )}</For>
        <Show when={hiddenPrimary() > 0}>
          <box flexDirection="row" gap={1} paddingLeft={2}>
            <text fg={COLORS.faint}>  •</text>
            <text fg={COLORS.faint}>+{hiddenPrimary()} more primary</text>
          </box>
        </Show>
        <Show when={subagentCatalog().length > 0}>
          <box flexDirection="row" gap={1} paddingLeft={2}>
            <text fg={COLORS.faint}>  •</text>
            <text fg={COLORS.faint}>{subagentCatalog().length} subagents available via @</text>
          </box>
        </Show>
      </Show>
    </box>
  );
}

function SubAgentsSection(props: {
  agents: AgentUsage[];
  expanded: () => boolean;
  onToggle: () => void;
  showCost: () => boolean;
  showTools: () => boolean;
  totalCost?: number;
  totalTokens: AgentUsage['tokens'];
}) {
  const [focusedAgentID, setFocusedAgentID] = createSignal<string | undefined>(undefined);
  const [collapsedAgentIDs, setCollapsedAgentIDs] = createSignal<ReadonlySet<string>>(new Set());
  const busy = createMemo(() => props.agents.filter(agent => agent.state === 'busy').length);
  const done = createMemo(() => props.agents.filter(agent => agent.state === 'done').length);
  const errored = createMemo(() => props.agents.filter(agent => agent.state === 'error').length);
  const headerSummary = createMemo(() => subagentSummary({
    busy: busy(),
    done: done(),
    errored: errored(),
    total: props.agents.length,
    tokens: props.totalTokens.total,
    cost: props.showCost() ? props.totalCost : undefined,
  }));
  const visibleAgents = createMemo(() => [...props.agents].sort(sortAgentsByRelevance).slice(0, VISIBLE_SUBAGENTS));
  const hiddenAgents = createMemo(() => Math.max(0, props.agents.length - visibleAgents().length));
  const shouldAutoExpandAgent = (agent: AgentUsage) => agent.state === 'busy' || agent.state === 'error';
  const isAgentExpanded = (agent: AgentUsage) => {
    const focused = focusedAgentID();
    if (focused) return focused === agent.sessionID;
    if (collapsedAgentIDs().has(agent.sessionID)) return false;
    return shouldAutoExpandAgent(agent);
  };
  const setAgentCollapsed = (sessionID: string, collapsed: boolean) => {
    setCollapsedAgentIDs((current) => {
      const next = new Set(current);
      if (collapsed) next.add(sessionID);
      else next.delete(sessionID);
      return next;
    });
  };
  const toggleAgent = (agent: AgentUsage) => {
    const focused = focusedAgentID();
    const autoExpanded = shouldAutoExpandAgent(agent);
    const manuallyCollapsed = collapsedAgentIDs().has(agent.sessionID);

    if (focused === agent.sessionID) {
      setFocusedAgentID(undefined);
      setAgentCollapsed(agent.sessionID, autoExpanded);
      return;
    }

    if (focused) {
      setAgentCollapsed(agent.sessionID, false);
      setFocusedAgentID(agent.sessionID);
      return;
    }

    if (autoExpanded && !manuallyCollapsed) {
      setAgentCollapsed(agent.sessionID, true);
      return;
    }

    setAgentCollapsed(agent.sessionID, false);
    setFocusedAgentID(autoExpanded ? undefined : agent.sessionID);
  };

  createEffect(() => {
    const currentIDs = new Set(props.agents.map(agent => agent.sessionID));
    const focused = focusedAgentID();
    if (focused && !currentIDs.has(focused)) setFocusedAgentID(undefined);
    setCollapsedAgentIDs((current) => {
      const next = new Set([...current].filter(sessionID => currentIDs.has(sessionID)));
      return next.size === current.size ? current : next;
    });
  });

  return (
    <box flexDirection="column" paddingTop={1} gap={0}>
      <box flexDirection="row" gap={1} onMouseDown={props.onToggle}>
        <DisclosureArrow expanded={props.expanded} color={props.expanded() ? COLORS.title : COLORS.faint} />
        <text fg={COLORS.title}><b>Subagents</b></text>
        <text fg={COLORS.faint}>· {truncateText(headerSummary(), 32)}</text>
      </box>
      <Show when={props.expanded()}>
        <Show when={props.agents.length > 0} fallback={<box paddingLeft={2}><text fg={COLORS.muted}>No subagent sessions yet</text></box>}>
          <For each={visibleAgents()}>{(agent) => <AgentLine agent={agent} expanded={() => isAgentExpanded(agent)} onToggle={() => toggleAgent(agent)} showCost={props.showCost} showTools={props.showTools} />}</For>
          <Show when={hiddenAgents() > 0}>
            <box flexDirection="row" gap={1} paddingLeft={2}>
              <text fg={COLORS.faint}>+{hiddenAgents()} older subagents hidden</text>
            </box>
          </Show>
        </Show>
      </Show>
    </box>
  );
}

function AgentLine(props: { agent: AgentUsage; expanded: () => boolean; onToggle: () => void; showCost: () => boolean; showTools: () => boolean }) {
  const activity = createMemo(() => agentActivity(props.agent));
  const toolSummary = createMemo(() => summarizeTools(props.agent));
  const collapsedMeta = createMemo(() => compactMeta([
    `${formatTokenCount(props.agent.tokens.total)} tok`,
    props.agent.durationMs ? formatDuration(props.agent.durationMs) : undefined,
  ]));
  const expandedMetrics = createMemo(() => compactMeta([
    formatTokenBreakdown(props.agent.tokens),
    props.showCost() && props.agent.cost ? formatCost(props.agent.cost) : undefined,
    props.agent.modelID,
    props.agent.durationMs ? formatDuration(props.agent.durationMs) : undefined,
    props.agent.compacted ? 'compacted' : undefined,
  ]));
  const visibleTools = createMemo(() => props.agent.tools.slice(0, VISIBLE_AGENT_TOOLS));

  return (
    <box flexDirection="column" gap={0}>
      <box flexDirection="row" gap={1} paddingLeft={2} onMouseDown={props.onToggle}>
        <text fg={stateColor(props.agent.state)}>{stateIcon(props.agent.state)}</text>
        <text fg={props.expanded() ? COLORS.text : COLORS.muted}>{shortAgentName(props.agent.name)}</text>
        <Show when={collapsedMeta()}><text fg={COLORS.faint}>· {collapsedMeta()}</text></Show>
      </box>
      <Show when={props.expanded()}>
        <Show when={props.agent.objective}>
          <box flexDirection="row" gap={1} paddingLeft={2}>
            <text fg={COLORS.muted}>goal:</text>
            <text fg={COLORS.text}>{truncateText(props.agent.objective!, DETAIL_MAX)}</text>
          </box>
        </Show>
        <Show when={activity()}>
          <box flexDirection="row" gap={1} paddingLeft={2}>
            <text fg={props.agent.state === 'busy' ? COLORS.busy : COLORS.muted}>{activityLabel(props.agent)}:</text>
            <text fg={COLORS.text}>{truncateText(activity()!, DETAIL_MAX)}</text>
          </box>
        </Show>
        <Show when={toolSummary()}>
          <box flexDirection="row" gap={1} paddingLeft={2}>
            <text fg={COLORS.muted}>tools:</text>
            <text fg={COLORS.faint}>{truncateText(toolSummary()!, DETAIL_MAX)}</text>
          </box>
        </Show>
        <box flexDirection="row" gap={1} paddingLeft={2}>
          <text fg={COLORS.muted}>meta:</text>
          <text fg={COLORS.faint}>{truncateText(expandedMetrics(), DETAIL_MAX)}</text>
        </box>
        <Show when={props.agent.errorReason}>
          <box flexDirection="row" gap={1} paddingLeft={2}>
            <text fg={COLORS.error}>error:</text>
            <text fg={COLORS.error}>{truncateText(props.agent.errorReason!, DETAIL_MAX)}</text>
          </box>
        </Show>
        <Show when={props.showTools() && visibleTools().length > 0}>
          <For each={visibleTools()}>{(tool) => <ToolLine tool={tool} />}</For>
        </Show>
      </Show>
    </box>
  );
}

function ToolLine(props: { tool: ToolActivity }) {
  const toolColor = () => props.tool.status === 'error' ? COLORS.error : COLORS.faint;
  const detail = () => formatToolDetail(props.tool.title || props.tool.input || '');

  return (
    <box flexDirection="row" gap={1} paddingLeft={4}>
      <text fg={toolColor()}>{toolStatusIcon(props.tool.status)}</text>
      <text fg={props.tool.status === 'error' ? COLORS.error : COLORS.muted}>{truncateText(props.tool.tool, TOOL_NAME_MAX)}</text>
      <Show when={detail()}><text fg={COLORS.faint}>{detail()}</text></Show>
      <Show when={props.tool.durationMs}><text fg={COLORS.faint}>{formatDuration(props.tool.durationMs)}</text></Show>
    </box>
  );
}

function Dot(props: { color: TuiColor; dim?: boolean }) {
  return <text fg={props.dim ? COLORS.faint : props.color}>•</text>;
}

function DisclosureArrow(props: { expanded: () => boolean; color: TuiColor }) {
  return <text fg={props.color}>{props.expanded() ? '▼' : '▶'}</text>;
}

function sortAgentsByRelevance(a: AgentUsage, b: AgentUsage): number {
  const stateRank = (agent: AgentUsage) => agent.state === 'busy' ? 0 : agent.state === 'error' ? 1 : 2;
  return stateRank(a) - stateRank(b) || (b.updatedAt || 0) - (a.updatedAt || 0);
}

function compactMeta(values: Array<string | false | undefined>): string {
  return values.filter(Boolean).join(' · ');
}

function subagentSummary(input: { busy: number; done: number; errored: number; total: number; tokens: number; cost?: number }): string {
  if (input.total === 0) return 'none yet';
  const idle = Math.max(0, input.total - input.busy - input.done - input.errored);
  const allDone = input.done === input.total && input.busy === 0 && input.errored === 0 && idle === 0;
  return compactMeta([
    allDone ? undefined : `${input.total} total`,
    input.busy > 0 ? `${input.busy} run` : undefined,
    input.done > 0 ? `${input.done} done` : undefined,
    input.errored > 0 ? `${input.errored} err` : undefined,
    idle > 0 ? `${idle} idle` : undefined,
    `${formatTokenCount(input.tokens)} tok`,
  ]);
}

function agentActivity(agent: AgentUsage): string | undefined {
  return agent.currentAction || agent.lastActivity || agent.title || undefined;
}

function activityLabel(agent: AgentUsage): string {
  return agent.currentAction ? 'now' : 'last';
}

function stateColor(state: AgentUsage['state']): TuiColor {
  if (state === 'error') return COLORS.error;
  if (state === 'busy') return COLORS.busy;
  if (state === 'done') return COLORS.success;
  return COLORS.muted;
}

function stateIcon(state: AgentUsage['state']): string {
  if (state === 'busy') return '●';
  if (state === 'done') return '✓';
  if (state === 'error') return '✕';
  return '○';
}

function toolStatusIcon(status: ToolActivity['status']): string {
  if (status === 'running') return '…';
  if (status === 'error') return '!';
  return '✓';
}

function shortAgentName(name: string): string {
  return truncateText(name, AGENT_NAME_MAX);
}

function summarizeTools(agent: AgentUsage): string | undefined {
  const running = agent.runningTools.map(formatCompactTool);
  const recent = agent.tools.slice(0, VISIBLE_AGENT_TOOLS).map(formatCompactTool);
  const summary = (running.length ? running : recent).filter(Boolean);
  const count = agent.runningTools.length || agent.completedToolCount;
  if (summary.length === 0) return count > 0 ? `${count} tools` : undefined;
  const suffix = count > summary.length ? ` +${count - summary.length}` : undefined;
  return compactMeta([summary.join(', '), suffix]);
}

function formatCompactTool(tool: ToolActivity): string {
  const label = formatToolDetail(tool.title || tool.input || tool.tool);
  const name = truncateText(tool.tool, TOOL_NAME_MAX);
  return `${toolStatusIcon(tool.status)} ${name}${label && label !== tool.tool ? `:${label}` : ''}`;
}

function formatToolDetail(value: string): string {
  const compact = compactText(value);
  const detail = looksLikePath(compact) ? pathBasename(compact) : compact;
  return truncateText(detail, TOOL_DETAIL_MAX);
}

function truncateText(value: string, maxLength: number): string {
  const compact = compactText(value);
  if (compact.length <= maxLength) return compact;
  return `${compact.slice(0, Math.max(0, maxLength - 1))}…`;
}

function compactText(value: string): string {
  return value.replace(/\s+/g, ' ').trim();
}

function looksLikePath(value: string): boolean {
  return value.startsWith('/Users/') || value.startsWith('.opencode/') || value.includes('/');
}

function pathBasename(value: string): string {
  const trimmed = value.replace(/[/?#]+$/, '');
  return trimmed.split('/').filter(Boolean).pop() || trimmed;
}

function createBooleanPreference(api: TuiPluginApi, key: TogglePreference, fallback: boolean): BooleanSignal {
  const [value, setValue] = createSignal(api.kv.get<boolean>(key, fallback));
  const set = (next: boolean) => { setValue(next); api.kv.set(key, next); };
  return { value, set, toggle: () => set(!value()) };
}

function registerEvents(api: TuiPluginApi, ingest: (event: ObservatoryEvent) => void) {
  const disposers = [
    api.event.on('session.created', (e) => ingest({ type: 'session.created', session: e.properties.info as SessionLike })),
    api.event.on('session.updated', (e) => ingest({ type: 'session.updated', session: e.properties.info as SessionLike })),
    api.event.on('session.status', (e) => ingest({ type: 'session.status', sessionID: e.properties.sessionID, status: e.properties.status })),
    api.event.on('session.compacted', (e) => ingest({ type: 'session.compacted', sessionID: e.properties.sessionID })),
    api.event.on('message.updated', (e) => ingest({ type: 'message.updated', message: e.properties.info as MessageLike })),
    api.event.on('message.part.updated', (e) => ingest({ type: 'message.part.updated', part: e.properties.part as PartLike })),
    api.event.on('message.part.removed', (e) => ingest({ type: 'message.part.removed', partID: e.properties.partID })),
  ];
  for (const d of disposers) api.lifecycle.onDispose(d);
}

function registerCommands(api: TuiPluginApi, prefs: Record<string, BooleanSignal>) {
  const cmds = [
    { title: 'Toggle Observatory', value: `${PLUGIN_ID}.toggle`, slash: { name: 'observatory' }, onSelect: prefs.enabled.toggle },
    { title: 'Toggle Agents', value: `${PLUGIN_ID}.toggle-agents`, slash: { name: 'observatory-agents' }, onSelect: prefs.availableAgentsExpanded.toggle },
    { title: 'Toggle Subagents', value: `${PLUGIN_ID}.toggle-subagents`, slash: { name: 'observatory-subagents' }, onSelect: prefs.agentsExpanded.toggle },
    { title: 'Show/Hide Cost', value: `${PLUGIN_ID}.toggle-cost`, slash: { name: 'observatory-cost' }, onSelect: prefs.showCost.toggle },
  ];
  const dispose = api.command.register(() => cmds);
  api.lifecycle.onDispose(dispose);
}

function hydrateSessionFromTuiState(api: TuiPluginApi, sessionID: string, ingest: (event: ObservatoryEvent) => void) {
  const messages = safeArray<MessageLike>(() => api.state.session.messages(sessionID) as readonly MessageLike[]);
  for (const msg of messages) {
    ingest({ type: 'message.updated', message: msg });
    for (const part of safeArray<PartLike>(() => api.state.part(msg.id) as readonly PartLike[])) {
      ingest({ type: 'message.part.updated', part });
    }
  }
}

async function hydrateChildSession(api: TuiPluginApi, sessionID: string, ingest: (event: ObservatoryEvent) => void) {
  const response = await api.client.session.messages({ sessionID, limit: 100 }) as { data?: Array<{ info: MessageLike; parts?: PartLike[] }> };
  for (const entry of response.data ?? []) {
    if (entry.info) ingest({ type: 'message.updated', message: entry.info });
    for (const part of entry.parts ?? []) {
      ingest({ type: 'message.part.updated', part });
    }
  }
}

function safeArray<T>(read: () => readonly T[] | undefined): T[] {
  try {
    return Array.from(read() ?? []);
  } catch {
    return [];
  }
}
