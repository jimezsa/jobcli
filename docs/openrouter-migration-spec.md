# Spec: Migrate `job_discriminator.py` from MiniMax to OpenRouter

## Goal
Replace the MiniMax provider in `job_discriminator.py` with OpenRouter as the LLM
backend for job relevance filtering.

## Background
The script currently speaks the **Anthropic Messages API** dialect:
- Endpoint `.../v1/messages`
- Auth via `x-api-key` + `anthropic-version` headers
- Top-level `system` field and content-block message format

OpenRouter speaks the **OpenAI Chat Completions** dialect. This migration is
therefore an envelope change (request + response shape), not just a URL/key swap.

## Decision
**Clean cutover to OpenRouter-only.** Do not keep a dual-dialect fallback — the
Anthropic request format is incompatible with OpenAI's, and supporting both in one
function is error-prone. Env-var fallbacks to other keys may remain, but a single
request/response path is used.

## Changes

### 1. Configuration constants
| Item | From | To |
|------|------|-----|
| API key env | `MINIMAX_API_KEY` | `OPENROUTER_API_KEY` |
| Endpoint | `ANTHROPIC_BASE_URL` + `/v1/messages` | `OPENROUTER_BASE_URL` (default `https://openrouter.ai/api/v1/chat/completions`) |
| Default model | `MiniMax-M2.5` | OpenRouter slug, e.g. `minimax/minimax-m2` |

- `DEFAULT_API_KEY`: read `OPENROUTER_API_KEY` first (keep `ANTHROPIC/OPENAI`
  fallbacks only if desired).
- Update `--api-key` help text and the missing-key error message to name
  `OPENROUTER_API_KEY`.

### 2. Request (in `llm_compare`)
- **Payload** (OpenAI format):
  - Remove top-level `system`.
  - `messages`: `[{"role": "system", "content": system_prompt},
    {"role": "user", "content": user_prompt}]` — content is a plain string
    (no `[{"type":"text",...}]` blocks).
  - Keep `model`, `max_tokens`, `temperature`.
  - Optional: add `"response_format": {"type": "json_object"}` to force JSON.
- **Headers**:
  - Replace `x-api-key` + `anthropic-version` with
    `Authorization: Bearer <api_key>`.
  - Keep `Content-Type: application/json`.
  - Optional: `HTTP-Referer` and `X-Title` (OpenRouter app-ranking; cosmetic).

### 3. Response parsing
- OpenRouter returns the OpenAI shape: `choices[0].message.content` as a string.
- Remove the Anthropic content-block branch and the MiniMax `base_resp` error
  check — neither applies.
- Optional: handle OpenRouter error shape `{"error": {...}}` and raise a clear
  `RuntimeError`.

### 4. Cosmetic
- Update module docstring and the "MiniMax/OpenAI-compatible" comment to reference
  OpenRouter.

## Non-goals
- No change to persona/job parsing, threading, CLI args (other than help text), or
  the decision/confidence output contract (`{"decision":"YES|NO","confidence":"HIGH|LOW"}`).

## Acceptance criteria
1. With `OPENROUTER_API_KEY` set, running against a sample jobs JSON produces YES/NO
   decisions and writes the output file.
2. No `x-api-key` / `anthropic-version` headers or `/v1/messages` calls remain.
3. Missing `OPENROUTER_API_KEY` (and no `--api-key`) exits with a clear error naming
   `OPENROUTER_API_KEY`.
4. `parse_decision` still yields valid `decision`/`confidence` on both clean JSON and
   JSON embedded in prose.

## Test plan
- Smoke test: `--max-jobs 3` against a known jobs file; verify decisions log and
  output list.
- Bad key: confirm graceful per-job error handling (network/API errors already caught
  in `evaluate`).
- Malformed model response: confirm `parse_decision` defaults to
  `{"decision":"NO","confidence":"LOW"}`.

## Rollback
Revert the single file; the env var `OPENROUTER_API_KEY` is additive and can stay.
