import gleam/dynamic/decode
import gleam/json
import gleam/list
import gleam/result
import gleam/string
import pipeline.{type Event, type Pipeline, type Projection}

pub const envelope_version = 1

pub type Envelope {
  Envelope(version: Int, sequence: Int, event: Event)
}

pub type JournalError {
  IoFailure(String)
  CorruptRecord(line: Int, detail: String)
  SequenceGap(expected: Int, actual: Int)
  InvalidTransition(sequence: Int, error: pipeline.KernelError)
}

pub fn append(
  path: String,
  sequence: Int,
  event: Event,
) -> Result(Nil, JournalError) {
  encode(Envelope(envelope_version, sequence, event))
  |> append_sync(path, _)
  |> result.map_error(IoFailure)
}

pub fn replay(
  path: String,
  definition: Pipeline,
) -> Result(Projection, JournalError) {
  use envelopes <- result.try(read_envelopes(path))
  replay_envelopes(envelopes, definition)
}

pub fn write_events(
  path: String,
  definition: Pipeline,
  events: List(Event),
) -> Result(Projection, JournalError) {
  write_loop(path, definition, events, pipeline.Empty)
}

fn write_loop(
  path: String,
  definition: Pipeline,
  events: List(Event),
  projection: Projection,
) -> Result(Projection, JournalError) {
  case events {
    [] -> Ok(projection)
    [event, ..rest] -> {
      let sequence = pipeline.next_sequence(projection)
      use next <- result.try(
        pipeline.apply(definition, projection, event)
        |> result.map_error(InvalidTransition(sequence, _)),
      )
      use _ <- result.try(append(path, sequence, event))
      write_loop(path, definition, rest, next)
    }
  }
}

pub fn read_envelopes(path: String) -> Result(List(Envelope), JournalError) {
  use lines <- result.try(read_lines(path) |> result.map_error(IoFailure))
  decode_lines(lines, 1, 1, [])
}

fn decode_lines(
  lines: List(String),
  line_number: Int,
  expected_sequence: Int,
  reversed: List(Envelope),
) -> Result(List(Envelope), JournalError) {
  case lines {
    [] -> Ok(list.reverse(reversed))
    [line, ..rest] ->
      case decode_envelope(line) {
        Error(detail) -> Error(CorruptRecord(line_number, detail))
        Ok(envelope) if envelope.sequence != expected_sequence ->
          Error(SequenceGap(expected_sequence, envelope.sequence))
        Ok(envelope) ->
          decode_lines(rest, line_number + 1, expected_sequence + 1, [
            envelope,
            ..reversed
          ])
      }
  }
}

pub fn replay_envelopes(
  envelopes: List(Envelope),
  definition: Pipeline,
) -> Result(Projection, JournalError) {
  replay_loop(envelopes, definition, pipeline.Empty)
}

fn replay_loop(
  envelopes: List(Envelope),
  definition: Pipeline,
  projection: Projection,
) -> Result(Projection, JournalError) {
  case envelopes {
    [] -> Ok(projection)
    [Envelope(sequence:, event:, ..), ..rest] ->
      case pipeline.apply(definition, projection, event) {
        Ok(next) -> replay_loop(rest, definition, next)
        Error(error) -> Error(InvalidTransition(sequence, error))
      }
  }
}

pub fn encode(envelope: Envelope) -> String {
  json.object([
    #("version", json.int(envelope.version)),
    #("sequence", json.int(envelope.sequence)),
    #("event", encode_event(envelope.event)),
  ])
  |> json.to_string
}

pub fn decode_envelope(line: String) -> Result(Envelope, String) {
  case json.parse(line, envelope_decoder()) {
    Ok(envelope) -> Ok(envelope)
    Error(error) -> Error(string.inspect(error))
  }
}

fn envelope_decoder() -> decode.Decoder(Envelope) {
  use version <- decode.field("version", decode.int)
  use sequence <- decode.field("sequence", decode.int)
  use event <- decode.field("event", event_decoder())
  case version == envelope_version, sequence > 0 {
    True, True -> decode.success(Envelope(version, sequence, event))
    _, _ ->
      decode.failure(
        Envelope(0, 0, pipeline.RunStarted("", "", 0, "", "")),
        expected: "PipelineJournalV1 version=1 sequence>0",
      )
  }
}

