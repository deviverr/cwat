# Setup Migration Status (OG -> Go)

This file tracks setup/onboarding parity between:

- OG reference (read-only): C:/Users/dedpu/Documents/CWAT_code_version_1_0_33
- Go rewrite: go/cwat-go-client

No OG files were modified.

## Ported in Go

- Interactive setup command: `cwat setup`
- Auto-prompt setup from `cwat`/`cwat chat` when API key env is missing
- Profile selection with readiness indicators
- API key onboarding choices:
  - Paste key for current cwat run
  - Show PowerShell session command
  - Show persistent `setx` command
- Theme preference step (dark/light)
- Syntax highlighting preference toggle
- Trust preference step (bash, MCP, hooks, env)
- Network preflight step (provider endpoint reachability)
- Terminal setup hints step
- Setup completion persistence in config (`setup.completed`)
- API key approval hint persistence (masked key hint, no full key storage)

## Still Missing (from OG setup ecosystem)

- Full React/Ink modal UX parity from OG onboarding screens
- OAuth login flow + browser auth onboarding
- Terminal auto-configuration command integration (actual config writes)
- Project onboarding state machine (`CWAT.md` guided tasks)
- Rich trust dialog behavior wired to runtime tool permissions
- Setup analytics events and telemetry hooks
- Security notes and legal/consent step parity

## Next Priority Ports

1. Wire `setup.permissions` into runtime permission enforcement.
2. Add OAuth provider onboarding path.
3. Add project onboarding wizard for `CWAT.md` initialization.
4. Add terminal integration commands (`cwat terminal-setup`).
