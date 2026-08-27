import evidence
import examples
import pipeline

pub fn committed_manifest_reproduces_in_a_fresh_reader_test() {
  let assert Ok(manifest) = evidence.read("fixtures/evidence/e2e-rev-b.json")
  let assert Ok(reproduction) =
    evidence.reproduce(manifest, evidence.fixture_validator)
  let evidence.Reproduction(subject:, claim:, actual:, ..) = reproduction
  assert subject == "rev-b"
  assert claim == "e2e"
  assert actual == pipeline.Supported
}

pub fn changed_input_cannot_inherit_the_recorded_verdict_test() {
  let assert Ok(manifest) = evidence.read("fixtures/evidence/e2e-rev-b.json")
  let evidence.ReplayManifest(
    subject,
    claim,
    pipeline_digest,
    tool,
    tool_version,
    command,
    input_path,
    _,
    seed,
    bounds,
    expected,
  ) = manifest
  let changed =
    evidence.ReplayManifest(
      subject,
      claim,
      pipeline_digest,
      tool,
      tool_version,
      command,
      input_path,
      "fixture:e2e:other:v1",
      seed,
      bounds,
      expected,
    )
  let assert Error(evidence.VerdictMismatch(
    pipeline.Supported,
    pipeline.Insufficient,
  )) = evidence.reproduce(changed, evidence.fixture_validator)
}

pub fn manifest_round_trip_preserves_recipe_test() {
  let manifest =
    evidence.ReplayManifest(
      "rev-b",
      "e2e",
      examples.shipping_digest,
      "fixture-validator-e2e",
      "1.0.0",
      "gleam run -m replay_evidence",
      "fixtures/evidence/e2e-rev-b-input.json",
      "44ff1c5f670a76aee904737ccd707b6873118ae71bba7b4bc15661b5057fcee3",
      "424242",
      ["cases=4", "timeout_ms=1000"],
      pipeline.Supported,
    )
  let assert Ok(decoded) = evidence.decode_manifest(evidence.encode(manifest))
  assert decoded == manifest
}
