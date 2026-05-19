import React from "react";

export type WriteLifecyclePhase = "idle" | "prepare" | "lease" | "validate" | "write" | "ack" | "done" | "error";

export type WriteLifecycleTraceProps = {
  phase: WriteLifecyclePhase;
  status?: string;
  unit?: string;
  paramId?: string | number;
  deviceId?: string;
  instance?: number;
  elapsedMs?: number;
  leaseHolder?: string;
  holderId?: string;
  busy?: boolean;
  staged?: string;
  dangerous?: boolean;
  commandName?: string;
  // Trace from backend if available
  trace?: {
    phase?: WriteLifecyclePhase;
    status?: string;
    unit?: string;
    paramId?: string | number;
    at?: number;
    confirmedMatched?: boolean;
    prevValue?: number | string;
    prospectiveValue?: number | string;
    submittedValue?: number | string;
    confirmedValue?: number | string;
    commandName?: string;
    error?: string;
  } | null;
  formatValue?: (val: any, unit: string, id?: any) => string;
};

export function WriteLifecycleTrace({
  phase = "idle",
  status,
  unit = "",
  paramId,
  deviceId,
  instance,
  elapsedMs = 0,
  leaseHolder,
  holderId,
  busy = false,
  staged = "",
  dangerous = false,
  commandName,
  trace = null,
  formatValue = (v) => String(v),
}: WriteLifecycleTraceProps) {
  const youHold = leaseHolder === holderId;
  const tracePhase = trace?.phase ?? phase;
  const traceStatus = trace?.status ?? (tracePhase === "done" ? "completed" : tracePhase === "error" ? "failed" : tracePhase);
  const activeUnit = trace?.unit ?? unit;
  const activeParamId = trace?.paramId ?? paramId;
  const phaseElapsed = trace?.at ? Date.now() - trace.at : elapsedMs;
  const confirmationMatched = trace?.confirmedMatched === true;
  const confirmationMismatched = trace?.confirmedMatched === false;
  const confirmationState = confirmationMatched ? "confirmedMatched" : (confirmationMismatched ? "readback mismatch" : traceStatus);

  const valueRows = trace ? [
    { label: "previous", value: formatValue(trace.prevValue, activeUnit, activeParamId) },
    { label: "prospective", value: formatValue(trace.prospectiveValue, activeUnit, activeParamId) },
    { label: "submitted", value: formatValue(trace.submittedValue, activeUnit, activeParamId) },
    { label: "confirmed", value: trace.confirmedValue !== undefined ? `${formatValue(trace.confirmedValue, activeUnit, activeParamId)} · ${confirmationState}` : confirmationState },
  ] : [];

  const steps = [
    { key: "prepare", label: "prepare", detail: staged ? `staged ${staged}` : "no staged value" },
    { key: "lease", label: "lease", detail: youHold ? "held locally" : (leaseHolder ? `held by ${leaseHolder}` : "available") },
    { key: "validate", label: "validate", detail: dangerous ? "confirmation required" : "range / type check" },
    { key: "write", label: "write", detail: commandName || trace?.commandName || "write" },
    { key: "ack", label: "ack", detail: tracePhase === "done" ? confirmationState : (tracePhase === "error" ? (trace?.error || "failed") : (busy ? "waiting" : "idle")) },
  ];

  return (
    <div className="sf-write-lifecycle-trace" data-phase={tracePhase} data-confirmation={confirmationMatched ? "matched" : (confirmationMismatched ? "mismatch" : "unknown")} title={`device=${deviceId} instance=${instance || 1}`}>
      <div className="sf-write-lifecycle-trace__head">
        <span className="sf-write-lifecycle-trace__title">Write lifecycle</span>
        <span className="sf-write-lifecycle-trace__meta">{busy ? "busy" : traceStatus} · {Math.max(0, Math.round(phaseElapsed))} ms</span>
      </div>
      <div className="sf-write-lifecycle-trace__steps">
        {steps.map((step) => (
          <span key={step.key} className={"sf-write-lifecycle-trace__step " + (tracePhase === step.key || (tracePhase === "done" && step.key === "ack") || (tracePhase === "error" && step.key === "ack") ? "on" : "")}>
            <b>{step.label}</b>
            <em>{step.detail}</em>
          </span>
        ))}
      </div>
      {valueRows.length > 0 && (
        <div className="sf-write-lifecycle-trace__values">
          {valueRows.map((row) => (
            <span key={row.label} className="sf-write-lifecycle-trace__value-row">
              <b>{row.label}</b>
              <em>{row.value}</em>
            </span>
          ))}
        </div>
      )}
    </div>
  );
}
