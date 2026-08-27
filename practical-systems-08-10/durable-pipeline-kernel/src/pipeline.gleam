import gleam/list

pub opaque type StepId {
  StepId(String)
}

pub fn step(name: String) -> StepId {
  StepId(name)
}

pub fn step_name(id: StepId) -> String {
  let StepId(name) = id
  name
}

pub type ReplayClass {
  Idempotent
  Deduplicated
  AtLeastOnce
  Manual
}

pub type Verdict {
  Supported
  Refuted
  Insufficient
}

pub type Recipe {
  Recipe(
    tool: String,
    version: String,
    command: String,
    input_digest: String,
    seed: String,
    bounds: List(String),
  )
}

pub type Evidence {
  Evidence(claim: String, subject: String, verdict: Verdict, recipe: Recipe)
}

pub type OutcomeKind {
  Success
  Failure
}

pub type Transition {
  GoTo(step: StepId, new_attempt: Bool)
  Finish
  Stop(reason: String)
}

pub type Rule {
  Rule(from: StepId, outcome: OutcomeKind, transition: Transition)
}

pub type StepSpec {
  StepSpec(id: StepId, replay_class: ReplayClass, requires: List(String))
}

pub type Pipeline {
  Pipeline(
    name: String,
    version: Int,
    digest: String,
    start: StepId,
    steps: List(StepSpec),
    rules: List(Rule),
  )
}

pub type RunStatus {
  Ready
  InFlight(
    step: StepId,
    attempt: Int,
    effect_id: String,
    replay_class: ReplayClass,
    input_digest: String,
  )
  Finished
  Parked(question: String)
}

pub type Projection {
  Empty
  Run(
    run_id: String,
    pipeline_name: String,
    pipeline_version: Int,
    pipeline_digest: String,
    subject: String,
    current_step: StepId,
    attempt: Int,
    status: RunStatus,
    evidence: List(Evidence),
    journal_sequence: Int,
  )
}

pub type Event {
  RunStarted(
    run_id: String,
    pipeline_name: String,
    pipeline_version: Int,
    pipeline_digest: String,
    subject: String,
  )
  EffectPrepared(
    step: StepId,
    attempt: Int,
    effect_id: String,
    replay_class: ReplayClass,
    input_digest: String,
  )
  StepSucceeded(
    step: StepId,
    attempt: Int,
    effect_id: String,
    previous_subject: String,
    subject: String,
    evidence: List(Evidence),
  )
  StepFailed(
    step: StepId,
    attempt: Int,
    effect_id: String,
    subject: String,
    reason: String,
    evidence: List(Evidence),
  )
  StepParked(
    step: StepId,
    attempt: Int,
    effect_id: String,
    subject: String,
    question: String,
  )
}

pub type KernelError {
  InvalidStart(String)
  PipelineMismatch(String)
  UnknownStep(String)
  MissingRule(String)
  MissingEvidence(claim: String, subject: String)
  InvalidEvidence(String)
  InvalidTransition(String)
}

pub type RecoveryAction {
  Execute(step: StepId, attempt: Int)
  RetrySameKey(step: StepId, attempt: Int, effect_id: String)
  ReconcilePossibleDuplicate(step: StepId, attempt: Int, effect_id: String)
  ReconcileManually(step: StepId, attempt: Int, effect_id: String)
  AwaitDecision(question: String)
  NothingToDo
}

pub fn next_sequence(projection: Projection) -> Int {
  case projection {
    Empty -> 1
    Run(journal_sequence:, ..) -> journal_sequence + 1
  }
}

pub fn prepare(
  definition: Pipeline,
  projection: Projection,
  effect_id: String,
  input_digest: String,
) -> Result(Event, KernelError) {
  case projection {
    Empty -> Error(InvalidTransition("run has not started"))
    Run(subject:, current_step:, attempt:, status: Ready, evidence:, ..) -> {
      use spec <- result_try(find_step(definition.steps, current_step))
      use _ <- result_try(require_evidence(spec.requires, subject, evidence))
      case effect_id != "", input_digest != "" {
        True, True ->
          Ok(EffectPrepared(
            current_step,
            attempt,
            effect_id,
            spec.replay_class,
            input_digest,
          ))
        _, _ ->
          Error(InvalidTransition(
            "effect identity and input digest are required",
          ))
      }
    }
    Run(status: InFlight(..), ..) ->
      Error(InvalidTransition("an effect is already in flight"))
    Run(status: Finished, ..) -> Error(InvalidTransition("run is complete"))
    Run(status: Parked(_), ..) -> Error(InvalidTransition("run is parked"))
  }
}

