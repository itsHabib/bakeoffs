// KnownFailures.qnt
//
// Each module adds one intentionally unsafe transition to the safe protocol.
// Verification is expected to fail and produce a minimal counterexample trace.

module DuplicateSaveBug {
  import CamFlow.* from "./CamFlow"

  // Remote save succeeded, but the client crashed before recording its local
  // receipt. A blind retry repeats the external effect.
  action retryWithoutReconciliation = all {
    flowStatus == PayloadSaved,
    not(receiptRecorded),
    taskStatus' = taskStatus,
    epochStatus' = epochStatus,
    flowStatus' = PayloadSaved,
    commandStatus' = commandStatus,
    snapshotReady' = snapshotReady,
    receiptRecorded' = receiptRecorded,
    recoveryAvailable' = recoveryAvailable,
    flowSaveCount' = flowSaveCount + 1,
  }

  action stepBug = any { step, retryWithoutReconciliation }
}

module RetainedWorkflowBug {
  import CamFlow.* from "./CamFlow"

  // A durable command is retained after its first attempt, but neither retry
  // nor operator recovery can resume or release it.
  action retainWithoutRecovery = all {
    commandStatus == CommandRunning,
    taskStatus' = taskStatus,
    epochStatus' = epochStatus,
    flowStatus' = flowStatus,
    commandStatus' = CommandRetained,
    snapshotReady' = snapshotReady,
    receiptRecorded' = receiptRecorded,
    recoveryAvailable' = false,
    flowSaveCount' = flowSaveCount,
  }

  action stepBug = any { step, retainWithoutRecovery }
}

module EarlyLocalCompletionBug {
  import CamFlow.* from "./CamFlow"

  // The local task is marked complete after a remote save response but before
  // the upload receipt is durably recorded.
  action completeBeforeReceipt = all {
    flowStatus == PayloadSaved,
    not(receiptRecorded),
    taskStatus == Started or taskStatus == CheckedIn,
    taskStatus' = TaskComplete,
    epochStatus' = epochStatus,
    flowStatus' = flowStatus,
    commandStatus' = commandStatus,
    snapshotReady' = snapshotReady,
    receiptRecorded' = receiptRecorded,
    recoveryAvailable' = recoveryAvailable,
    flowSaveCount' = flowSaveCount,
  }

  action stepBug = any { step, completeBeforeReceipt }
}

module InvalidateAfterRemoteSaveBug {
  import CamFlow.* from "./CamFlow"

  // Invalidation terminates the local epoch after a remote payload exists,
  // leaving that effect without a live local owner or compensation path.
  action invalidateAfterRemoteSave = all {
    epochStatus == EpochOpen,
    flowStatus == PayloadSaved,
    taskStatus != TaskComplete,
    taskStatus' = TaskInvalidated,
    epochStatus' = EpochInvalidated,
    flowStatus' = flowStatus,
    commandStatus' = commandStatus,
    snapshotReady' = snapshotReady,
    receiptRecorded' = receiptRecorded,
    recoveryAvailable' = recoveryAvailable,
    flowSaveCount' = flowSaveCount,
  }

  action stepBug = any { step, invalidateAfterRemoteSave }
}
