# GoSched Algorithms

## Rate Monotonic (RM)

Static priorities are assigned by period: shorter period ⇒ higher priority. RM is optimal among **fixed-priority** preemptive schedulers for periodic tasks.

**Liu & Layland utilization bound** (sufficient test for n periodic tasks on one CPU):

\[
U = \sum_{i=1}^{n} \frac{C_i}{T_i} \le n\left(2^{1/n} - 1\right)
\]

As \(n \to \infty\), the bound approaches \(\ln 2 \approx 0.693\) (**69% CPU**). Below this utilization, RM can guarantee all deadlines for a specific task set model.

## Earliest Deadline First (EDF)

Dynamic priority: the ready task with the **earliest absolute deadline** runs next.

**Optimality (uniprocessor):** If any scheduling algorithm can meet all deadlines for a task set, EDF meets all deadlines. EDF maximizes utilization among uniprocessor algorithms (up to 100% in ideal models).

### Equal deadlines?

GoSched breaks ties by arrival time, then static priority, then task ID — deterministic and stable.

## Least Laxity First (LLF)

Laxity (slack):

\[
L(t) = D - t - C_{rem}
\]

The task with **minimum laxity** is most urgent and runs next. LLF reacts as time progresses; GoSched rebuilds the heap on each tick.

## Complexity

| Operation | Complexity |
|-----------|------------|
| Heap push/pop | O(log n) |
| Deadline scan (naive) | O(n) |
| LLF reorder tick | O(n) heapify |

## FAQ

**EDF with identical deadlines?** Tie-breakers keep ordering deterministic; behavior is documented in tests.

**Many short tasks?** EDF or LLF typically outperform static RM under bursty loads because priorities adapt.

**Embedded use?** The library is suitable as a reference/simulation core. Hard real-time firmware would also need interrupt latency guarantees, WCET analysis, and hardware-specific validation — not provided here.
