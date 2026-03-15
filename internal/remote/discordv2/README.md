# discordv2 - Discord Bot Integration (v2)

A clean rewrite of the Discord bot integration (`internal/remote/discord`) with
improved separation of concerns, testability, and maintainability.

## Features

### Event Publishing
All events from the bot supervisor pipeline are forwarded to Discord when
enabled:

| Event | Behavior |
|-------|----------|
| `GameCreatedEvent` | Text message with game name and password |
| `GameFinishedEvent` | Text message (filtered by reason + config flags) |
| `RunStartedEvent` | Text message with run name |
| `RunFinishedEvent` | Text message with run name and reason |
| `NgrokTunnelEvent` | Text message with tunnel URL |
| `ItemStashedEvent` | Rich embed **or** JPEG screenshot (configurable) |
| Any event with an image | JPEG screenshot attachment (80% quality) |

Each event type is independently togglable via the `Options` struct.

### Item Embeds
When `DisableItemStashScreenshots` is `true`, item drops produce a rich embed
with:

- Identified name (preferred) or base name
- Item type and quality badge
- Defense stat
- Combined "All Attributes" and "All Resistances" display (including 3-of-4 partial detection)
- Enhanced damage percentage
- Elemental damage lines (Fire, Lightning, Cold, Magic, Poison)
- Individual stat lines (internal-only stat IDs filtered out)
- Socket count and Ethereal flag
- Pickit rule info (file + line number, controlled by config flag)
- Quality-based embed color coding
- Supervisor name and timestamp footer

### Commands (Bot-Token Mode)
When running with a bot token (not webhook), the bot responds to `!`-prefixed
commands from authorized admins:

| Command | Description |
|---------|-------------|
| `!help` | Embed with command reference |
| `!list` | All supervisors with status indicators and uptime |
| `!start <sup> [sup2] ...` | Start one or more supervisors |
| `!stop <sup> [sup2] ...` | Stop one or more supervisors |
| `!status <sup> [sup2] ...` | Current status text for supervisors |
| `!stats <sup> [sup2] ...` | Embed with games, drops, deaths, chickens, errors, uptime |
| `!drops <sup> [count]` | Embed with recent drops (default 5, max 20), quality emojis |

Commands are restricted to user IDs listed in `BotAdmins`.

### Delivery Modes
A `MessageSender` interface abstracts delivery, supporting two modes:

- **Bot-API mode** (`sessionSender`): Uses `discordgo.Session` to send messages
  via the Discord gateway. Supports receiving commands.
- **Webhook mode** (`webhookSender`): HTTP POST to a webhook URL. Send-only, no
  command handling. Supports separate item webhook URL.

## Differences from v1

| Aspect | v1 (`internal/remote/discord`) | v2 (`internal/remote/discordv2`) |
|--------|------|------|
| **Config access** | Reads `config.Koolo` globals inside handlers | All config passed via `Options` struct at construction |
| **Delivery abstraction** | Inline `if useWebhook` branches throughout | `MessageSender` interface; handlers are delivery-agnostic |
| **Manager dependency** | Direct `*bot.SupervisorManager` import (pulls in Windows-only packages) | `SupervisorControl` interface with local `SupervisorStats` type; no `bot` package import |
| **File layout** | Two large files (`discord_bot.go`, `discord_event_handler.go`) | Single-responsibility files: `bot.go`, `commands.go`, `events.go`, `embeds.go`, `sender.go`, `webhook.go`, `util.go` |
| **Testability** | Difficult to test; requires a live `discordgo.Session` and `bot.SupervisorManager` | Embed builders are pure functions; `spySender` and `mockManager` enable full unit testing without Discord or Windows dependencies |
| **Cross-platform** | Cannot compile tests on Linux/macOS due to `bot` package transitively importing `lxn/win` | Package compiles on all platforms; tests require cross-compilation only because the `event` package imports `config` which imports `x/sys/windows` |

## File Layout

```
discordv2/
  bot.go              -- Bot struct, New(), Start(), Stop(), message routing
  commands.go         -- Command handlers + embed builders (stats, list, help, drops)
  events.go           -- event.Handler implementation, shouldPublish(), screenshot encoding
  embeds.go           -- Item stash embed builder with full stat formatting
  sender.go           -- MessageSender interface + sessionSender (bot-API delivery)
  webhook.go          -- webhookSender (HTTP webhook delivery)
  util.go             -- Shared helpers (newByteReader, contentTypeForFile)
  PLAN.md             -- Implementation plan with phase tracking
  README.md           -- This file
```

## Enabling v2

Set `useV2: true` under the `discord:` section in your Koolo YAML config:

```yaml
discord:
  enabled: true
  useV2: true           # Switch to v2 implementation
  token: "your-token"
  channelId: "123456"
  # ... all other discord settings remain the same
```

The default (`false` or omitted) keeps the original v1 implementation active.
No other configuration changes are needed -- v2 reads the same config fields as
v1.

### HTTP API

When running, you can check which Discord version is active:

```
GET /api/discord/version
```

Returns:
```json
{
  "enabled": true,
  "version": "v2",
  "useV2": true
}
```

## Testing

The discordv2 package has comprehensive unit tests covering all embed builders,
event routing, config-based filtering, and command logic. Tests use mock
implementations (`spySender`, `mockManager`) rather than live Discord
connections.

### Test files

| File | Coverage |
|------|----------|
| `bot_test.go` | Constructor, `getItemSender` fallback logic |
| `events_test.go` | `shouldPublish` filtering, `Handle` routing for all event types |
| `embeds_test.go` | Item stash embed: stats, defense, attributes, resists, damage, sockets, ethereal, pickit info, quality colors |
| `commands_test.go` | `buildHelpEmbed`, `buildStatsEmbed`, `buildListEmbed`, `buildDropsEmbed`, `dropQualityEmoji`, `supervisorExists` |
| `util_test.go` | `contentTypeForFile` MIME detection |
| `spy_sender_test.go` | Test doubles: `spySender`, `newTestBot`, `newTestBotNoItemSender` |
| `mock_manager_test.go` | Test double: `mockManager` implementing `SupervisorControl` |
| `testdata_test.go` | Test data builders: `makeTestDrop`, `makeTestDropWithStats`, etc. |

### Running tests

Because the `event` package transitively imports `x/sys/windows` (via `config`
-> `utils`), tests must be cross-compiled for Windows. If you're on WSL, the
compiled test binary runs natively via WSL interop:

```bash
# Cross-compile the test binary
GOOS=windows GOARCH=amd64 go test -c -o /tmp/discordv2_test.exe ./internal/remote/discordv2/

# Run all tests with verbose output
/tmp/discordv2_test.exe -test.v -test.count=1

# Run a specific test
/tmp/discordv2_test.exe -test.v -test.run TestBuildItemStashEmbed
```

### Verifying the full build

To verify the entire project compiles with v2 wired in:

```bash
GOOS=windows GOARCH=amd64 go build -o /tmp/koolo_test.exe ./cmd/koolo/
```
