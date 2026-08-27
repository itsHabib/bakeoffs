import digest
import pipeline.{type Event, type Evidence, type Pipeline, type Projection}

pub const shipping_digest = "sha256:shipping-pipeline-v1"

pub const maintenance_digest = "sha256:maintenance-pipeline-v1"

pub fn shipping() -> Pipeline {
  let ship = pipeline.step("ship")
  let validate = pipeline.step("validate")
  let e2e = pipeline.step("e2e")
  let assure = pipeline.step("assure")
  let land = pipeline.step("land")
  pipeline.Pipeline(
    name: "shipping",
    version: 1,
    digest: shipping_digest,
    start: ship,
    steps: [
      pipeline.StepSpec(ship, pipeline.Deduplicated, []),
      pipeline.StepSpec(validate, pipeline.Idempotent, []),
      pipeline.StepSpec(e2e, pipeline.Idempotent, ["unit"]),
      pipeline.StepSpec(assure, pipeline.Idempotent, ["unit", "e2e"]),
      pipeline.StepSpec(land, pipeline.Deduplicated, [
        "unit",
        "e2e",
        "assurance",
      ]),
    ],
    rules: [
      pipeline.Rule(ship, pipeline.Success, pipeline.GoTo(validate, False)),
      pipeline.Rule(ship, pipeline.Failure, pipeline.GoTo(ship, True)),
      pipeline.Rule(validate, pipeline.Success, pipeline.GoTo(e2e, False)),
      pipeline.Rule(validate, pipeline.Failure, pipeline.GoTo(ship, True)),
      pipeline.Rule(e2e, pipeline.Success, pipeline.GoTo(assure, False)),
      pipeline.Rule(e2e, pipeline.Failure, pipeline.GoTo(ship, True)),
      pipeline.Rule(assure, pipeline.Success, pipeline.GoTo(land, False)),
      pipeline.Rule(assure, pipeline.Failure, pipeline.GoTo(ship, True)),
      pipeline.Rule(land, pipeline.Success, pipeline.Finish),
      pipeline.Rule(
        land,
        pipeline.Failure,
        pipeline.Stop("landing requires judgment"),
      ),
    ],
  )
}

pub fn maintenance() -> Pipeline {
  let detect = pipeline.step("detect")
  let diagnose = pipeline.step("diagnose")
  let repair = pipeline.step("repair")
  let validate = pipeline.step("validate")
  let observe = pipeline.step("observe")
  pipeline.Pipeline(
    name: "maintenance",
    version: 1,
    digest: maintenance_digest,
    start: detect,
    steps: [
      pipeline.StepSpec(detect, pipeline.Idempotent, []),
      pipeline.StepSpec(diagnose, pipeline.Idempotent, []),
      pipeline.StepSpec(repair, pipeline.Deduplicated, []),
      pipeline.StepSpec(validate, pipeline.Idempotent, []),
      pipeline.StepSpec(observe, pipeline.Idempotent, ["maintenance_validation"]),
    ],
    rules: [
      pipeline.Rule(detect, pipeline.Success, pipeline.GoTo(diagnose, False)),
      pipeline.Rule(detect, pipeline.Failure, pipeline.Stop("detection failed")),
      pipeline.Rule(diagnose, pipeline.Success, pipeline.GoTo(repair, False)),
      pipeline.Rule(
        diagnose,
        pipeline.Failure,
        pipeline.Stop("diagnosis needs judgment"),
      ),
      pipeline.Rule(repair, pipeline.Success, pipeline.GoTo(validate, False)),
      pipeline.Rule(repair, pipeline.Failure, pipeline.GoTo(repair, True)),
      pipeline.Rule(validate, pipeline.Success, pipeline.GoTo(observe, False)),
      pipeline.Rule(validate, pipeline.Failure, pipeline.GoTo(repair, True)),
      pipeline.Rule(observe, pipeline.Success, pipeline.Finish),
      pipeline.Rule(observe, pipeline.Failure, pipeline.GoTo(diagnose, True)),
    ],
  )
}

pub fn recipe(claim: String, subject: String) -> pipeline.Recipe {
  pipeline.Recipe(
    tool: "fixture-validator-" <> claim,
    version: "1.0.0",
    command: "gleam run -m replay_evidence",
    input_digest: digest.sha256("fixture:" <> claim <> ":" <> subject <> ":v1"),
    seed: "424242",
    bounds: ["cases=4", "timeout_ms=1000"],
  )
}

pub fn supported(claim: String, subject: String) -> Evidence {
  pipeline.Evidence(claim, subject, pipeline.Supported, recipe(claim, subject))
}

pub fn refuted(claim: String, subject: String) -> Evidence {
  pipeline.Evidence(claim, subject, pipeline.Refuted, recipe(claim, subject))
}

