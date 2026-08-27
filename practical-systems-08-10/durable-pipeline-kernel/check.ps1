$ErrorActionPreference = 'Stop'

$projectPath = (Resolve-Path $PSScriptRoot).Path
$image = 'ghcr.io/gleam-lang/gleam:v1.18.1-erlang-alpine@sha256:7c82e4a284b7c05c26eac34db497ea0e63ce7cb04bd019d966d70338eb172b68'

docker run --rm `
  --mount "type=bind,source=$projectPath,target=/app" `
  --workdir /app `
  $image `
  sh -lc 'gleam deps download && gleam format --check src test && gleam check && gleam test && gleam run && gleam run -m replay_evidence'

if ($LASTEXITCODE -ne 0) {
  exit $LASTEXITCODE
}
