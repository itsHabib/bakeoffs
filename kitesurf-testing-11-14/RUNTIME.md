# Local Kitesurf setup

The demos were tested with upstream [cloudflare/kitesurf](https://github.com/cloudflare/kitesurf) revision `4a2903813a22b08d4f3cc64a4b35339d0293a5d1`. Kitesurf is experimental; use the pinned revision to reproduce this snapshot.

Install Node 22+, npm, Git, Bazelisk and GNU tar. On macOS, the latter two are available with `brew install bazelisk gnu-tar`. The upstream build downloads its toolchain and dependencies; allow time and disk space for the initial Rust/Wasm build.

```sh
git clone https://github.com/cloudflare/kitesurf.git
cd kitesurf
git checkout 4a2903813a22b08d4f3cc64a4b35339d0293a5d1
npx --yes pnpm@9.15.4 install --frozen-lockfile
bazel run --jobs=4 //packages/engine:dev_build
export KITESURF_SOURCE="$PWD"
```

Then change to this archive's `kitesurf-testing-11-14` directory and run `./demo.sh` with that environment variable still set. The supported setup was tested on macOS; the upstream README is the source for other platforms.

To start one engine manually:

```sh
KITESURF_SOURCE=/absolute/path/to/kitesurf ./start-runtime.sh 18891
```

Use 18891 for Fixture Pruner, 18892 for Needle, or 18893 for Proofgap. Each has a separate inspector port and persistence directory. Startup refuses an occupied port. `lsof` is required for that check. The combined demo can reuse already-running Kitesurf endpoints.

Live `/json/version` identifies `Kitesurf/0.0.1`; `Browser.getVersion` uses Chrome compatibility branding and an `@kitesurf` marker. That branding is not evidence of Chromium execution. Source revision metadata describes the supplied checkout, not a cryptographic attestation of the running engine.