pub fn shipping_trace() -> List(Event) {
  let ship = pipeline.step("ship")
  let validate = pipeline.step("validate")
  let e2e = pipeline.step("e2e")
  let assure = pipeline.step("assure")
  let land = pipeline.step("land")
  [
    pipeline.RunStarted(
      "shipping-demo",
      "shipping",
      1,
      shipping_digest,
      "base-a",
    ),
    pipeline.EffectPrepared(
      ship,
      1,
      "ship-1",
      pipeline.Deduplicated,
      "input:base-a",
    ),
    pipeline.StepSucceeded(ship, 1, "ship-1", "base-a", "rev-a", []),
    pipeline.EffectPrepared(
      validate,
      1,
      "validate-1",
      pipeline.Idempotent,
      "input:rev-a",
    ),
    pipeline.StepSucceeded(validate, 1, "validate-1", "rev-a", "rev-a", [
      supported("unit", "rev-a"),
    ]),
    pipeline.EffectPrepared(e2e, 1, "e2e-1", pipeline.Idempotent, "input:rev-a"),
    pipeline.StepFailed(e2e, 1, "e2e-1", "rev-a", "browser assertion failed", [
      refuted("e2e", "rev-a"),
    ]),
    pipeline.EffectPrepared(
      ship,
      2,
      "ship-2",
      pipeline.Deduplicated,
      "input:rev-a",
    ),
    pipeline.StepSucceeded(ship, 2, "ship-2", "rev-a", "rev-b", []),
    pipeline.EffectPrepared(
      validate,
      2,
      "validate-2",
      pipeline.Idempotent,
      "input:rev-b",
    ),
    pipeline.StepSucceeded(validate, 2, "validate-2", "rev-b", "rev-b", [
      supported("unit", "rev-b"),
    ]),
    pipeline.EffectPrepared(e2e, 2, "e2e-2", pipeline.Idempotent, "input:rev-b"),
    pipeline.StepSucceeded(e2e, 2, "e2e-2", "rev-b", "rev-b", [
      supported("e2e", "rev-b"),
    ]),
    pipeline.EffectPrepared(
      assure,
      2,
      "assure-2",
      pipeline.Idempotent,
      "input:rev-b",
    ),
    pipeline.StepSucceeded(assure, 2, "assure-2", "rev-b", "rev-b", [
      supported("assurance", "rev-b"),
    ]),
    pipeline.EffectPrepared(
      land,
      2,
      "land-2",
      pipeline.Deduplicated,
      "input:rev-b",
    ),
    pipeline.StepSucceeded(land, 2, "land-2", "rev-b", "rev-b", []),
  ]
}

pub fn maintenance_trace() -> List(Event) {
  let detect = pipeline.step("detect")
  let diagnose = pipeline.step("diagnose")
  let repair = pipeline.step("repair")
  let validate = pipeline.step("validate")
  let observe = pipeline.step("observe")
  [
    pipeline.RunStarted(
      "maintenance-demo",
      "maintenance",
      1,
      maintenance_digest,
      "machine-a",
    ),
    pipeline.EffectPrepared(
      detect,
      1,
      "detect-1",
      pipeline.Idempotent,
      "input:machine-a",
    ),
    pipeline.StepSucceeded(detect, 1, "detect-1", "machine-a", "machine-a", []),
    pipeline.EffectPrepared(
      diagnose,
      1,
      "diagnose-1",
      pipeline.Idempotent,
      "input:machine-a",
    ),
    pipeline.StepSucceeded(
      diagnose,
      1,
      "diagnose-1",
      "machine-a",
      "machine-a",
      [],
    ),
    pipeline.EffectPrepared(
      repair,
      1,
      "repair-1",
      pipeline.Deduplicated,
      "input:machine-a",
    ),
    pipeline.StepSucceeded(repair, 1, "repair-1", "machine-a", "machine-b", []),
    pipeline.EffectPrepared(
      validate,
      1,
      "validate-1",
      pipeline.Idempotent,
      "input:machine-b",
    ),
    pipeline.StepSucceeded(validate, 1, "validate-1", "machine-b", "machine-b", [
      supported("maintenance_validation", "machine-b"),
    ]),
    pipeline.EffectPrepared(
      observe,
      1,
      "observe-1",
      pipeline.Idempotent,
      "input:machine-b",
    ),
    pipeline.StepSucceeded(
      observe,
      1,
      "observe-1",
      "machine-b",
      "machine-b",
      [],
    ),
  ]
}

pub fn apply_all(
  definition: Pipeline,
  events: List(Event),
) -> Result(Projection, pipeline.KernelError) {
  apply_loop(definition, events, pipeline.Empty)
}

fn apply_loop(
  definition: Pipeline,
  events: List(Event),
  projection: Projection,
) -> Result(Projection, pipeline.KernelError) {
  case events {
    [] -> Ok(projection)
    [event, ..rest] ->
      case pipeline.apply(definition, projection, event) {
        Ok(next) -> apply_loop(definition, rest, next)
        Error(error) -> Error(error)
      }
  }
}
