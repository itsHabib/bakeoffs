// The channel scheduler: priority is code, not vibes.
//
// Rules, in order:
//   1. Priority traffic (escalations) queues ahead of all routine traffic,
//      FIFO among itself, and is NEVER coalesced or dropped.
//   2. A routine call supersedes an older queued routine call with the same
//      key (same agent + stream) — the ear only needs the latest state.
//   3. When estimated backlog exceeds maxBacklogSec, the oldest routine call
//      is dropped, like a saturated frequency. The newest routine call is
//      protected so fresh state always has a chance to air.
//
// Pure data-structure policy: no timers, no audio, fully table-testable.

export function createChannel({ maxBacklogSec = 9, secPerWord = 0.42, overheadSec = 0.7 } = {}) {
  const queue = [];
  const stats = { transmitted: 0, coalesced: 0, dropped: 0 };
  const itemSec = (it) => it.words * secPerWord + overheadSec;
  const backlogSec = () => queue.reduce((s, it) => s + itemSec(it), 0);

  function enqueue(item) {
    if (item.priority) {
      let i = 0;
      while (i < queue.length && queue[i].priority) i++;
      queue.splice(i, 0, item);
      return;
    }
    const dup = queue.findIndex((it) => !it.priority && it.key === item.key);
    if (dup >= 0) {
      queue.splice(dup, 1);
      stats.coalesced++;
    }
    queue.push(item);
    while (backlogSec() > maxBacklogSec) {
      const oldest = queue.findIndex((it) => !it.priority && it !== item);
      if (oldest < 0) break;
      queue.splice(oldest, 1);
      stats.dropped++;
    }
  }

  function next() {
    const it = queue.shift() ?? null;
    if (it) stats.transmitted++;
    return it;
  }

  return {
    enqueue,
    next,
    stats,
    backlogSec,
    get pending() {
      return queue.length;
    },
  };
}