pub fn apply(
  definition: Pipeline,
  projection: Projection,
  event: Event,
) -> Result(Projection, KernelError) {
  case projection, event {
    Empty, RunStarted(run_id, name, version, digest, subject) ->
      start(definition, run_id, name, version, digest, subject)

    Run(
      run_id:,
      pipeline_name:,
      pipeline_version:,
      pipeline_digest:,
      subject:,
      current_step:,
      attempt:,
      status: Ready,
      evidence:,
      journal_sequence:,
    ),
      EffectPrepared(step, event_attempt, effect_id, replay_class, input_digest)
      if step == current_step && event_attempt == attempt
    -> {
      use spec <- result_try(find_step(definition.steps, step))
      use _ <- result_try(require_evidence(spec.requires, subject, evidence))
      case
        replay_class == spec.replay_class,
        effect_id != "",
        input_digest != ""
      {
        True, True, True ->
          Ok(Run(
            run_id,
            pipeline_name,
            pipeline_version,
            pipeline_digest,
            subject,
            current_step,
            attempt,
            InFlight(step, attempt, effect_id, replay_class, input_digest),
            evidence,
            journal_sequence + 1,
          ))
        _, _, _ ->
          Error(InvalidTransition(
            "prepared effect does not match the step contract",
          ))
      }
    }

    Run(
      run_id:,
      pipeline_name:,
      pipeline_version:,
      pipeline_digest:,
      subject:,
      current_step:,
      attempt:,
      status: InFlight(in_flight_step, in_flight_attempt, in_flight_id, _, _),
      evidence: existing_evidence,
      journal_sequence:,
    ),
      StepSucceeded(
        step,
        event_attempt,
        effect_id,
        previous_subject,
        next_subject,
        new_evidence,
      )
      if step == current_step
      && step == in_flight_step
      && event_attempt == attempt
      && event_attempt == in_flight_attempt
      && effect_id == in_flight_id
      && previous_subject == subject
    -> {
      use _ <- result_try(validate_evidence(new_evidence, next_subject))
      use transition <- result_try(find_rule(definition.rules, step, Success))
      transition_projection(
        run_id,
        pipeline_name,
        pipeline_version,
        pipeline_digest,
        next_subject,
        attempt,
        list.append(existing_evidence, new_evidence),
        journal_sequence + 1,
        transition,
      )
    }

    Run(
      run_id:,
      pipeline_name:,
      pipeline_version:,
      pipeline_digest:,
      subject:,
      current_step:,
      attempt:,
      status: InFlight(in_flight_step, in_flight_attempt, in_flight_id, _, _),
      evidence: existing_evidence,
      journal_sequence:,
    ),
      StepFailed(
        step,
        event_attempt,
        effect_id,
        event_subject,
        reason,
        new_evidence,
      )
      if step == current_step
      && step == in_flight_step
      && event_attempt == attempt
      && event_attempt == in_flight_attempt
      && effect_id == in_flight_id
      && event_subject == subject
    -> {
      use _ <- result_try(validate_evidence(new_evidence, subject))
      use transition <- result_try(find_rule(definition.rules, step, Failure))
      case reason == "" {
        True -> Error(InvalidTransition("a failed step requires a reason"))
        False ->
          transition_projection(
            run_id,
            pipeline_name,
            pipeline_version,
            pipeline_digest,
            subject,
            attempt,
            list.append(existing_evidence, new_evidence),
            journal_sequence + 1,
            transition,
          )
      }
    }

    Run(
      run_id:,
      pipeline_name:,
      pipeline_version:,
      pipeline_digest:,
      subject:,
      current_step:,
      attempt:,
      status: InFlight(in_flight_step, in_flight_attempt, in_flight_id, _, _),
      evidence:,
      journal_sequence:,
    ),
      StepParked(step, event_attempt, effect_id, event_subject, question)
      if step == current_step
      && step == in_flight_step
      && event_attempt == attempt
      && event_attempt == in_flight_attempt
      && effect_id == in_flight_id
      && event_subject == subject
      && question != ""
    ->
      Ok(Run(
        run_id,
        pipeline_name,
        pipeline_version,
        pipeline_digest,
        subject,
        current_step,
        attempt,
        Parked(question),
        evidence,
        journal_sequence + 1,
      ))

    _, _ ->
      Error(InvalidTransition(
        "event is impossible for the recovered projection",
      ))
  }
}

pub fn recovery_action(projection: Projection) -> RecoveryAction {
  case projection {
    Empty -> NothingToDo
    Run(current_step:, attempt:, status: Ready, ..) ->
      Execute(current_step, attempt)
    Run(status: InFlight(step, attempt, effect_id, replay_class, _), ..) ->
      case replay_class {
        Idempotent | Deduplicated -> RetrySameKey(step, attempt, effect_id)
        AtLeastOnce -> ReconcilePossibleDuplicate(step, attempt, effect_id)
        Manual -> ReconcileManually(step, attempt, effect_id)
      }
    Run(status: Parked(question), ..) -> AwaitDecision(question)
    Run(status: Finished, ..) -> NothingToDo
  }
}

