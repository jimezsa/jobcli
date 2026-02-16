---
name: pirate-motivator
description: Generate funny, short ultra-motivational audio messages for job hunting using an angry pirate voice. Use when the user wants motivation for job searching, needs encouragement during unemployment, asks for a pep talk about applications/interviews, or requests pirate-style motivation audio.
---

# Pirate Motivator

Generate short motivational audio clips with an angry pirate voice for job-hunt encouragement.

## Requirements

- Minimax API key (set as `MINIMAX_API_KEY` env var or pass via `--api-key`)

## Usage

```bash
python3 scripts/generate_motivation.py "YOUR MOTIVATIONAL TEXT" -o /tmp/motivation.mp3 -k YOUR_API_KEY
```

The script outputs `MEDIA: /path/to/file.mp3` which OpenClaw will send to the user.

## Workflow

1. **Generate the message**: Write a short (1-2 sentences, max 35 words), intense, funny motivational message about job hunting. Channel angry pirate energy and use phrases like "Arrr!", "ye scallywag", "landlubber", etc.

2. **Run the script**:

   ```bash
   python3 scripts/generate_motivation.py "MESSAGE" -o /tmp/pirate_motivation.mp3 -k API_KEY
   ```

3. **Send the audio**: Use the message tool to send the generated MP3 to the user.

## Ranking Integration Rules

- Use this skill only after job ranking output is complete.
- Never feed motivational text/audio back into ranking logic.
- Keep motivational output concise to avoid token waste.

## Example Messages

- "Arrr! Get off yer lazy stern and send them applications, ye scallywag! Every rejection be just another wave before ye reach treasure island!"
- "Listen here, landlubber! Yer skills be worth more than all the gold in Davy Jones' locker! Now get out there and CRUSH IT!"
- "Shiver me timbers! Stop doubting yerself! Ye be a BEAST of the seven seas! Now hoist the sails and land that job!"

## Voice Settings

- **Model**: `speech-2.8-turbo` (fast, good quality)
- **Voice**: `angry_pirate_1` (intense, pirate character)
- **Format**: MP3, 32kHz, 128kbps
