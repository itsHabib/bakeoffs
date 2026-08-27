import examples
import journal
import pipeline

pub fn shipping_failure_loops_and_finishes_on_new_subject_test() {
  let assert Ok(projection) =
    examples.apply_all(examples.shipping(), examples.shipping_trace())
  let assert pipeline.Run(
    subject: "rev-b",
    attempt: 2,
    status: pipeline.Finished,
    evidence:,
    journal_sequence: 17,
    ..,
  ) = projection
  assert pipeline.claim_verdict(evidence, "rev-a", "e2e") == pipeline.Refuted
  assert pipeline.claim_verdict(evidence, "rev-b", "e2e") == pipeline.Supported
}

pub fn stale_validation_cannot_unlock_new_subject_test() {
  let definition = examples.shipping()
  let ship = pipeline.step("ship")
  let validate = pipeline.step("validate")
  let e2e = pipeline.step("e2e")
  let assert Ok(p1) =
    pipeline.apply(
      definition,
      pipeline.Empty,
      pipeline.RunStarted(
        "stale-evidence",
        "shipping",
        1,
        examples.shipping_digest,
        "base-a",
      ),
    )
  let assert Ok(p2) =
    pipeline.apply(
      definition,
      p1,
      pipeline.EffectPrepared(
        ship,
        1,
        "ship-1",
        pipeline.Deduplicated,
        "base-a",
      ),
    )
  let assert Ok(p3) =
    pipeline.apply(
      definition,
      p2,
      pipeline.StepSucceeded(ship, 1, "ship-1", "base-a", "rev-a", []),
    )
  let assert Ok(p4) =
    pipeline.apply(
      definition,
      p3,
      pipeline.EffectPrepared(
        validate,
        1,
        "validate-1",
        pipeline.Idempotent,
        "rev-a",
      ),
    )
  let assert Ok(p5) =
    pipeline.apply(
      definition,
      p4,
      pipeline.StepSucceeded(validate, 1, "validate-1", "rev-a", "rev-a", [
        examples.supported("unit", "rev-a"),
      ]),
    )
  let assert Ok(p6) =
    pipeline.apply(
      definition,
      p5,
      pipeline.EffectPrepared(e2e, 1, "e2e-1", pipeline.Idempotent, "rev-a"),
    )
  let assert Ok(p7) =
    pipeline.apply(
      definition,
      p6,
      pipeline.StepFailed(e2e, 1, "e2e-1", "rev-a", "failed", [
        examples.refuted("e2e", "rev-a"),
      ]),
    )
  let assert Ok(p8) =
    pipeline.apply(
      definition,
      p7,
      pipeline.EffectPrepared(ship, 2, "ship-2", pipeline.Deduplicated, "rev-a"),
    )
  let assert Ok(p9) =
    pipeline.apply(
      definition,
      p8,
      pipeline.StepSucceeded(ship, 2, "ship-2", "rev-a", "rev-b", []),
    )
  let assert Ok(p10) =
    pipeline.apply(
      definition,
      p9,
      pipeline.EffectPrepared(
        validate,
        2,
        "validate-2",
        pipeline.Idempotent,
        "rev-b",
      ),
    )
  let assert Ok(p11) =
    pipeline.apply(
      definition,
      p10,
      pipeline.StepSucceeded(validate, 2, "validate-2", "rev-b", "rev-b", []),
    )
  let assert Error(pipeline.MissingEvidence("unit", "rev-b")) =
    pipeline.prepare(definition, p11, "e2e-2", "rev-b")
}

pub fn prepared_deduplicated_effect_reuses_the_same_key_test() {
  let definition = examples.shipping()
  let ship = pipeline.step("ship")
  let assert Ok(started) =
    pipeline.apply(
      definition,
      pipeline.Empty,
      pipeline.RunStarted(
        "crash-demo",
        "shipping",
        1,
        examples.shipping_digest,
        "base-a",
      ),
    )
  let assert Ok(prepared) =
    pipeline.apply(
      definition,
      started,
      pipeline.EffectPrepared(
        ship,
        1,
        "stable-dispatch-key",
        pipeline.Deduplicated,
        "base-a",
      ),
    )
  assert pipeline.recovery_action(prepared)
    == pipeline.RetrySameKey(ship, 1, "stable-dispatch-key")
}

