# Pi Package Bridge V1 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add Pi package management, slash-command completion, and basic extension UI dialogs to Pi Desktop.

**Architecture:** Keep Pi as the package source of truth. A focused Go `PackageManager` invokes the official Pi CLI for package operations, Wails exposes package and RPC response methods, and Vue adds Codex-style compact UI surfaces inside the existing settings panel and message composer.

**Tech Stack:** Go 1.22+, Wails v2, Vue 3, Pinia, TypeScript, Vite, Vitest, Vue Test Utils, jsdom.

---

## File Map

- Create `backend/piagent/executable.go`: shared Pi executable lookup and environment helpers.
- Create `backend/piagent/packages.go`: package CLI runner, scope handling, validation, and structured results.
- Create `backend/piagent/packages_test.go`: package manager red-green coverage.
- Modify `backend/piagent/client.go`: use shared executable helper.
- Modify `backend/piagent/types.go`: extension UI request and response protocol fields.
- Modify `app.go`: package Wails methods, Agent restart recovery, and extension UI response method.
- Create `frontend/src/components/ExtensionPackagesPanel.vue`: compact Codex-style package settings surface.
- Create `frontend/src/components/ExtensionUIDialog.vue`: blocking extension dialogs.
- Create `frontend/src/components/ExtensionNotifications.vue`: non-blocking notices.
- Create `frontend/src/components/SlashCommandMenu.vue`: composer command completion.
- Create `frontend/src/composables/useExtensionUI.ts`: dialog queue and notification state.
- Modify `frontend/src/components/SettingsPanel.vue`: add the Extensions tab.
- Modify `frontend/src/components/InputBox.vue`: slash completion.
- Modify `frontend/src/App.vue`: render extension dialogs and notifications.
- Modify `frontend/src/composables/usePiAgent.ts`: package API wrappers, command loading, and extension event dispatch.
- Modify `frontend/src/types/index.ts`: package, command provenance, and extension UI types.
- Modify `frontend/src/env.d.ts`: complete Wails binding declarations.
- Modify `frontend/package.json`: add Vitest scripts and test dependencies.
- Modify `frontend/vite.config.ts`: configure jsdom tests.
- Create `frontend/src/**/*.test.ts`: component and composable tests.

### Task 1: Repair Baseline And Add Vitest

**Files:**
- Modify: `frontend/package.json`
- Modify: `frontend/vite.config.ts`
- Modify: `frontend/src/env.d.ts`
- Test: `frontend/src/test/smoke.test.ts`

- [ ] Add `"test": "vitest run"` and `"test:watch": "vitest"` scripts.
- [ ] Configure `test.environment = "jsdom"` in Vite config.
- [ ] Add the existing Wails bindings missing from `env.d.ts`: `EnsureSessionNamed`, `GetAuthKeys`, and `SetApiKey`.
- [ ] Write `src/test/smoke.test.ts`:

```ts
import { describe, expect, it } from 'vitest'

describe('frontend test harness', () => {
  it('runs in jsdom', () => {
    expect(document.createElement('div').tagName).toBe('DIV')
  })
})
```

- [ ] Run `npm run test`, `npm run typecheck`, and `npm run build`.
- [ ] Run `go mod tidy`, then run `go test ./...` after the frontend build creates `frontend/dist`.
- [ ] Commit baseline test infrastructure.

### Task 2: Add A Tested Pi Package Manager

**Files:**
- Create: `backend/piagent/executable.go`
- Create: `backend/piagent/packages.go`
- Create: `backend/piagent/packages_test.go`
- Modify: `backend/piagent/client.go`

- [ ] Write failing table tests for CLI argument construction:

```go
func TestPackageArgs(t *testing.T) {
    tests := []struct {
        name string
        op PackageOperation
        want []string
    }{
        {"project install", PackageOperation{Action: "install", Source: "npm:demo", Scope: ScopeProject}, []string{"install", "npm:demo", "-l"}},
        {"global remove", PackageOperation{Action: "remove", Source: "npm:demo", Scope: ScopeGlobal}, []string{"remove", "npm:demo"}},
        {"single update", PackageOperation{Action: "update", Source: "npm:demo"}, []string{"update", "--extension", "npm:demo"}},
        {"all updates", PackageOperation{Action: "update"}, []string{"update", "--extensions"}},
        {"list", PackageOperation{Action: "list"}, []string{"list"}},
    }
}
```

- [ ] Write failing tests for empty source rejection, invalid project directory rejection, missing local path rejection, timeout errors, and stderr propagation.
- [ ] Extract executable resolution from `client.go` into `executable.go`. Cover PATH lookup plus Windows npm and pnpm global locations.
- [ ] Implement `PackageManager.Run(ctx, cwd, operation)` with injected command execution for tests, a 60-second default timeout, `-l` for project installs/removals, and structured JSON-friendly results.
- [ ] Keep list output as a structured array of display rows parsed from `pi list`; preserve raw output when a line cannot be parsed so the UI remains useful across Pi CLI versions.
- [ ] Run `go test ./backend/piagent`.
- [ ] Commit package manager.

### Task 3: Expose Package Operations And Agent Restart Recovery

**Files:**
- Modify: `app.go`
- Create: `app_test.go`

- [ ] Write failing tests around a small injected package runner and Agent starter interface:
  - successful package changes restart with the same `cwd` and session path;
  - failed package changes do not restart;
  - retry startup reuses the last restart context.
