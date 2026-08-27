import { runQuint, withApalache } from './quint-runner.mjs';

await withApalache(({ serverEndpoint }) => {
  const result = runQuint([
    'verify',
    'model/CamFlow.qnt',
    '--main=CamFlow',
    '--invariant=allInvariants',
    '--max-steps=12',
    `--server-endpoint=${serverEndpoint}`,
    '--verbosity=1',
  ]);

  process.stdout.write(result.stdout || '[ok] Safe protocol verified.\n');
});
