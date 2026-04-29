# lufy-ai

AI-first workflow kit for OpenCode projects with OpenSpec, specialized agents, delivery policy, and a local observability panel.

## What This Repository Actually Ships

This repository is not an app framework by itself. It is a repository-local operating layer that gets copied into another project and gives that project:

- a primary `orchestrator` agent plus focused subagents
- an OpenSpec / spec-driven workflow
- repository delivery rules
- slash commands for explore, propose, apply, verify, and archive
- an Agent Observatory TUI plugin
- an `AGENTS.md` bootstrap template for project conventions

Today, the repo contains these building blocks:

- `.opencode/agents/`: `orchestrator`, `explorer`, `implementer`, `validator`, `reviewer`, `delivery`
- `.opencode/commands/`: `opsx-explore`, `opsx-propose`, `opsx-apply`, `opsx-verify`, `opsx-archive`
- `.opencode/skills/sdd-workflow/`: OpenSpec lifecycle skills
- `.opencode/policies/delivery.md`: delivery and traceability policy
- `.opencode/plugins/agent-observatory.tsx`: local TUI observability plugin
- `AGENTS.md.template`: project-specific conventions bootstrap
- `scripts/install.sh`: installer for target repositories
- `openspec/`: starter OpenSpec structure and config

## End-to-End Flow

### 1. Install into a target repository

The installer:

1. checks dependencies
2. warns if `.opencode/` or `AGENTS.md` already exist
3. tries to detect the project stack
4. copies the local OpenCode assets into the target project
5. creates `AGENTS.md` from `AGENTS.md.template` if needed
6. copies `tui.json`
7. optionally enables Engram-oriented memory usage if the user already has it installed

Install:

```bash
git clone https://github.com/adrianrojas/lufy-ai.git /tmp/lufy-ai
cd /tmp/lufy-ai
./scripts/install.sh
```

Or:

```bash
curl -fsSL https://raw.githubusercontent.com/adrianrojas/lufy-ai/main/scripts/install.sh | bash
```

### 2. Let the orchestrator route the work

Once installed, OpenCode uses the repository-local agents under `.opencode/agents/`.

Current topology:

| Agent | Responsibility |
| --- | --- |
| `orchestrator` | Primary router. Chooses the right specialist and enforces minimal-overhead coordination. |
| `explorer` | Read-only impact analysis, architecture reading, file discovery, and implementation handoff. |
| `implementer` | Bounded code, tests, docs, and configuration changes. |
| `validator` | Read-only compile/test evidence and failure diagnosis. |
| `reviewer` | Read-only code quality, architecture, risk, and missing-test review. |
| `delivery` | Git / GitHub operations, branch hygiene, push, PR creation, and traceability gates. |

### 3. Work through the OpenSpec lifecycle

The OpenSpec workflow in this repo is centered on five commands:

- `/opsx-explore`: read-only exploration and requirement clarification
- `/opsx-propose`: create change artifacts such as `proposal.md`, `design.md`, and `tasks.md`
- `/opsx-apply`: implement tasks from an active change
- `/opsx-verify`: verify completeness, correctness, and coherence against the artifacts
- `/opsx-archive`: archive a completed change

At a repo level, the lifecycle is:

1. explore the problem or existing code
2. propose a named change in `openspec/changes/<name>/`
3. apply tasks with focused implementation
4. verify with explicit evidence
5. archive only when the change is complete

### 4. Enforce delivery separately from implementation

`implementer` is intentionally not the delivery owner.

Delivery rules live in `.opencode/policies/delivery.md` and establish:

- protected source branches
- default PR base branch
- validation tiers for iteration vs final delivery
- OpenSpec task closure gates
- GitHub Project sync expectations
- Spanish as the default language for human-facing delivery artifacts

This keeps code editing, validation, and Git/GitHub operations clearly separated.

### 5. Observe local agent activity

The repository ships a local Agent Observatory plugin for the OpenCode TUI. It is loaded through `tui.json` and is explicitly local-only.

That gives visibility into:

- current active agents
- subagent activity
- tool usage summaries
- optional cost display

No external telemetry is part of the default design.

## Current Documentation Drift

This README now reflects the repository as it exists today.

The previous version had drift in a few places:

- it referenced stack template files that are not present in this repository
- it grouped `React`, `Next.js`, and `Vue` under a single `frontend-react` bucket
- it documented a `backend-node` template as if it were a first-class direction
- it linked to documentation files that do not exist in `docs/`

The sections below describe the recommended direction for stack templates and subagents, without pretending those templates already exist as dedicated files in this repo.

## Recommended Stack Template Direction

After reviewing the current repo and current official documentation, the stack template lineup should be narrower and more opinionated.

### Frontend templates that should exist

These should replace the overloaded "frontend-react" idea:

| Template | Why it should be separate | Suggested specialist subagents |
| --- | --- | --- |
| `frontend-react` | Pure React projects still need their own conventions around component boundaries, hooks, state, accessibility, and render performance. | `react-ui`, `react-state-performance`, `react-testing-a11y` |
| `frontend-nextjs` | Next.js App Router adds server/client component boundaries, route handlers, caching, streaming, and deployment/runtime decisions that deserve their own agent behavior. | `nextjs-app-router`, `nextjs-server-runtime`, `nextjs-data-cache` |
| `frontend-astro` | Astro has a very different model: islands architecture, content collections, integrations, adapters, and static/hybrid/server rendering modes. | `astro-islands-content`, `astro-integrations`, `astro-ssr-adapter` |

### Templates that should stay

These still make sense as separate tracks:

- `mobile-expo`
- `backend-spring`

