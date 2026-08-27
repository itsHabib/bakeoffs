// CamFlow.qnt
//
// A deliberately small model of a CAM operator client handing a completed
// programming payload to a factory workflow service. The names are generic;
// the protocol shape is derived from real production failures involving local
// task state, a remote proveout revision, retries, and durable command custody.

module CamFlow {
  type TaskStatus =
    | NotStarted
    | Started
    | CheckedIn
    | CheckedOut
    | TaskComplete
    | TaskInvalidated

  type EpochStatus = EpochOpen | EpochClosed | EpochInvalidated
  type FlowStatus = NoPayload | WriteOpen | PayloadSaved | FlowComplete
  type CommandStatus = NoCommand | CommandRunning | CommandRetained | CommandComplete

  var taskStatus: TaskStatus
  var epochStatus: EpochStatus
  var flowStatus: FlowStatus
  var commandStatus: CommandStatus
  var snapshotReady: bool
  var receiptRecorded: bool
  var recoveryAvailable: bool
  var flowSaveCount: int

  action init = all {
    taskStatus' = NotStarted,
    epochStatus' = EpochOpen,
    flowStatus' = NoPayload,
    commandStatus' = NoCommand,
    snapshotReady' = false,
    receiptRecorded' = false,
    recoveryAvailable' = false,
    flowSaveCount' = 0,
  }

  // Task ownership is historical: after the first start, a checked-out task
  // resumes as CheckedIn rather than creating a second Started interval.
  action startFirst = all {
    taskStatus == NotStarted,
    taskStatus' = Started,
    epochStatus' = epochStatus,
    flowStatus' = flowStatus,
    commandStatus' = commandStatus,
    snapshotReady' = snapshotReady,
    receiptRecorded' = receiptRecorded,
    recoveryAvailable' = recoveryAvailable,
    flowSaveCount' = flowSaveCount,
  }

  action checkIn = all {
    taskStatus == Started,
    taskStatus' = CheckedIn,
    epochStatus' = epochStatus,
    flowStatus' = flowStatus,
    commandStatus' = commandStatus,
    snapshotReady' = snapshotReady,
    receiptRecorded' = receiptRecorded,
    recoveryAvailable' = recoveryAvailable,
    flowSaveCount' = flowSaveCount,
  }

  action checkOut = all {
    taskStatus == Started or taskStatus == CheckedIn,
    taskStatus' = CheckedOut,
    epochStatus' = epochStatus,
    flowStatus' = flowStatus,
    commandStatus' = commandStatus,
    snapshotReady' = snapshotReady,
    receiptRecorded' = receiptRecorded,
    recoveryAvailable' = recoveryAvailable,
    flowSaveCount' = flowSaveCount,
  }

  action resume = all {
    taskStatus == CheckedOut,
    taskStatus' = CheckedIn,
    epochStatus' = epochStatus,
    flowStatus' = flowStatus,
    commandStatus' = commandStatus,
    snapshotReady' = snapshotReady,
    receiptRecorded' = receiptRecorded,
    recoveryAvailable' = recoveryAvailable,
    flowSaveCount' = flowSaveCount,
  }

  action stageSnapshot = all {
    taskStatus == Started or taskStatus == CheckedIn,
    not(snapshotReady),
    taskStatus' = taskStatus,
    epochStatus' = epochStatus,
    flowStatus' = flowStatus,
    commandStatus' = commandStatus,
    snapshotReady' = true,
    receiptRecorded' = receiptRecorded,
    recoveryAvailable' = recoveryAvailable,
    flowSaveCount' = flowSaveCount,
  }

  action openFlowWrite = all {
    epochStatus == EpochOpen,
    taskStatus == Started or taskStatus == CheckedIn,
    snapshotReady,
    flowStatus == NoPayload,
    taskStatus' = taskStatus,
    epochStatus' = epochStatus,
    flowStatus' = WriteOpen,
    commandStatus' = commandStatus,
    snapshotReady' = snapshotReady,
    receiptRecorded' = receiptRecorded,
    recoveryAvailable' = recoveryAvailable,
    flowSaveCount' = flowSaveCount,
  }

  action saveFlowPayload = all {
    epochStatus == EpochOpen,
    taskStatus == Started or taskStatus == CheckedIn,
    flowStatus == WriteOpen,
    taskStatus' = taskStatus,
    epochStatus' = epochStatus,
    flowStatus' = PayloadSaved,
    commandStatus' = commandStatus,
    snapshotReady' = snapshotReady,
    receiptRecorded' = receiptRecorded,
    recoveryAvailable' = recoveryAvailable,
    flowSaveCount' = flowSaveCount + 1,
  }

  // The local receipt is deliberately separate from the remote save. A crash
  // can occur between them; a retry must reconcile instead of saving again.
  action recordUploadReceipt = all {
    flowStatus == PayloadSaved,
    not(receiptRecorded),
    taskStatus' = taskStatus,
    epochStatus' = epochStatus,
    flowStatus' = flowStatus,
    commandStatus' = commandStatus,
    snapshotReady' = snapshotReady,
    receiptRecorded' = true,
    recoveryAvailable' = recoveryAvailable,
    flowSaveCount' = flowSaveCount,
  }

