/** @jsxImportSource @opentui/solid */
import type { TuiPluginApi, TuiPluginModule } from '@opencode-ai/plugin/tui';
import { createEffect, createMemo, createSignal, ErrorBoundary, For, Show } from 'solid-js';

import {
  createObservatoryState,
  formatCost,
  formatDuration,
  formatTokenCount,
  reduceObservatoryState,
  replaceSessionMessages,
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

type TogglePreference =
  | 'observatory.enabled'
  | 'observatory.availableAgentsExpanded'
  | 'observatory.subAgentsExpanded'
  | 'observatory.toolsExpanded'
  | 'observatory.showCost'
  | 'observatory.showEmoji'
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
      availableAgentsExpanded: createBooleanPreference(api, 'observatory.availableAgentsExpanded', false),
      agentsExpanded: createBooleanPreference(api, 'observatory.subAgentsExpanded', false),
      toolsExpanded: createBooleanPreference(api, 'observatory.toolsExpanded', true),
      showCost: createBooleanPreference(api, 'observatory.showCost', true),
      showEmoji: createBooleanPreference(api, 'observatory.showEmoji', true),
      showTools: createBooleanPreference(api, 'observatory.showTools', true),
    };

    const scheduleRefresh = () => {
      if (refreshTimer) {
        clearTimeout(refreshTimer);
      }
      refreshTimer = setTimeout(() => {
        refreshTimer = undefined;
        setRevision((value) => value + 1);
      }, REFRESH_DEBOUNCE_MS);
    };

    const commit = (next: ObservatoryState) => {
      observatoryState = next;
      scheduleRefresh();
    };

    const ingest = (event: ObservatoryEvent) => {
      commit(reduceObservatoryState(observatoryState, event));
    };

    const hydrateSession = async (sessionID: string) => {
      if (hydratingSessions.has(sessionID)) {
        return;
      }
      hydratingSessions.add(sessionID);
      try {
        setUnavailable(undefined);
        hydrateSessionFromTuiState(sessionID);
        const children = await api.client.session.children({ sessionID }) as { data?: SessionLike[] };
        for (const child of (children.data ?? []).slice(0, MAX_CHILDREN_TO_HYDRATE)) {
          ingest({ type: 'session.created', session: child });
          await hydrateChildSession(child.id);
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
              showEmoji={preferences.showEmoji.value}
              showTools={preferences.showTools.value}
              toolsExpanded={preferences.toolsExpanded.value}
              toggleAgents={preferences.agentsExpanded.toggle}
              toggleAvailableAgents={preferences.availableAgentsExpanded.toggle}
              unavailable={unavailable()}
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
  showEmoji: () => boolean;
  showTools: () => boolean;
  toggleAvailableAgents: () => void;
  toggleAgents: () => void;
  unavailable: () => string | undefined;
  hydrateSession: (sessionID: string) => Promise<void>;
  getSnapshot: (sessionID: string) => ObservatorySnapshot;
  sessionID: string;
}) {
  createEffect(() => { void props.hydrateSession(props.sessionID); });
  const snapshot = createMemo(() => props.getSnapshot(props.sessionID));

  return (
    <Show when={props.enabled()}>
      <box flexDirection="column" paddingTop={0} gap={0}>
        <Show when={!props.unavailable()} fallback={<text fg={COLORS.error}>Observatory unavailable</text>}>
          <RootModelLine showEmoji={props.showEmoji} snapshot={snapshot()} />
          <AvailableAgentsSection agents={snapshot().availableAgents} expanded={props.availableAgentsExpanded} onToggle={props.toggleAvailableAgents} showEmoji={props.showEmoji} snapshot={snapshot()} showCost={props.showCost} />
          <SubAgentsSection agents={snapshot().agents} expanded={props.agentsExpanded} onToggle={props.toggleAgents} showCost={props.showCost} showEmoji={props.showEmoji} showTools={() => props.showTools() && props.toolsExpanded()} />
        </Show>
      </box>
    </Show>
  );
}

function RootModelLine(props: { showEmoji: () => boolean; snapshot: ObservatorySnapshot }) {
  return (
    <box flexDirection="row" paddingTop={1}>
      <text fg={COLORS.bullet}>{props.showEmoji() ? '🤖 ' : '• '}</text>
      <text fg={COLORS.text}>{snapshot().root.name} · {snapshot().totalTokens.total} tok</text>
    </box>
  );
}

function AvailableAgentsSection(props: {
  agents: AvailableAgent[];
  expanded: () => boolean;
  onToggle: () => void;
  showEmoji: () => boolean;
  snapshot: ObservatorySnapshot;
  showCost: () => boolean;
}) {
  return (
    <box flexDirection="column" paddingTop={0} gap={0}>
      <box flexDirection="row" gap={1} onMouseDown={props.onToggle}>
        <text fg={COLORS.title} bold>{props.expanded() ? '▼' : '▶'}</text>
        <text fg={COLORS.title}><b>Agents</b></text>
      </box>
      <Show when={props.expanded()}>
        <For each={props.agents}>{(agent) => (
          <box flexDirection="row" gap={1}>
            <text fg={agent.active ? COLORS.success : COLORS.muted}>{props.showEmoji() ? (agent.active ? '🟢' : '⚪') : (agent.active ? '•' : '·')}</text>
            <text fg={COLORS.text}>{agent.name}</text>
          </box>
        )}</For>
      </Show>
    </box>
  );
}

function SubAgentsSection(props: {
  agents: AgentUsage[];
  expanded: () => boolean;
  onToggle: () => void;
  showCost: () => boolean;
  showEmoji: () => boolean;
  showTools: () => boolean;
}) {
  return (
    <box flexDirection="column" paddingTop={1} gap={0}>
      <box flexDirection="row" gap={1} onMouseDown={props.onToggle}>
        <text fg={COLORS.accent}>{props.showEmoji() ? '🟡' : '•'}</text>
        <text fg={COLORS.accent}>{props.agents.filter(a => a.state === 'busy').length} running</text>
        <text fg={COLORS.faint}>·</text>
        <text fg={COLORS.success}>{props.showEmoji() ? '✅' : '✓'} {props.agents.filter(a => a.state === 'done').length} done</text>
      </box>
    </box>
  );
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
    api.event.on('message.updated', (e) => ingest({ type: 'message.updated', message: e.properties.info as MessageLike })),
  ];
  for (const d of disposers) api.lifecycle.onDispose(d);
}

