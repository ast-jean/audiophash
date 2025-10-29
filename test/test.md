### Test Types and Expected Hash Distance Behavior

All tests compute two hashes **h1** and **h2**, measure Hamming distance **d bits**, and evaluate:

```
percent = (d / 64) * 100
```

#### Same Audio File

* Expected: **≤ 1.6%** (≤ 1 bit difference)
* Pass if: `percent <= 1.6`

#### Completely Different Audio Files

* Expected: **≥ 40%**
* Pass if: `percent >= 40`

#### Volume Change (e.g., +6 dB, −6 dB)

* Expected: **≤ 9.4%** (≤ 6 bits difference)
* Pass if: `percent <= 9.4`

#### Audio Truncation (tail removed)

| Truncation Amount | Expected Max Percent | Pass Rule     |
| ----------------- | -------------------- | ------------- |
| 5%                | ≤ 6%                 | percent <= 6  |
| 10%               | ≤ 8%                 | percent <= 8  |
| 25%               | ≤ 14%                | percent <= 14 |
| 50%               | ≤ 28%                | percent <= 28 |

#### Compression / Re-encode (MP3 / Opus, moderate bitrate)

* Expected: **≤ 12.5%** (≤ 8 bits difference)
* Pass if: `percent <= 12.5`

#### Portion of Audio Replaced

| Replacement Portion | Expected Percent | Pass Rule     |
| ------------------- | ---------------- | ------------- |
| 5%                  | ≤ 8%             | percent <= 8  |
| 10%                 | ≤ 12%            | percent <= 12 |
| 25%                 | ≤ 20%            | percent <= 20 |
| 50%                 | ≤ 40%            | percent <= 40 |
| 75%                 | ≥ 45%            | percent >= 45 |

---
