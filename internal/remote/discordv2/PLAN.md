# discordv2 - Discord Bot Integration Overhaul

## Overview

A clean rewrite of the Discord bot integration (`internal/remote/discord`) with improved
separation of concerns, testability, and maintainability. Uses `github.com/bwmarrin/discordgo` v0.29.0.

## Architecture

```
discordv2/
  bot.go              -- Bot lifecycle (New, Start, Stop), session management
  commands.go         -- Command routing and handler implementations
  events.go           -- event.Handler implementation, event filtering, publishing
  embeds.go           -- All embed building logic (items, stats, drops, help, list)
  sender.go           -- Message delivery abstraction (bot API vs webhook)
  webhook.go          -- Webhook HTTP client
```

### Design Principles

- **Single responsibility per file**: lifecycle, commands, events, formatting, and delivery are separate.
- **Sender interface**: A `MessageSender` interface abstracts bot-API and webhook delivery so
  command handlers and event publishers don't branch on `useWebhook` everywhere.
- **No global config reads inside handlers**: Configuration is passed at construction time or via
  the `Options` struct. This makes the code testable without a loaded config singleton.
- **Context-driven lifecycle**: `Start` blocks on `context.Context`; cancellation triggers cleanup.

## Implementation Phases

### Phase 1 - Scaffold & Lifecycle (current)
- [x] `bot.go`: `Bot` struct, `New()`, `Start()`, `Stop()` with discordgo session management
- [x] `sender.go`: `MessageSender` interface with `Send`, `SendEmbed`, `SendFile` methods
- [x] `webhook.go`: Webhook-based `MessageSender` implementation
- [x] `events.go`: Stub `Handle(ctx, event.Event) error` that satisfies `event.Handler`

### Phase 2 - Event Publishing
Implement the `Handle` method to publish events to Discord, matching all current v1 behavior:

- [x] `GameCreatedEvent` -> text message with game name/password
- [x] `GameFinishedEvent` -> text message (filtered by reason + config flags)
- [x] `RunStartedEvent` -> text message
- [x] `RunFinishedEvent` -> text message with run name and reason
- [x] `NgrokTunnelEvent` -> text message
- [x] `ItemStashedEvent` -> screenshot OR embed (based on config flag)
- [x] Generic events with images -> JPEG screenshot attachment
- [x] `shouldPublish` filtering logic (respects all Discord config toggles)
- [x] JPEG encoding for image events (80% quality, matching v1)

### Phase 3 - Item Embeds
Port the item stash embed builder from v1, keeping full feature parity:

- [x] Item name (identified name preferred)
- [x] Item type and quality badge
- [x] Defense stat
- [x] All Attributes combined display
- [x] All Resistances combined display (including partial 3-of-4 detection)
- [x] Enhanced damage display
- [x] Elemental damage lines (Fire, Lightning, Cold, Magic, Poison)
- [x] Individual stat lines (excluding internal stat IDs)
- [x] Socket count and Ethereal flag
- [x] Pickit rule info (file + line, controlled by config flag)
- [x] Quality-based embed color coding
- [x] Supervisor name and timestamp footer

### Phase 4 - Command Handlers
Port all 7 commands with identical behavior:

- [x] `!help` -> embed with command reference
- [x] `!list` -> embed with all supervisors, status indicators, uptime
- [x] `!start <sup1> [sup2] ...` -> start supervisors, confirmation messages
- [x] `!stop <sup1> [sup2] ...` -> stop supervisors, confirmation messages
- [x] `!status <sup1> [sup2] ...` -> supervisor status text
- [x] `!stats <sup1> [sup2] ...` -> embed with games, drops, deaths, chickens, errors, uptime
- [x] `!drops <supervisor> [count]` -> embed with recent drops (default 5, max 20), quality emojis

Command infrastructure:
- [x] Message routing (prefix-based `!command` dispatch)
- [x] Admin authorization (check `BotAdmins` list)
- [x] Unknown command response with help hint
- [x] Supervisor existence validation helper

### Phase 5 - Integration & Wiring
- [x] `Bot.Handle` satisfies `event.Handler` signature: `func(ctx context.Context, e event.Event) error`
- [x] `Options` struct accepts the full v1 config surface
- [x] `managerAdapter` (in `cmd/koolo/discord_adapter.go`) bridges `*bot.SupervisorManager` → `SupervisorControl`
- [x] `config.KooloCfg.Discord.UseV2` toggle selects v1 or v2 at startup
- [x] `cmd/koolo/main.go` conditionally wires v1 or v2 based on config
- [x] HTTP endpoint `GET /api/discord/version` reports active version

## Current v1 Feature Inventory

### Config Fields Used
| Field | Purpose |
|-------|---------|
| `Discord.Enabled` | Master toggle (checked externally) |
| `Discord.Token` | Bot token for discordgo session |
| `Discord.ChannelID` | Main notification channel |
| `Discord.ItemChannelID` | Optional separate channel for item drops |
| `Discord.BotAdmins` | User IDs authorized to run commands |
| `Discord.UseWebhook` | Toggle webhook mode (no commands, send-only) |
| `Discord.WebhookURL` | Main webhook URL |
| `Discord.ItemWebhookURL` | Item-specific webhook URL |
| `Discord.EnableGameCreatedMessages` | Toggle game creation notifications |
| `Discord.EnableNewRunMessages` | Toggle run start notifications |
| `Discord.EnableRunFinishMessages` | Toggle run finish notifications |
| `Discord.EnableDiscordErrorMessages` | Toggle error finish notifications |
| `Discord.EnableDiscordChickenMessages` | Toggle death/chicken finish notifications |
| `Discord.DisableItemStashScreenshots` | Use embeds instead of screenshots for items |
| `Discord.IncludePickitInfoInItemText` | Show pickit rule info in item embeds |

### Dependencies
- `github.com/bwmarrin/discordgo` v0.29.0
- `github.com/hectorgimenez/koolo/internal/bot` (`SupervisorManager`, `Stats`, `SupervisorStatus`)
- `github.com/hectorgimenez/koolo/internal/event` (`Event`, `Handler`, all event types)
- `github.com/hectorgimenez/koolo/internal/config` (`Koolo.Discord.*`)
- `github.com/hectorgimenez/d2go/pkg/data` (`Drop`, `Item`, quality types)
- `github.com/hectorgimenez/d2go/pkg/data/stat` (stat IDs for item display)
