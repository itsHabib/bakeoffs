import digest
import gleam/dynamic/decode
import gleam/json
import gleam/list
import gleam/string
import journal
import pipeline.{type Verdict}

pub const schema = "EvidenceReplayManifestV1"

pub type ReplayManifest {
  ReplayManifest(
    subject: String,
    claim: String,
    pipeline_digest: String,
    tool: String,
    tool_version: String,
    command: String,
    input_path: String,
    input_digest: String,
    seed: String,
    bounds: List(String),
    expected: Verdict,
  )
}

pub type Reproduction {
  Reproduction(
    subject: String,
    claim: String,
    expected: Verdict,
    actual: Verdict,
  )
}

pub type ReplayError {
  ReadFailure(String)
  MalformedManifest(String)
  VerdictMismatch(expected: Verdict, actual: Verdict)
}

pub fn read(path: String) -> Result(ReplayManifest, ReplayError) {
  case journal.read_raw(path) {
    Error(error) -> Error(ReadFailure(error))
    Ok(bytes) -> decode_manifest(bytes)
  }
}

pub fn decode_manifest(bytes: String) -> Result(ReplayManifest, ReplayError) {
  case json.parse(bytes, manifest_decoder()) {
    Error(error) -> Error(MalformedManifest(string.inspect(error)))
    Ok(manifest) -> validate(manifest)
  }
}

pub fn encode(manifest: ReplayManifest) -> String {
  json.object([
    #("schema", json.string(schema)),
    #("subject", json.string(manifest.subject)),
    #("claim", json.string(manifest.claim)),
    #("pipeline_digest", json.string(manifest.pipeline_digest)),
    #("tool", json.string(manifest.tool)),
    #("tool_version", json.string(manifest.tool_version)),
    #("command", json.string(manifest.command)),
    #("input_path", json.string(manifest.input_path)),
    #("input_digest", json.string(manifest.input_digest)),
    #("seed", json.string(manifest.seed)),
    #("bounds", json.array(manifest.bounds, json.string)),
    #("expected_verdict", json.string(journal.verdict_name(manifest.expected))),
  ])
  |> json.to_string
}

pub fn reproduce(
  manifest: ReplayManifest,
  validator: fn(ReplayManifest) -> Verdict,
) -> Result(Reproduction, ReplayError) {
  let actual = validator(manifest)
  case actual == manifest.expected {
    True ->
      Ok(Reproduction(
        manifest.subject,
        manifest.claim,
        manifest.expected,
        actual,
      ))
    False -> Error(VerdictMismatch(manifest.expected, actual))
  }
}

pub fn result_json(reproduction: Reproduction) -> String {
  json.object([
    #("schema", json.string("EvidenceReproductionV1")),
    #("subject", json.string(reproduction.subject)),
    #("claim", json.string(reproduction.claim)),
    #("expected", json.string(journal.verdict_name(reproduction.expected))),
    #("actual", json.string(journal.verdict_name(reproduction.actual))),
    #("reproduced", json.bool(reproduction.expected == reproduction.actual)),
  ])
  |> json.to_string
}

pub fn fixture_validator(manifest: ReplayManifest) -> Verdict {
  case
    manifest.subject == "rev-b",
    manifest.claim == "e2e",
    manifest.pipeline_digest == "sha256:shipping-pipeline-v1",
    manifest.tool == "fixture-validator-e2e",
    manifest.tool_version == "1.0.0",
    manifest.seed == "424242",
    manifest.bounds == ["cases=4", "timeout_ms=1000"]
  {
    True, True, True, True, True, True, True -> reproduce_fixture(manifest)
    _, _, _, _, _, _, _ -> pipeline.Insufficient
  }
}

type FixtureCheck {
  FixtureCheck(name: String, passed: Bool)
}

type E2EFixture {
  E2EFixture(subject: String, checks: List(FixtureCheck))
}