  action completeTask = all {
    taskStatus == Started or taskStatus == CheckedIn,
    receiptRecorded,
    taskStatus' = TaskComplete,
    epochStatus' = epochStatus,
    flowStatus' = flowStatus,
    commandStatus' = commandStatus,
    snapshotReady' = snapshotReady,
    receiptRecorded' = receiptRecorded,
    recoveryAvailable' = recoveryAvailable,
    flowSaveCount' = flowSaveCount,
  }

  action beginCompleteCommand = all {
    taskStatus == TaskComplete,
    receiptRecorded,
    flowStatus == PayloadSaved,
    commandStatus == NoCommand,
    taskStatus' = taskStatus,
    epochStatus' = epochStatus,
    flowStatus' = flowStatus,
    commandStatus' = CommandRunning,
    snapshotReady' = snapshotReady,
    receiptRecorded' = receiptRecorded,
    recoveryAvailable' = false,
    flowSaveCount' = flowSaveCount,
  }

  // Retention is safe only when the same invocation remains resumable or an
  // explicit operator recovery path exists.
  action retainCommand = all {
    commandStatus == CommandRunning,
    taskStatus' = taskStatus,
    epochStatus' = epochStatus,
    flowStatus' = flowStatus,
    commandStatus' = CommandRetained,
    snapshotReady' = snapshotReady,
    receiptRecorded' = receiptRecorded,
    recoveryAvailable' = true,
    flowSaveCount' = flowSaveCount,
  }

  action resumeRetainedCommand = all {
    commandStatus == CommandRetained,
    recoveryAvailable,
    taskStatus' = taskStatus,
    epochStatus' = epochStatus,
    flowStatus' = flowStatus,
    commandStatus' = CommandRunning,
    snapshotReady' = snapshotReady,
    receiptRecorded' = receiptRecorded,
    recoveryAvailable' = false,
    flowSaveCount' = flowSaveCount,
  }

  action finishCompleteCommand = all {
    commandStatus == CommandRunning,
    taskStatus' = taskStatus,
    epochStatus' = epochStatus,
    flowStatus' = FlowComplete,
    commandStatus' = CommandComplete,
    snapshotReady' = snapshotReady,
    receiptRecorded' = receiptRecorded,
    recoveryAvailable' = false,
    flowSaveCount' = flowSaveCount,
  }

  action closeEpoch = all {
    epochStatus == EpochOpen,
    taskStatus == TaskComplete,
    flowStatus == FlowComplete,
    commandStatus == CommandComplete,
    receiptRecorded,
    taskStatus' = taskStatus,
    epochStatus' = EpochClosed,
    flowStatus' = flowStatus,
    commandStatus' = commandStatus,
    snapshotReady' = snapshotReady,
    receiptRecorded' = receiptRecorded,
    recoveryAvailable' = recoveryAvailable,
    flowSaveCount' = flowSaveCount,
  }

  action invalidateEpoch = all {
    epochStatus == EpochOpen,
    taskStatus != TaskComplete,
    flowStatus == NoPayload,
    taskStatus' = TaskInvalidated,
    epochStatus' = EpochInvalidated,
    flowStatus' = flowStatus,
    commandStatus' = commandStatus,
    snapshotReady' = snapshotReady,
    receiptRecorded' = receiptRecorded,
    recoveryAvailable' = recoveryAvailable,
    flowSaveCount' = flowSaveCount,
  }

  action idle = all {
    epochStatus == EpochClosed or epochStatus == EpochInvalidated,
    taskStatus' = taskStatus,
    epochStatus' = epochStatus,
    flowStatus' = flowStatus,
    commandStatus' = commandStatus,
    snapshotReady' = snapshotReady,
    receiptRecorded' = receiptRecorded,
    recoveryAvailable' = recoveryAvailable,
    flowSaveCount' = flowSaveCount,
  }

  action step = any {
    startFirst,
    checkIn,
    checkOut,
    resume,
    stageSnapshot,
    openFlowWrite,
    saveFlowPayload,
    recordUploadReceipt,
    completeTask,
    beginCompleteCommand,
    retainCommand,
    resumeRetainedCommand,
    finishCompleteCommand,
    closeEpoch,
    invalidateEpoch,
    idle,
  }

  val receiptReflectsRemoteReality = receiptRecorded implies
    (flowStatus == PayloadSaved or flowStatus == FlowComplete)

  val localCompletionHasReceipt =
    (taskStatus == TaskComplete) implies receiptRecorded

  val closedEpochIsFullyComplete = (epochStatus == EpochClosed) implies all {
    taskStatus == TaskComplete,
    flowStatus == FlowComplete,
    commandStatus == CommandComplete,
    receiptRecorded,
  }

  val uploadIsDeduplicated = flowSaveCount <= 1

  val retainedCommandIsRecoverable =
    (commandStatus == CommandRetained) implies recoveryAvailable

  val recoveryFlagIsScoped =
    recoveryAvailable implies commandStatus == CommandRetained

  val remoteWorkIsNotOrphaned =
    (flowStatus != NoPayload) implies epochStatus != EpochInvalidated

  val allInvariants = all {
    receiptReflectsRemoteReality,
    localCompletionHasReceipt,
    closedEpochIsFullyComplete,
    uploadIsDeduplicated,
    retainedCommandIsRecoverable,
    recoveryFlagIsScoped,
    remoteWorkIsNotOrphaned,
  }
}
