# Concepts

Shared domain vocabulary for this project — entities, named processes, and status concepts with project-specific meaning. Seeded with core domain vocabulary, then accretes as ce-compound and ce-compound-refresh process learnings; direct edits are fine. Glossary only, not a spec or catch-all.

## Monitoring

### Check

A configured monitoring task that combines a Checker, Assertions, a Backoff policy, and selected Notifiers to track the health of a target over repeated runs.

A Check starts with an unknown outcome when enabled, becomes successful or failing after a run, and is disabled when monitoring stops. A run fails when evidence collection errors, an Assertion cannot be evaluated, or an Assertion is not satisfied; otherwise it succeeds. Enabling a Check resets its displayed status to unknown, but does not clear its accumulated Check statistics.

### Checker

The evidence-gathering mechanism selected by a Check's type, which produces a Check response or reports an execution error.

A Checker gathers evidence without deciding whether it satisfies the Check's Assertions. Evidence may come from an active probe or a snapshot of measurements received from a monitored host.

### Assertion

A configured condition that compares evidence in a Check response with a target value to determine whether the Check's requirements are satisfied.

Assertions may evaluate a Metric, Metadata, or raw response content interpreted as text or structured data. Every configured Assertion must be satisfied for a run to succeed.

## Evidence and history

### Check response

The evidence from one Checker execution, consisting of numeric Metrics, textual Metadata, and any raw response content available for Assertions.

### Metric

A named numeric measurement in a Check response that can be compared by an Assertion and displayed across a Check's recent Events.

### Metadata

Named textual attributes in a Check response that describe the observed outcome and can be evaluated by Assertions, distinct from numeric Metrics and raw response content.

### Event

The timestamped record of one Check run, associating its Check response with the Check's identity and any alert classification and failure messages.

Every completed run produces an Event, including successful runs without an alert. A failing run records a Red alert, and the first successful run following failure records a Green alert; recording either does not establish that a notification was delivered.

### Check statistics

The accumulated outcome counts, consecutive success and failure counts, alert-attempt count, and observation and state-transition times maintained for a Check, distinct from its individual Events.

A successful run resets the consecutive failure count and alert-attempt count; a failing run resets the consecutive success count. Cumulative success and failure totals survive these transitions. The alert-attempt count tracks notification processing, rather than confirmed deliveries.

## Scheduling

### Backoff

A Check's scheduling policy that determines the wait between runs from its base interval and consecutive failure count.

The policy may keep the interval constant or increase it as failures accumulate. Recovery restores the base interval, while an explicit trigger can end the wait before the next scheduled run.

## Alerts and delivery

### Red alert

The failure classification attached to an Event when a Check run fails, carrying the reasons the run was unsuccessful.

Notification preferences can defer delivery until the consecutive failure count reaches a configured threshold and can suppress further failure notifications during the same failure sequence. These preferences do not prevent failed Events from being recorded as Red alerts.

### Green alert

The recovery classification attached to the first successful Event after a Check was failing.

Recovery resets the Backoff interval and makes a recovery notification eligible even if the preceding Red alerts did not reach the failure-notification threshold.

### Notifier

A named alert-delivery destination that a Check selects to send messages about Red alerts and Green alerts.

A Notifier can serve multiple Checks, and a Check can select multiple Notifiers. Notification delivery can fail independently of the Check's outcome.