function registerCommands(api: TuiPluginApi, prefs: Record<string, BooleanSignal>) {
  const cmds = [
    { title: 'Toggle Observatory', value: `${PLUGIN_ID}.toggle`, slash: { name: 'observatory' }, onSelect: prefs.enabled.toggle },
    { title: 'Toggle Agents', value: `${PLUGIN_ID}.toggle-agents`, slash: { name: 'observatory-agents' }, onSelect: prefs.availableAgentsExpanded.toggle },
    { title: 'Toggle Subagents', value: `${PLUGIN_ID}.toggle-subagents`, slash: { name: 'observatory-subagents' }, onSelect: prefs.agentsExpanded.toggle },
    { title: 'Show/Hide Cost', value: `${PLUGIN_ID}.toggle-cost`, slash: { name: 'observatory-cost' }, onSelect: prefs.showCost.toggle },
    { title: 'Show/Hide Emoji', value: `${PLUGIN_ID}.toggle-emoji`, slash: { name: 'observatory-emoji' }, onSelect: prefs.showEmoji.toggle },
  ];
  const dispose = api.command.register(() => cmds);
  api.lifecycle.onDispose(dispose);
}

function hydrateSessionFromTuiState(api: TuiPluginApi, sessionID: string) {
  const messages = Array.from(api.state.session.messages(sessionID) as readonly MessageLike[]);
  for (const msg of messages) {
    const parts = Array.from(api.state.part(msg.id) as readonly PartLike[]);
    // simplified - full version would update state
  }
}

async function hydrateChildSession(api: TuiPluginApi, sessionID: string) {
  const response = await api.client.session.messages({ sessionID, limit: 100 }) as { data?: Array<{ info: MessageLike; parts: PartLike[] }> };
  // simplified
}