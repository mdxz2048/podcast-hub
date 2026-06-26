#!/usr/bin/env python3
import json
import pathlib
import sys


def emit(event):
    print(json.dumps(event, separators=(",", ":")), flush=True)


def main():
    if len(sys.argv) != 3:
        emit({"type": "failed", "level": "error", "message": "fixture requires input and output paths"})
        return 1
    input_path = pathlib.Path(sys.argv[1])
    output_dir = pathlib.Path(sys.argv[2])
    job = json.loads(input_path.read_text())
    rel = pathlib.Path("episodes") / "fixture-episode.json"
    target = output_dir / rel
    target.parent.mkdir(parents=True, exist_ok=True)
    target.write_text(json.dumps({"schema_version": "1.0", "job_id": job["job"]["id"], "title": "Fixture"}))
    emit({"type": "log", "level": "info", "message": "fixture connector started"})
    emit({"type": "artifact_ready", "level": "info", "artifact_type": "episode_metadata", "path": str(rel)})
    emit({"type": "completed", "level": "info", "message": "fixture connector completed"})
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
