# Kimi Code CLI support

The `kimi` agentic-system plugin targets the maintained Kimi Code CLI. Its
source version used to verify the launch surface is 2.6.3.

## Supported

- Interactive and prompt launches pass `--model <alias>` and use the requested
  working directory. Other interactive native arguments are forwarded verbatim
  after the model and module-owned permission flag; the model and prompt
  options must use their typed launch fields.
- Interactive resume accepts `--continue`/`-c`/`-C` or one identified
  `--session <id>`/`-S <id>` or `--resume <id>`/`-r <id>` selector. A bare
  optional session selector is refused because Kimi opens an interactive
  picker; repeated or conflicting selectors are refused.
- Prompt launches use `--model <alias> --prompt <text>`. A prompt file is read
  by the launch planner and supplied as the prompt text. Dry-run uses a
  `<prompt>` placeholder and does not read that file.
- The default data home is `~/.kimi-code`; the CLI reads `KIMI_CODE_HOME` to
  select another home. Kimi-specific global instructions are
  `$KIMI_CODE_HOME/AGENTS.md` (default `~/.kimi-code/AGENTS.md`); project
  instructions include `AGENTS.md` and `.kimi-code/AGENTS.md`. Kimi Code also
  reads generic instructions from `~/.agents/AGENTS.md` when present.
- MCP configuration is read from `$KIMI_CODE_HOME/mcp.json` and
  `.kimi-code/mcp.json`. The CLI has no per-launch MCP configuration flag, so
  the plugin declares no launch-composition grammar.
- `PermissionModeYolo` maps to `--yolo` only when `kimi --version` establishes
  release 2.6.3. Unknown, missing or unparseable releases refuse yolo. Native
  permission mode adds no flag. The CLI's `--yolo`, `-y`, `--yes` and
  `--auto-approve` spellings select yolo; callers must use the typed permission
  field, so those raw suffix forms are refused.

## Not represented by this integration

- Kimi's `--auto` (Never Ask) posture is not part of the shared permission-mode
  API and is refused in interactive native arguments.
- Model aliases are loaded from Kimi configuration. The plugin forwards an
  opaque non-empty alias and cannot identify an unknown configured alias
  without consulting user configuration; Kimi itself decides whether it is
  valid.
- Reasoning effort, goals, budgets, service tiers and per-launch MCP
  composition are not supported.
- Yolo is not mapped for releases other than 2.6.3 until the permission flag is
  verified for those releases.

References: [Kimi command reference](https://moonshotai.github.io/kimi-code/en/reference/kimi-command),
[sessions](https://moonshotai.github.io/kimi-code/en/guides/sessions),
[data locations](https://moonshotai.github.io/kimi-code/en/configuration/data-locations),
[agents](https://moonshotai.github.io/kimi-code/en/customization/agents), and
[MCP configuration](https://moonshotai.github.io/kimi-code/en/customization/mcp).
