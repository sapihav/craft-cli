---
title: M4 — `--format mcp` output mode (drop-in swap)
type: task
priority: P3
status: todo
created: 2026-04-18
---

# M4 — `--format mcp` output mode

## Problem Statement

Agents that want to switch from the Craft MCP to this CLI must currently adapt to two different JSON shapes. A `--format mcp` mode on every subcommand would emit the exact MCP JSON-RPC `content[].text` envelope, making the CLI a zero-shape-change replacement.

## Acceptance Criteria

- Global `--format json|mcp|pretty` flag (default `json`). `pretty` remains equivalent to current `--pretty`.
- With `--format mcp`: stdout emits `{content:[{type:"text", text:"<JSON string of result>"}]}` matching the MCP tool response shape.
- `docs/llm/output-parity.md` (existing) gets a per-tool mapping table showing CLI command → MCP tool name → shared payload.
- Unit tests for at least 5 commands asserting `--format mcp` output matches the MCP envelope.

## Context / Notes

- Purely additive — no breaking change to default output.
- Depends on: M1–M3 ideally (so the table is complete), but technically can ship first.
- Budget: ~150 LoC (one helper function + per-command output wrapper).
- Polish milestone — do AFTER functional gaps are closed.
