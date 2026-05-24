# Result Contract Envelope V1

Use this YAML envelope for substantive routed agent handoffs, context recovery, and final status summaries. Keep simple T3 Express work compact with `not_applicable`; do not invent command evidence.

```yaml
schema_version: result-contract/v1
status: ready | implemented | validated | delivery_pending | sync_pending | blocked | escalated | delivered | closed
legacy_fallback: false
executive_summary: <1-3 line summary>
artifacts:
  changed:
    - <path or none>
  referenced:
    - <path/spec/PR or none>
evidence:
  commands:
    - command: <command or none>
      result: passed | failed | blocked | not_run
      notes: <key output or reason>
  static:
    - <manual/static evidence or not_applicable>
workflow_decision:
  tier: T1 | T2 | T3 | not_applicable
  workflow_limits_source: workflow_limits | not_available
  workflow_limits_paths:
    sizing: workflow_limits.sizing | not_available
    routing: workflow_limits.routing | not_available
    proposal_slicing: workflow_limits.proposal_slicing_strategy | not_available
    delivery_batching: workflow_limits.delivery_batch_strategy | not_applicable
    preflight: workflow_limits.preflight | not_available
    stop_rules: workflow_limits.stop_rules | not_available
  workload_decision_needed: true | false
  review_slices:
    - <slice summary or not_applicable>
  preflight_status: passed | blocked | not_applicable | not_available
  stop_rule_status: clear | triggered | not_applicable | not_available
  delivery_batching_guidance: <guidance or not_applicable>
risks:
  - <risk/follow-up or none>
next_recommended:
  owner: orchestrator | explorer | implementer | validator | reviewer | delivery | user | none
  action: <next action>
skill_resolution:
  local_skills_used:
    - <skill or none>
  bootstrap_recommended: true | false
  notes: <notes or none>
```
