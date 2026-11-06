# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

Go web server (single `package main`, no subpackages) that turns Twitch chat/bot requests into ElevenLabs TTS audio and streams it over WebSocket to an OBS browser source. Static frontend lives in `static/` (plain HTML/CSS/JS, no build step). There are no tests.

## Commands

```bash
go build -o main .          # build
go run . 8080 debug         # run: <port> [log level: info | debug | fountain]
go vet ./...
docker compose up           # app + MongoDB
```

- Port arg is required; `setupENV()` calls `log.Fatal` without it.
- `.env` is loaded via godotenv. Required: `ELEVENLABS_KEY`, `SERVER_URL`, `TTS_KEY`, `FFMPEG_ENABLED=true`. If any is missing, `setupENV()` logs an error and **returns early**, skipping arg parsing and all voice/DB setup. See `.env.example` for the per-voice JSON env vars (`VOICES`, `VOICE_MODELS`, `VOICE_STYLES`, `VOICE_SPEEDS`, `VOICE_STABILITIES`, `VOICE_CLONE_STRENGTHS`, etc.).
- `ffmpeg` must be on PATH (modifiers and audio-length measurement shell out to it).
- The server reads `static/`, `effects/`, `alerts/<channel>/` relative to the working directory, so run from the repo root.
- CI (`.github/workflows/main.yml`): pushes to `main`/`dev` publish Docker image `johnnycyan/ai-twitch-tts`; pushes to `main` also cut a timestamped GitHub release with Windows/Linux zips (binary + `static/`).

## Architecture

### Request flow

1. `GET /tts?channel=&key=&voice=&text=` → `handleRequest` (`request.go`). Rejects with 429 if the channel already has an entry in the global `requests` slice, 404 if no WebSocket client is connected for that channel.
2. Pally.gg tips (`pally.go`) connect per channel from `PALLY_KEYS` and enter the same pipeline with `kind == "pally"` and `PlayAlert: true`.
3. Both build a `Message` and call `ProcessAndPlay` (`processor.go`), the unified pipeline:
   - `ParseMessage` → `parseTextToSegments` splits text into `AudioSegment`s on tags; `assignContext` fills `PreviousText`/`NextText` for consecutive same-voice segments.
   - **Phase 1**: pre-generate every segment (effect file bytes, or `generateAudio` → ElevenLabs stream, then ffmpeg modifiers; logs to Mongo if enabled).
   - **Phase 2**: for each segment send `start <time> [durationMs]` text, then binary audio, then busy-wait until the client replies `confirm <time>` (clears `playing[time]`). 120 s timeout sends `reload` to the client.

### Tag syntax (parsed in `processor.go` / `audioeffects.go`)

- `(name)` — voice, effect (file in `effects/`), or modifier (`reverb`, `phone`; see `modifierFuncs` in `modifiers.go`). Legacy `[v-name]` / `[e-name]` still supported.
- `[tag]` — ElevenLabs v3/v4 inline expression tags, passed through to the API; stripped in `generateAudio` for turbo/v2 models. This is why voice/effect tags use `()`, not `[]`.
- `<tag>` — ElevenLabs v4 `audio_effects` (filter/environment/noise/distance, plus `-off` resets) via the `audioEffectTags` map. Sticky until reset; ignored for non-`eleven_v4` models. Unknown `<...>` is spoken as text.

### Per-voice settings

`tts.go` holds the voice list and parallel slices of per-voice settings, each loaded from its own JSON env var by a `setupVoice*()` function and read by a `getVoice*(voiceID)` lookup (ID → name → case-insensitive match). To add a new per-voice setting, follow the existing pattern: struct + `setupVoiceX` in `tts.go`, call it from `setupENV()`, look it up in `ProcessAndPlay`, send it in `ttsStream`, document it in `.env.example`. Model aliases (`turbo`, `v2`, `v3`, `v4`) map to ElevenLabs model IDs in `generateAudio`; default is `eleven_v4`.

### WebSocket / client

- `static/index.html` is the OBS overlay client (served as a Go template by `serveClient` on `/?channel=...`). It connects to `/ws?channel=&v=<md5 of index.html>`. A hash mismatch makes the server send `update <hash>` and close, forcing the client to reload. Editing `index.html` therefore forces connected overlays to refresh. `/update?channel=` triggers this manually.
- Client sends `ping` (keepalive, 120 s/60 s timeout) and `confirm <time>`.
- Global state (`clients`, `requests`, `playing`) is package-level maps/slices, only partly mutex-guarded.

### Web UI

`/create`, `/voices`, `/effects`, `/chart` all serve `static/app.html` (an SPA driven by `app.js`; message builder in `create.js`). Data comes from `/api/voices`, `/api/effects`, `/data/{channel}` (Mongo usage stats, registered only when Mongo env vars are set), and `/eleven/characters`.

### ffmpeg modifiers

`modifiers.go` writes `input-<channel>.mp3`, runs ffmpeg, and reads `output-<channel>.mp3` in the working directory. Reverb uses `static/reverb.wav` as the impulse response; the pre-reverb duration is sent with `start` so the client can overlap the reverb tail.