fn encode_event(event: Event) -> json.Json {
  case event {
    pipeline.RunStarted(run_id, name, version, digest, subject) ->
      json.object([
        #("type", json.string("run_started")),
        #("run_id", json.string(run_id)),
        #("pipeline_name", json.string(name)),
        #("pipeline_version", json.int(version)),
        #("pipeline_digest", json.string(digest)),
        #("subject", json.string(subject)),
      ])
    pipeline.EffectPrepared(
      step,
      attempt,
      effect_id,
      replay_class,
      input_digest,
    ) ->
      json.object([
        #("type", json.string("effect_prepared")),
        #("step", json.string(pipeline.step_name(step))),
        #("attempt", json.int(attempt)),
        #("effect_id", json.string(effect_id)),
        #("replay_class", json.string(replay_class_name(replay_class))),
        #("input_digest", json.string(input_digest)),
      ])
    pipeline.StepSucceeded(
      step,
      attempt,
      effect_id,
      previous_subject,
      subject,
      evidence,
    ) ->
      json.object([
        #("type", json.string("step_succeeded")),
        #("step", json.string(pipeline.step_name(step))),
        #("attempt", json.int(attempt)),
        #("effect_id", json.string(effect_id)),
        #("previous_subject", json.string(previous_subject)),
        #("subject", json.string(subject)),
        #("evidence", json.array(evidence, encode_evidence)),
      ])
    pipeline.StepFailed(step, attempt, effect_id, subject, reason, evidence) ->
      json.object([
        #("type", json.string("step_failed")),
        #("step", json.string(pipeline.step_name(step))),
        #("attempt", json.int(attempt)),
        #("effect_id", json.string(effect_id)),
        #("subject", json.string(subject)),
        #("reason", json.string(reason)),
        #("evidence", json.array(evidence, encode_evidence)),
      ])
    pipeline.StepParked(step, attempt, effect_id, subject, question) ->
      json.object([
        #("type", json.string("step_parked")),
        #("step", json.string(pipeline.step_name(step))),
        #("attempt", json.int(attempt)),
        #("effect_id", json.string(effect_id)),
        #("subject", json.string(subject)),
        #("question", json.string(question)),
      ])
  }
}

fn event_decoder() -> decode.Decoder(Event) {
  use event_type <- decode.field("type", decode.string)
  case event_type {
    "run_started" -> {
      use run_id <- decode.field("run_id", decode.string)
      use name <- decode.field("pipeline_name", decode.string)
      use version <- decode.field("pipeline_version", decode.int)
      use digest <- decode.field("pipeline_digest", decode.string)
      use subject <- decode.field("subject", decode.string)
      decode.success(pipeline.RunStarted(run_id, name, version, digest, subject))
    }
    "effect_prepared" -> {
      use step <- decode.field("step", decode.string)
      use attempt <- decode.field("attempt", decode.int)
      use effect_id <- decode.field("effect_id", decode.string)
      use replay_class <- decode.field("replay_class", replay_class_decoder())
      use input_digest <- decode.field("input_digest", decode.string)
      decode.success(pipeline.EffectPrepared(
        pipeline.step(step),
        attempt,
        effect_id,
        replay_class,
        input_digest,
      ))
    }
    "step_succeeded" -> {
      use step <- decode.field("step", decode.string)
      use attempt <- decode.field("attempt", decode.int)
      use effect_id <- decode.field("effect_id", decode.string)
      use previous_subject <- decode.field("previous_subject", decode.string)
      use subject <- decode.field("subject", decode.string)
      use evidence <- decode.field("evidence", decode.list(evidence_decoder()))
      decode.success(pipeline.StepSucceeded(
        pipeline.step(step),
        attempt,
        effect_id,
        previous_subject,
        subject,
        evidence,
      ))
    }
    "step_failed" -> {
      use step <- decode.field("step", decode.string)
      use attempt <- decode.field("attempt", decode.int)
      use effect_id <- decode.field("effect_id", decode.string)
      use subject <- decode.field("subject", decode.string)
      use reason <- decode.field("reason", decode.string)
      use evidence <- decode.field("evidence", decode.list(evidence_decoder()))
      decode.success(pipeline.StepFailed(
        pipeline.step(step),
        attempt,
        effect_id,
        subject,
        reason,
        evidence,
      ))
    }
    "step_parked" -> {
      use step <- decode.field("step", decode.string)
      use attempt <- decode.field("attempt", decode.int)
      use effect_id <- decode.field("effect_id", decode.string)
      use subject <- decode.field("subject", decode.string)
      use question <- decode.field("question", decode.string)
      decode.success(pipeline.StepParked(
        pipeline.step(step),
        attempt,
        effect_id,
        subject,
        question,
      ))
    }
    _ ->
      decode.failure(
        pipeline.RunStarted("", "", 0, "", ""),
        expected: "known pipeline event",
      )
  }
}