- [ ] Add Wails methods:

```go
func (a *App) ListPackages(scope string) (string, error)
func (a *App) InstallPackage(source string, scope string) (string, error)
func (a *App) RemovePackage(source string, scope string) (string, error)
func (a *App) UpdatePackage(source string) (string, error)
func (a *App) UpdateAllPackages() (string, error)
func (a *App) RetryAgentStartup() error
```

- [ ] Capture the active session through `get_state` before stopping the old client. Preserve the current working directory.
- [ ] Restart only after successful mutating package operations.
- [ ] Return restart status separately so Vue can explain package success with Agent reload failure.
- [ ] Run `go test ./...`.
- [ ] Commit Wails package operations.

### Task 4: Add RPC Extension UI Bridge

**Files:**
- Modify: `backend/piagent/types.go`
- Modify: `app.go`
- Create: `backend/piagent/types_test.go`

- [ ] Write failing JSON round-trip tests for official RPC request fields:

```go
{
  "type": "extension_ui_request",
  "id": "uuid-1",
  "method": "select",
  "title": "Allow?",
  "options": ["Allow", "Block"],
  "timeout": 10000
}
```

- [ ] Add fields for `id`, `method`, `title`, `message`, `options`, `timeout`, `placeholder`, `prefill`, and `notifyType`.
- [ ] Add extension UI response fields to `RPCCommand`: `value`, `confirmed`, and `cancelled`.
- [ ] Add:

```go
func (a *App) RespondToExtensionUI(id string, value string, confirmed bool, cancelled bool) error
```

- [ ] Send `extension_ui_response` asynchronously because the protocol does not return a normal command response.
- [ ] Run `go test ./...`.
- [ ] Commit RPC UI bridge.

### Task 5: Build Codex-Style Package Settings

**Files:**
- Create: `frontend/src/components/ExtensionPackagesPanel.vue`
- Create: `frontend/src/components/ExtensionPackagesPanel.test.ts`
- Modify: `frontend/src/components/SettingsPanel.vue`
- Modify: `frontend/src/composables/usePiAgent.ts`
- Modify: `frontend/src/types/index.ts`
- Modify: `frontend/src/env.d.ts`

- [ ] Write failing Vitest component tests for:
  - default `project` scope;
  - explicit switch to `global`;
  - install security confirmation;
  - disabled controls while an operation runs;
  - error feedback;
  - package rows with update and remove actions.
- [ ] Implement a compact Settings tab matching the existing panel:
  - quiet neutral surfaces;
  - one primary blue install action;
  - compact package rows;
  - subtle status text;
  - `Browse Package Catalog` link.
- [ ] Add composable Wails wrappers for list/install/remove/update-all/update-one/retry-startup.
- [ ] Refresh the list after every successful operation.
- [ ] Run `npm run test`, `npm run typecheck`, and `npm run build`.
- [ ] Commit package settings UI.

### Task 6: Add Slash Completion And Basic Extension Dialogs

**Files:**
- Create: `frontend/src/components/SlashCommandMenu.vue`
- Create: `frontend/src/components/SlashCommandMenu.test.ts`
- Create: `frontend/src/components/ExtensionUIDialog.vue`
- Create: `frontend/src/components/ExtensionUIDialog.test.ts`
- Create: `frontend/src/components/ExtensionNotifications.vue`
- Create: `frontend/src/composables/useExtensionUI.ts`
- Create: `frontend/src/composables/useExtensionUI.test.ts`
- Modify: `frontend/src/components/InputBox.vue`
- Modify: `frontend/src/composables/usePiAgent.ts`
- Modify: `frontend/src/App.vue`
- Modify: `frontend/src/types/index.ts`
- Modify: `frontend/src/env.d.ts`

- [ ] Write failing slash menu tests for `/` opening, text filtering, ArrowUp/ArrowDown movement, Enter selection, and Escape dismissal.
- [ ] Add `getCommands()` wrapper using the existing `GetCommands()` Wails method. Treat command loading errors as an empty list.
- [ ] Render the menu directly above the composer with compact Codex-style rows and source badges.
- [ ] Write failing extension UI tests for:
  - FIFO dialog queue;
  - `select`, `confirm`, `input`, and `editor`;
  - cancellation response;
  - unknown blocking method cancellation;
  - non-blocking `notify`.
- [ ] Implement `useExtensionUI()` state and event handling.
- [ ] Route `extension_ui_request` from `usePiAgent.ts` to the UI composable.
- [ ] Render one active dialog and non-blocking notifications from `App.vue`.
- [ ] Clear pending dialogs during Agent stop, abort, and session switching.
- [ ] Run `npm run test`, `npm run typecheck`, and `npm run build`.
- [ ] Commit slash completion and extension UI.

### Task 7: Verify The Integrated Desktop App

**Files:**
- Review only.

- [ ] Run `go test ./...`.
- [ ] Run `npm run test`.
- [ ] Run `npm run typecheck`.
- [ ] Run `npm run build`.
- [ ] Start Wails through `powershell -ExecutionPolicy Bypass -File .\scripts\restart-wails.ps1`.
- [ ] Verify `http://127.0.0.1:5173` and the Wails dev URL return HTTP 200.
- [ ] Open the desktop app and manually inspect:
  - Codex-style Extensions tab;
  - slash menu opening from `/`;
  - package operation confirmation;
  - recoverable empty/error states.
- [ ] Review `git diff --stat` and `git status --short`.

