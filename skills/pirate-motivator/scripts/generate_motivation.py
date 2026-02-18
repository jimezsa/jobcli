#!/usr/bin/env python3
"""Generate ultra-motivational job hunting audio with Minimax TTS."""

import argparse
import json
import os
import sys
from pathlib import Path
import urllib.request
import urllib.error


def load_dotenv(path: Path = Path(".env")) -> None:
    """Load .env file into os.environ (no-op if file missing)."""
    if not path.is_file():
        return
    for line in path.read_text(encoding="utf-8").splitlines():
        line = line.strip()
        if not line or line.startswith("#"):
            continue
        key, _, value = line.partition("=")
        key = key.strip()
        value = value.strip().strip("\"'")
        if key:
            os.environ.setdefault(key, value)


_PROJECT_ROOT = Path(__file__).resolve().parent.parent.parent.parent
load_dotenv(_PROJECT_ROOT / ".env")

DEFAULT_API_KEY = os.environ.get("MINIMAX_API_KEY", "")
DEFAULT_VOICE = "angry_pirate_1"
DEFAULT_MODEL = "speech-2.8-turbo"

def generate_audio(text: str, output_path: str, api_key: str, voice: str = DEFAULT_VOICE, model: str = DEFAULT_MODEL) -> bool:
    """Call Minimax TTS API and save audio to file."""

    url = "https://api.minimax.io/v1/t2a_v2"

    payload = {
        "model": model,
        "text": text,
        "voice_setting": {
            "voice_id": voice,
            "speed": 1.0,
            "vol": 1.0,
            "pitch": 0
        },
        "audio_setting": {
            "sample_rate": 32000,
            "bitrate": 128000,
            "format": "mp3"
        }
    }

    headers = {
        "Authorization": f"Bearer {api_key}",
        "Content-Type": "application/json"
    }

    try:
        req = urllib.request.Request(
            url,
            data=json.dumps(payload).encode("utf-8"),
            headers=headers,
            method="POST"
        )

        with urllib.request.urlopen(req, timeout=60) as response:
            data = json.loads(response.read().decode("utf-8"))

        if "data" in data and "audio" in data["data"]:
            audio_bytes = bytes.fromhex(data["data"]["audio"])
            with open(output_path, "wb") as f:
                f.write(audio_bytes)
            return True
        else:
            print(f"API Error: {data}", file=sys.stderr)
            return False

    except urllib.error.URLError as e:
        print(f"Network error: {e}", file=sys.stderr)
        return False
    except Exception as e:
        print(f"Error: {e}", file=sys.stderr)
        return False

def main():
    parser = argparse.ArgumentParser(description="Generate motivational audio with Minimax TTS")
    parser.add_argument("text", help="Text to convert to speech")
    parser.add_argument("-o", "--output", default="/tmp/pirate_motivation.mp3", help="Output file path")
    parser.add_argument("-k", "--api-key", default=DEFAULT_API_KEY, help="Minimax API key")
    parser.add_argument("-v", "--voice", default=DEFAULT_VOICE, help="Voice ID (default: angry_pirate_1)")
    parser.add_argument("-m", "--model", default=DEFAULT_MODEL, help="Model (default: speech-2.8-turbo)")

    args = parser.parse_args()

    if not args.api_key:
        print("Error: MINIMAX_API_KEY not set and --api-key not provided", file=sys.stderr)
        sys.exit(1)

    if generate_audio(args.text, args.output, args.api_key, args.voice, args.model):
        print(f"Audio saved to: {args.output}")
        print(f"MEDIA: {args.output}")
    else:
        sys.exit(1)

if __name__ == "__main__":
    main()
                