### Template that should be removed

This should stop being documented as a target template:

- `backend-node`

If a repository is Node-based, that should usually be expressed through a more concrete stack profile such as `frontend-nextjs`, `frontend-react`, or a future explicit backend profile with stronger boundaries than "Node.js".

## Why These Frontend Splits Matter

### React

The React docs currently recommend starting new React apps with a framework, and `Create React App` is deprecated. Even so, a non-Next React template still makes sense for projects using a lighter stack or custom architecture. The agent posture should focus on:

- component composition and hook correctness
- `useEffectEvent` for effect/event separation where appropriate
- `startTransition` and `useDeferredValue` for non-blocking UI updates
- accessibility and testability of interactive components
- avoiding framework-specific assumptions

### Next.js

Next.js deserves its own template because App Router projects are not just "React plus routing". The official docs center the stack around:

- Server Components by default
- explicit Client Component boundaries
- Route Handlers inside `app/`
- caching and dynamic rendering behavior
- streaming and navigation performance

That requires stack-specific agent judgment about where code should run, how data should be fetched, and which parts belong to server or client boundaries.

### Astro

Astro also deserves a separate template because its architecture is materially different:

- islands architecture instead of broad client hydration
- content collections as a primary content model
- integration- and adapter-driven capabilities
- clear static, hybrid, and server rendering modes

An Astro-aware agent should optimize for small client payloads, limited hydration, content typing, and the right adapter or integration strategy.

## Recommended New Subagent: Infrastructure / Cloud / SRE

This is the biggest missing specialist in the current topology.

Recommended new agent:

- `infra-cloud-sre`

Suggested scope:

- Dockerfile design and hardening
- `docker compose` for local dev, staging, and single-host production
- production compose overlays, health checks, restart policies, and remote-host flows
- reverse proxy setup with NGINX
- API gateway modeling with Kong Gateway: Services, Routes, Plugins
- VPS bootstrap and deployment topology
- VPN-aware network concerns and private service exposure
- CI/CD design with GitHub Actions
- environments, approvals, secret boundaries, and rollback paths
- operational runbooks, observability, and release risk assessment

Suggested ownership boundaries:

- owns infra files, deployment manifests, proxy config, gateway config, and workflows
- does not own business logic implementation unless the task is explicitly cross-cutting
- should work with `implementer`, `validator`, and `delivery` rather than replacing them

Example files this agent should own when present:

- `Dockerfile`
- `docker-compose.yml`, `compose.yaml`, `compose.production.yaml`
- `.github/workflows/*`
- `nginx.conf`, `nginx/*.conf`
- `kong.yaml`, decK config, gateway bootstrap scripts
- deployment scripts and runbooks

## Suggested Next Agent Map

If the project evolves toward stack-aware templates, a better map would be:

| Layer | Agents |
| --- | --- |
| Core routing | `orchestrator` |
| Cross-project execution | `explorer`, `implementer`, `validator`, `reviewer`, `delivery` |
| Frontend React | `react-ui`, `react-state-performance`, `react-testing-a11y` |
| Frontend Next.js | `nextjs-app-router`, `nextjs-server-runtime`, `nextjs-data-cache` |
| Frontend Astro | `astro-islands-content`, `astro-integrations`, `astro-ssr-adapter` |
| Platform | `infra-cloud-sre` |

That keeps the current core topology intact while making the stack-specific knowledge explicit instead of overloading a generic implementer.

## Repository Layout

```text
.
├── .opencode/
│   ├── agents/
│   ├── commands/
│   ├── plugins/
│   ├── policies/
│   └── skills/
├── docs/
├── openspec/
├── scripts/
├── AGENTS.md.template
├── README.md
└── tui.json
```

## Local Docs

- [Getting Started](docs/getting-started.md)
- [OpenSpec Overview](openspec/README.md)
- [AGENTS Template](AGENTS.md.template)

## External References Used For Template Direction

- React: [Creating a React App](https://react.dev/learn/start-a-new-react-project), [Installation](https://react.dev/learn/installation), [useEffectEvent](https://react.dev/reference/react/useEffectEvent), [startTransition](https://react.dev/reference/react/startTransition), [useDeferredValue](https://react.dev/reference/react/useDeferredValue)
- Next.js: [App Router](https://nextjs.org/docs/app), [Server and Client Components](https://nextjs.org/docs/app/getting-started/server-and-client-components), [Route Handlers](https://nextjs.org/docs/app/getting-started/route-handlers)
- Astro: [Islands Architecture](https://docs.astro.build/en/concepts/islands/), [Content Collections](https://docs.astro.build/en/guides/content-collections/), [Working with Integrations](https://docs.astro.build/en/guides/integrations/), [On-demand Rendering](https://docs.astro.build/en/guides/on-demand-rendering/)
- Docker Compose: [Quickstart](https://docs.docker.com/compose/gettingstarted/), [Use Compose in production](https://docs.docker.com/compose/how-tos/production/), [Compose Develop Specification](https://docs.docker.com/reference/compose-file/develop/)
- NGINX: [Reverse Proxy](https://docs.nginx.com/nginx/admin-guide/web-server/reverse-proxy)
- Kong Gateway: [Kong Gateway Overview](https://developer.konghq.com/gateway/), [Gateway Services](https://developer.konghq.com/gateway/entities/service/), [Routes](https://developer.konghq.com/gateway/entities/route/), [Plugins](https://developer.konghq.com/gateway/entities/plugin/)
- GitHub Actions: [Deployment environments](https://docs.github.com/en/actions/concepts/workflows-and-actions/deployment-environments)

## License

MIT
