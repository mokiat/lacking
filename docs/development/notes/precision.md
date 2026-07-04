# Floating-Point Precision

## 32-bit vs 64-bit Performance

Benchmark tests show a measurable performance difference between 32-bit (`float32`) and 64-bit (`float64`) implementations, where 32-bit is generally faster.

### Package gjk2d comparisons

| Platform     | 64-bit vs 32-bit (sec/op) |
|------------- | ------------------------: |
| Linux        | +0.86%                    |
| MacOS        | +10.79%                   |
| WASM (Linux) | -11.51%                   |
| WASM (MacOS) | +0.37%                    |

### Package query2d comparisons

| Platform     | 64-bit vs 32-bit (sec/op) |
|------------- | ------------------------: |
| Linux        | +5.64%                    |
| MacOS        | +3.27%                    |
| WASM (Linux) | +0.02%                    |
| WASM (MacOS) | +2.75%                    |

## Decision

Despite the performance implications, the project proceeds with **64-bit (`float64`) implementations** as the default for the majority of packages, with the following rationale:

- **Larger scene sizes** - 64-bit coordinates avoid precision loss at large world-space distances, which would otherwise cause jitter and incorrect collision results far from the origin.
- **Better numerical stability** - algorithms such as impulse-based physics constraints accumulate floating-point errors across iterations and 64-bit floats help mitigate that.
- **Ease of use** - it's easier to work with float64 in Go and since this engine does not aim for AAA games, it should be fine.

The main **trade-off** is the increased memory usage (up to 2x, though often less).
