import examples
import gleam/io
import journal
import pipeline
import view

pub fn main() -> Nil {
  let path = "tmp/demo/shipping.ndjson"
  let assert Ok(Nil) = journal.delete(path)
  let definition = examples.shipping()
  let assert Ok(executed) =
    journal.write_events(path, definition, examples.shipping_trace())
  let assert Ok(replayed) = journal.replay(path, definition)
  assert executed == replayed
  let assert pipeline.Run(status: pipeline.Finished, ..) = replayed
  io.println(view.encode(replayed))
}
