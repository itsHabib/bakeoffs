import evidence
import gleam/io

pub fn main() -> Nil {
  let assert Ok(manifest) = evidence.read("fixtures/evidence/e2e-rev-b.json")
  let assert Ok(reproduction) =
    evidence.reproduce(manifest, evidence.fixture_validator)
  io.println(evidence.result_json(reproduction))
}