fn encode_evidence(evidence: pipeline.Evidence) -> json.Json {
  json.object([
    #("claim", json.string(evidence.claim)),
    #("subject", json.string(evidence.subject)),
    #("verdict", json.string(verdict_name(evidence.verdict))),
    #("recipe", encode_recipe(evidence.recipe)),
  ])
}

fn evidence_decoder() -> decode.Decoder(pipeline.Evidence) {
  use claim <- decode.field("claim", decode.string)
  use subject <- decode.field("subject", decode.string)
  use verdict <- decode.field("verdict", verdict_decoder())
  use recipe <- decode.field("recipe", recipe_decoder())
  decode.success(pipeline.Evidence(claim, subject, verdict, recipe))
}

fn encode_recipe(recipe: pipeline.Recipe) -> json.Json {
  json.object([
    #("tool", json.string(recipe.tool)),
    #("version", json.string(recipe.version)),
    #("command", json.string(recipe.command)),
    #("input_digest", json.string(recipe.input_digest)),
    #("seed", json.string(recipe.seed)),
    #("bounds", json.array(recipe.bounds, json.string)),
  ])
}

fn recipe_decoder() -> decode.Decoder(pipeline.Recipe) {
  use tool <- decode.field("tool", decode.string)
  use version <- decode.field("version", decode.string)
  use command <- decode.field("command", decode.string)
  use input_digest <- decode.field("input_digest", decode.string)
  use seed <- decode.field("seed", decode.string)
  use bounds <- decode.field("bounds", decode.list(decode.string))
  decode.success(pipeline.Recipe(
    tool,
    version,
    command,
    input_digest,
    seed,
    bounds,
  ))
}

fn replay_class_decoder() -> decode.Decoder(pipeline.ReplayClass) {
  use value <- decode.then(decode.string)
  case value {
    "idempotent" -> decode.success(pipeline.Idempotent)
    "deduplicated" -> decode.success(pipeline.Deduplicated)
    "at_least_once" -> decode.success(pipeline.AtLeastOnce)
    "manual" -> decode.success(pipeline.Manual)
    _ -> decode.failure(pipeline.Manual, expected: "known replay class")
  }
}

fn verdict_decoder() -> decode.Decoder(pipeline.Verdict) {
  use value <- decode.then(decode.string)
  case value {
    "supported" -> decode.success(pipeline.Supported)
    "refuted" -> decode.success(pipeline.Refuted)
    "insufficient" -> decode.success(pipeline.Insufficient)
    _ -> decode.failure(pipeline.Insufficient, expected: "known verdict")
  }
}

pub fn replay_class_name(replay_class: pipeline.ReplayClass) -> String {
  case replay_class {
    pipeline.Idempotent -> "idempotent"
    pipeline.Deduplicated -> "deduplicated"
    pipeline.AtLeastOnce -> "at_least_once"
    pipeline.Manual -> "manual"
  }
}

pub fn verdict_name(verdict: pipeline.Verdict) -> String {
  case verdict {
    pipeline.Supported -> "supported"
    pipeline.Refuted -> "refuted"
    pipeline.Insufficient -> "insufficient"
  }
}

pub fn write_raw(path: String, bytes: String) -> Result(Nil, String) {
  write_raw_ffi(path, bytes)
}

pub fn read_raw(path: String) -> Result(String, String) {
  read_raw_ffi(path)
}

pub fn delete(path: String) -> Result(Nil, String) {
  delete_ffi(path)
}

pub fn inject_torn_write(path: String, bytes: String) -> Result(Nil, String) {
  append_torn(path, bytes)
}

@external(erlang, "durable_pipeline_kernel_ffi", "append_sync")
fn append_sync(path: String, line: String) -> Result(Nil, String)

@external(erlang, "durable_pipeline_kernel_ffi", "read_lines_recover")
fn read_lines(path: String) -> Result(List(String), String)

@external(erlang, "durable_pipeline_kernel_ffi", "append_torn")
fn append_torn(path: String, bytes: String) -> Result(Nil, String)

@external(erlang, "durable_pipeline_kernel_ffi", "write_raw")
fn write_raw_ffi(path: String, bytes: String) -> Result(Nil, String)

@external(erlang, "durable_pipeline_kernel_ffi", "read_raw")
fn read_raw_ffi(path: String) -> Result(String, String)

@external(erlang, "durable_pipeline_kernel_ffi", "delete")
fn delete_ffi(path: String) -> Result(Nil, String)