pub fn at_least_once_effect_parks_for_possible_duplication_test() {
  let send = pipeline.step("send")
  let definition =
    pipeline.Pipeline(
      "notification",
      1,
      "sha256:notification-v1",
      send,
      [pipeline.StepSpec(send, pipeline.AtLeastOnce, [])],
      [
        pipeline.Rule(send, pipeline.Success, pipeline.Finish),
        pipeline.Rule(send, pipeline.Failure, pipeline.Stop("send failed")),
      ],
    )
  let assert Ok(started) =
    pipeline.apply(
      definition,
      pipeline.Empty,
      pipeline.RunStarted(
        "notify",
        "notification",
        1,
        "sha256:notification-v1",
        "message-a",
      ),
    )
  let assert Ok(prepared) =
    pipeline.apply(
      definition,
      started,
      pipeline.EffectPrepared(
        send,
        1,
        "send-1",
        pipeline.AtLeastOnce,
        "message-a",
      ),
    )
  assert pipeline.recovery_action(prepared)
    == pipeline.ReconcilePossibleDuplicate(send, 1, "send-1")
}

pub fn maintenance_uses_the_same_kernel_test() {
  let assert Ok(projection) =
    examples.apply_all(examples.maintenance(), examples.maintenance_trace())
  let assert pipeline.Run(
    pipeline_name: "maintenance",
    subject: "machine-b",
    status: pipeline.Finished,
    ..,
  ) = projection
}

pub fn journal_replay_matches_in_memory_execution_test() {
  let path = "tmp/tests/shipping-replay.ndjson"
  let assert Ok(Nil) = journal.delete(path)
  let definition = examples.shipping()
  let assert Ok(executed) =
    journal.write_events(path, definition, examples.shipping_trace())
  let assert Ok(replayed) = journal.replay(path, definition)
  assert executed == replayed
}

pub fn replay_refuses_pipeline_definition_drift_test() {
  let path = "tmp/tests/pipeline-drift.ndjson"
  let assert Ok(Nil) = journal.delete(path)
  let original = examples.shipping()
  let assert Ok(_) =
    journal.write_events(path, original, examples.shipping_trace())
  let pipeline.Pipeline(name, version, _, start, steps, rules) = original
  let changed =
    pipeline.Pipeline(name, version, "sha256:changed", start, steps, rules)
  let assert Error(journal.InvalidTransition(1, pipeline.PipelineMismatch(_))) =
    journal.replay(path, changed)
}

pub fn torn_tail_is_removed_before_replay_test() {
  let path = "tmp/tests/torn-tail.ndjson"
  let assert Ok(Nil) = journal.delete(path)
  let definition = examples.shipping()
  let start =
    pipeline.RunStarted(
      "torn",
      "shipping",
      1,
      examples.shipping_digest,
      "base-a",
    )
  let assert Ok(Nil) = journal.append(path, 1, start)
  let torn =
    journal.encode(journal.Envelope(
      1,
      2,
      pipeline.EffectPrepared(
        pipeline.step("ship"),
        1,
        "not-committed",
        pipeline.Deduplicated,
        "base-a",
      ),
    ))
  let assert Ok(Nil) = journal.inject_torn_write(path, torn)
  let assert Ok(replayed) = journal.replay(path, definition)
  let assert pipeline.Run(status: pipeline.Ready, journal_sequence: 1, ..) =
    replayed
}

pub fn invalid_transition_is_never_appended_test() {
  let path = "tmp/tests/refuse-invalid-append.ndjson"
  let assert Ok(Nil) = journal.delete(path)
  let definition = examples.shipping()
  let ship = pipeline.step("ship")
  let events = [
    pipeline.RunStarted(
      "invalid-append",
      "shipping",
      1,
      examples.shipping_digest,
      "base-a",
    ),
    pipeline.StepSucceeded(ship, 1, "never-prepared", "base-a", "rev-a", []),
  ]
  let assert Error(journal.InvalidTransition(2, _)) =
    journal.write_events(path, definition, events)
  let assert Ok([journal.Envelope(sequence: 1, ..)]) =
    journal.read_envelopes(path)
}