fn reproduce_fixture(manifest: ReplayManifest) -> Verdict {
  case journal.read_raw(manifest.input_path) {
    Error(_) -> pipeline.Insufficient
    Ok(bytes) ->
      case json.parse(bytes, fixture_decoder()) {
        Error(_) -> pipeline.Insufficient
        Ok(fixture) -> {
          let canonical = encode_fixture(fixture)
          case
            fixture.subject == manifest.subject,
            digest.sha256(canonical) == manifest.input_digest,
            list.all(fixture.checks, fn(check) { check.passed })
          {
            True, True, True -> pipeline.Supported
            True, True, False -> pipeline.Refuted
            _, _, _ -> pipeline.Insufficient
          }
        }
      }
  }
}

fn encode_fixture(fixture: E2EFixture) -> String {
  json.object([
    #("schema", json.string("E2EFixtureV1")),
    #("subject", json.string(fixture.subject)),
    #(
      "checks",
      json.array(fixture.checks, fn(check) {
        json.object([
          #("name", json.string(check.name)),
          #("passed", json.bool(check.passed)),
        ])
      }),
    ),
  ])
  |> json.to_string
}

fn fixture_decoder() -> decode.Decoder(E2EFixture) {
  use found_schema <- decode.field("schema", decode.string)
  use subject <- decode.field("subject", decode.string)
  use checks <- decode.field("checks", decode.list(fixture_check_decoder()))
  case found_schema == "E2EFixtureV1", subject != "", checks != [] {
    True, True, True -> decode.success(E2EFixture(subject, checks))
    _, _, _ ->
      decode.failure(E2EFixture("", []), expected: "valid E2EFixtureV1")
  }
}

fn fixture_check_decoder() -> decode.Decoder(FixtureCheck) {
  use name <- decode.field("name", decode.string)
  use passed <- decode.field("passed", decode.bool)
  case name != "" {
    True -> decode.success(FixtureCheck(name, passed))
    False ->
      decode.failure(FixtureCheck("", False), expected: "named fixture check")
  }
}

fn validate(manifest: ReplayManifest) -> Result(ReplayManifest, ReplayError) {
  case
    manifest.subject != "",
    manifest.claim != "",
    manifest.pipeline_digest != "",
    manifest.tool != "",
    manifest.tool_version != "",
    manifest.command != "",
    manifest.input_path != "",
    manifest.input_digest != "",
    manifest.seed != "",
    manifest.bounds != []
  {
    True, True, True, True, True, True, True, True, True, True -> Ok(manifest)
    _, _, _, _, _, _, _, _, _, _ ->
      Error(MalformedManifest("replay manifest has empty required fields"))
  }
}

fn manifest_decoder() -> decode.Decoder(ReplayManifest) {
  use found_schema <- decode.field("schema", decode.string)
  use subject <- decode.field("subject", decode.string)
  use claim <- decode.field("claim", decode.string)
  use pipeline_digest <- decode.field("pipeline_digest", decode.string)
  use tool <- decode.field("tool", decode.string)
  use tool_version <- decode.field("tool_version", decode.string)
  use command <- decode.field("command", decode.string)
  use input_path <- decode.field("input_path", decode.string)
  use input_digest <- decode.field("input_digest", decode.string)
  use seed <- decode.field("seed", decode.string)
  use bounds <- decode.field("bounds", decode.list(decode.string))
  use expected <- decode.field("expected_verdict", verdict_decoder())
  case found_schema == schema {
    True ->
      decode.success(ReplayManifest(
        subject,
        claim,
        pipeline_digest,
        tool,
        tool_version,
        command,
        input_path,
        input_digest,
        seed,
        bounds,
        expected,
      ))
    False ->
      decode.failure(
        ReplayManifest(
          "",
          "",
          "",
          "",
          "",
          "",
          "",
          "",
          "",
          [],
          pipeline.Insufficient,
        ),
        expected: schema,
      )
  }
}

fn verdict_decoder() -> decode.Decoder(Verdict) {
  use value <- decode.then(decode.string)
  case value {
    "supported" -> decode.success(pipeline.Supported)
    "refuted" -> decode.success(pipeline.Refuted)
    "insufficient" -> decode.success(pipeline.Insufficient)
    _ -> decode.failure(pipeline.Insufficient, expected: "known verdict")
  }
}
