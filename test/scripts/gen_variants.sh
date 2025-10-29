#!/usr/bin/env bash
set -euo pipefail

BASE=../fixtures/base
OUT=../fixtures/variants

FILE=b.wav

mkdir -p "$OUT"

# volume +/-6dB
ffmpeg -y -i "$BASE/$FILE" -filter:a "volume=2.0" "$OUT/plus6dB_$FILE"
ffmpeg -y -i "$BASE/$FILE" -filter:a "volume=0.5" "$OUT/minus6dB_$FILE"

# truncations by percent (keep first X%)
duration=$(ffprobe -v error -show_entries format=duration -of default=nw=1:nk=1 "$BASE/$FILE")
# compute durations and create truncated files
for p in 95 90 75 50; do
  keep=$(awk -v d="$duration" -v p="$p" 'BEGIN{printf "%.3f", d*(p/100)}')
  ffmpeg -y -i "$BASE/$FILE" -t "$keep" "$OUT/trunc_${p}p_$FILE"
done

# recompress to mp3 128k and back to wav
ffmpeg -y -i "$BASE/$FILE" -c:a libmp3lame -b:a 128k "$OUT/a_128kbps.mp3"
ffmpeg -y -i "$OUT/$FILE.mp3" "$OUT/a_128kbps$FILE"

# replace contiguous portion (example 10% center) - produce a_replace_10p.wav
# This uses ffmpeg segment+concat; compute times as needed. For deterministic scripts compute t1,t2.
# Implement similarly for 5,10,25,50,75 by computing replacement durations and splicing.

# Note: keep deterministic paths & names; commit into repo or run script in CI before tests.
