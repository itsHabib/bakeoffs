#!/bin/sh
set -eu
cd "$(dirname "$0")"
printf 'category\tphysical_lines\tnonblank_non_line_comment_lines\n'
count() {
  category=$1
  shift
  awk -v category="$category" ' { physical++ } NF && $0 !~ /^[[:space:]]*\/\// { code++ } END { printf "%s\t%d\t%d\n", category, physical, code }' "$@"
}
count consumer_and_recovery model.go journal.go worker.go main.go
count local_sink sink.go
count controller_and_oracles controller.go
count entry_tests main_test.go
count reused_core ../warrant/action.go ../warrant/event.go ../warrant/pipeline.go ../warrant/reduce.go ../warrant/view.go
count upstream_runner_compiled_but_unused ../warrant/runner.go
count reused_tests_executed ../warrant/conformance_test.go ../warrant/runner_test.go
count new_Go_total ./*.go
count new_Go_plus_reused_package ./*.go ../warrant/action.go ../warrant/event.go ../warrant/pipeline.go ../warrant/reduce.go ../warrant/view.go ../warrant/runner.go
count including_upstream_tests ./*.go ../warrant/action.go ../warrant/event.go ../warrant/pipeline.go ../warrant/reduce.go ../warrant/view.go ../warrant/runner.go ../warrant/conformance_test.go ../warrant/runner_test.go
# Shell lines are reported separately: '#' comments are included in this metric.
count shell_tools run.sh verify-source.sh count.sh
