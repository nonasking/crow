# crow

A webhook server that parses card payment SMS and auto-logs them to a Notion budget tracker.

[![CI](https://github.com/nonasking/crow/actions/workflows/ci.yml/badge.svg?branch=develop)](https://github.com/nonasking/crow/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/nonasking/crow/branch/develop/graph/badge.svg)](https://codecov.io/gh/nonasking/crow)
[![Go Version](https://img.shields.io/github/go-mod/go-version/nonasking/crow)](go.mod)

## Why

Manually adding each card transaction to my Notion budget page got old fast. The payment SMS already carries everything I need — amount, merchant, date — so the natural move is to parse it on arrival and write it straight to Notion.

An SMS forwarder app on my phone POSTs the message to this server. A regex parser dispatches on issuer, extracts the fields, and creates a page via the Notion API. 30~40 manual entries per month, gone.

## Architecture

```mermaid
flowchart LR
    SMS[Card payment SMS] --> FWD[SMS forwarder app]

    subgraph crow["crow (Gin)"]
        direction LR
        H["handler<br/>POST /webhook"]
        P{"parser<br/>ParseWebhookAuto"}
        PS["parseShinhanCard<br/>old + new format"]
        PW[parseWooriCard]
        CR{{"CardRecorder<br/>(interface)"}}
        NC["notion.Client<br/>CreateCardRecord"]

        H -- "parse" --> P
        P -->|Shinhan| PS
        P -->|Woori| PW
        PS -- "amount, place,<br/>issuer, date" --> H
        PW -- "amount, place,<br/>issuer, date" --> H
        H -- "record" --> CR
        CR -.implemented by.-> NC
    end

    FWD -- "POST {message}" --> H
    NC --> DB[(Notion DB<br/>budget tracker)]
```

Three layers, separated by intent:

- `internal/parser` — issuer-specific SMS parsers. Pure functions, no external dependencies.
- `internal/notion` — Notion API client. Maps parsed fields to page properties and POSTs.
- `internal/handler` — Gin handler. Wires the two together.

The handler depends on the `CardRecorder` interface, not on the Notion client directly — so both the parser and the handler's error paths (bad JSON, unparseable SMS, Notion down) are unit-testable with a fake recorder and no network.

## Quick start

```bash
cp .env.example .env
# fill in NOTION_TOKEN, NOTION_DATABASE_ID, NOTION_VERSION

go run ./cmd/server
```

Defaults to port `8080`. Override with `PORT`.

Request shape:

```bash
curl -X POST http://localhost:8080/webhook \
  -H "Content-Type: application/json" \
  -d '{"message": "[Web발신]\n1차 민생회복 신한(4557)승인 강*성 16,980원 08/24 17:50 땀땀 잔액 0원"}'
```

## Adding a new card issuer

Three places to touch:

1. Add `parseXxxCard(msg) (...)` in `internal/parser/message_parser.go`
2. Add a switch arm to `ParseWebhookAuto`
3. Add fixture cases to `internal/parser/message_parser_test.go`

The handler and Notion client stay untouched by design.

## Tradeoffs

- **Regex-based parser** — breaks if an issuer changes their SMS format (Shinhan did once, which is why there are dual old/new format branches). Mitigated by accumulating real SMS samples as fixtures in `*_test.go`. CI goes red before production does.
- **Notion as the data store** — zero infra cost. In exchange, throughput is bounded by Notion's rate limit (~3 req/sec average), so concurrent bursts are a non-goal. Fine for personal use.
- **Render free tier deploy** — 5~10s cold start. Hidden inside the natural delay between swiping a card and the SMS arriving, so it's not noticeable.

## Layout

```
cmd/server/        # entry point
internal/
  config/          # .env loading and validation
  constants/       # issuer names, error messages
  handler/         # /webhook handler
  notion/          # Notion API client
  parser/          # issuer-dispatching SMS parser + tests
.github/workflows/ # CI
```

## Testing

```bash
go test ./... -cover
```

**8 test functions, 50 cases.**

| Package | Cases | What it covers |
|---|---|---|
| `internal/handler` | 14 | Success per issuer (parsed fields reach the recorder), bad JSON / wrong field type / missing `message`, unparseable SMS, Notion failure → 500, cancellation SMS → 200 `ignored` with nothing recorded. `CardRecorder` is faked, so no network. |
| `internal/parser` | 25 | Both issuers over real SMS fixtures — Shinhan old (`일시불` / `할부`) and new formats, Woori line layout, merchant names containing 잔/액, cancellation SMS (without misfiring on a merchant named 취소…), plus malformed input: missing merchant, missing amount, too few lines, unsupported issuer, empty message. |
| `internal/notion` | 11 | `MM/DD` → `YYYY-MM-DD` in KST: future dates roll back a year, the KST new-year early-morning case, impossible dates (`13/45`, `02/29` in a non-leap year) and malformed input fall back to today. |

Coverage: `internal/handler` 100%, `internal/parser` 92%.

Two invariants the tests pin down, because getting them wrong writes bad data to Notion:

- **No record on failure** — if binding or parsing fails, `CardRecorder` is never called. Notion writes can't be undone.
- **No partial results** — a parse error returns zero values, never a half-filled record.

New SMS fixtures go in `internal/parser/message_parser_test.go`. When an issuer changes their format, that's the file that goes red first.