pub fn claim_verdict(
  evidence: List(Evidence),
  subject: String,
  claim: String,
) -> Verdict {
  claim_verdict_loop(evidence, subject, claim, False, False)
}

fn claim_verdict_loop(
  evidence: List(Evidence),
  subject: String,
  claim: String,
  supported: Bool,
  refuted: Bool,
) -> Verdict {
  case evidence {
    [] if refuted -> Refuted
    [] if supported -> Supported
    [] -> Insufficient
    [Evidence(claim: found_claim, subject: found_subject, verdict:, ..), ..rest] -> {
      let relevant = found_claim == claim && found_subject == subject
      claim_verdict_loop(
        rest,
        subject,
        claim,
        supported || { relevant && verdict == Supported },
        refuted || { relevant && verdict == Refuted },
      )
    }
  }
}

fn start(
  definition: Pipeline,
  run_id: String,
  name: String,
  version: Int,
  digest: String,
  subject: String,
) -> Result(Projection, KernelError) {
  case run_id != "", subject != "" {
    False, _ | _, False ->
      Error(InvalidStart("run id and subject are required"))
    True, True ->
      case
        name == definition.name,
        version == definition.version,
        digest == definition.digest
      {
        True, True, True ->
          Ok(Run(
            run_id,
            name,
            version,
            digest,
            subject,
            definition.start,
            1,
            Ready,
            [],
            1,
          ))
        _, _, _ ->
          Error(PipelineMismatch(
            "journal is bound to another pipeline definition",
          ))
      }
  }
}

fn transition_projection(
  run_id: String,
  pipeline_name: String,
  pipeline_version: Int,
  pipeline_digest: String,
  subject: String,
  attempt: Int,
  evidence: List(Evidence),
  sequence: Int,
  transition: Transition,
) -> Result(Projection, KernelError) {
  case transition {
    GoTo(next, new_attempt) ->
      Ok(Run(
        run_id,
        pipeline_name,
        pipeline_version,
        pipeline_digest,
        subject,
        next,
        case new_attempt {
          True -> attempt + 1
          False -> attempt
        },
        Ready,
        evidence,
        sequence,
      ))
    Finish ->
      Ok(Run(
        run_id,
        pipeline_name,
        pipeline_version,
        pipeline_digest,
        subject,
        step("done"),
        attempt,
        Finished,
        evidence,
        sequence,
      ))
    Stop(reason) ->
      Ok(Run(
        run_id,
        pipeline_name,
        pipeline_version,
        pipeline_digest,
        subject,
        step("stopped"),
        attempt,
        Parked(reason),
        evidence,
        sequence,
      ))
  }
}

fn find_step(
  steps: List(StepSpec),
  id: StepId,
) -> Result(StepSpec, KernelError) {
  case steps {
    [] -> Error(UnknownStep(step_name(id)))
    [StepSpec(id: candidate, ..) as spec, ..rest] ->
      case candidate == id {
        True -> Ok(spec)
        False -> find_step(rest, id)
      }
  }
}

fn find_rule(
  rules: List(Rule),
  from: StepId,
  outcome: OutcomeKind,
) -> Result(Transition, KernelError) {
  case rules {
    [] -> Error(MissingRule(step_name(from)))
    [Rule(from: candidate, outcome: candidate_outcome, transition:), ..rest] ->
      case candidate == from && candidate_outcome == outcome {
        True -> Ok(transition)
        False -> find_rule(rest, from, outcome)
      }
  }
}

fn require_evidence(
  requirements: List(String),
  subject: String,
  evidence: List(Evidence),
) -> Result(Nil, KernelError) {
  case requirements {
    [] -> Ok(Nil)
    [claim, ..rest] ->
      case claim_verdict(evidence, subject, claim) {
        Supported -> require_evidence(rest, subject, evidence)
        Refuted | Insufficient -> Error(MissingEvidence(claim, subject))
      }
  }
}

fn validate_evidence(
  evidence: List(Evidence),
  subject: String,
) -> Result(Nil, KernelError) {
  case evidence {
    [] -> Ok(Nil)
    [Evidence(claim:, subject: evidence_subject, recipe:, ..), ..rest] ->
      case evidence_subject == subject, recipe_valid(recipe), claim != "" {
        True, True, True -> validate_evidence(rest, subject)
        _, _, _ ->
          Error(InvalidEvidence(
            "evidence is incomplete or bound to another subject",
          ))
      }
  }
}

fn recipe_valid(recipe: Recipe) -> Bool {
  recipe.tool != ""
  && recipe.version != ""
  && recipe.command != ""
  && recipe.input_digest != ""
  && recipe.seed != ""
  && recipe.bounds != []
}

fn result_try(
  result: Result(a, e),
  next: fn(a) -> Result(b, e),
) -> Result(b, e) {
  case result {
    Ok(value) -> next(value)
    Error(error) -> Error(error)
  }
}
