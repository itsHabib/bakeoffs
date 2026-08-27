import gleam/json
import journal
import pipeline.{type Projection}

pub fn encode(projection: Projection) -> String {
  case projection {
    pipeline.Empty ->
      json.object([
        #("schema", json.string("PipelineRunViewV1")),
        #("status", json.string("empty")),
      ])
      |> json.to_string
    pipeline.Run(
      run_id:,
      pipeline_name:,
      pipeline_version:,
      pipeline_digest:,
      subject:,
      current_step:,
      attempt:,
      status:,
      evidence:,
      journal_sequence:,
    ) ->
      json.object([
        #("schema", json.string("PipelineRunViewV1")),
        #("run_id", json.string(run_id)),
        #("pipeline", json.string(pipeline_name)),
        #("pipeline_version", json.int(pipeline_version)),
        #("pipeline_digest", json.string(pipeline_digest)),
        #("subject", json.string(subject)),
        #("step", json.string(pipeline.step_name(current_step))),
        #("attempt", json.int(attempt)),
        #("status", json.string(status_name(status))),
        #("journal_sequence", json.int(journal_sequence)),
        #("evidence", json.array(evidence, evidence_json)),
      ])
      |> json.to_string
  }
}

fn evidence_json(evidence: pipeline.Evidence) -> json.Json {
  json.object([
    #("claim", json.string(evidence.claim)),
    #("subject", json.string(evidence.subject)),
    #("verdict", json.string(journal.verdict_name(evidence.verdict))),
    #("input_digest", json.string(evidence.recipe.input_digest)),
  ])
}

fn status_name(status: pipeline.RunStatus) -> String {
  case status {
    pipeline.Ready -> "ready"
    pipeline.InFlight(_, _, effect_id, replay_class, _) ->
      "in_flight:"
      <> effect_id
      <> ":"
      <> journal.replay_class_name(replay_class)
    pipeline.Finished -> "finished"
    pipeline.Parked(question) -> "parked:" <> question
  }
}
