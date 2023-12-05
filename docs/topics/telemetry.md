<!-- markdownlint-disable MD033 -->

# KubeFox Federated Telemetry

When we first conceived of KubeFox, we thought of the things that should be
present in the framework as first class citizens, as opposed to tack on
afterthoughts. Telemetry falls into category. Wouldn’t it be nice to have
context-sensitive logs, metrics, traces and auditing data? Without going
through the travails of configuring them?

With KubeFox, these capabilities are intrinsic to the platform.

You can visualize the behavior of the entire System (all applications composing
the Retail System) as shown in Figure 1:

<figure markdown>
  <img src="../../diagrams/telemetry/telemetry_system.svg" width=100% height=100%>
  <figcaption>Figure 1 - Span-based telemetry for the entire System</figcaption>
</figure>

You can visualize behavior by Application (Figure 2) - for instance for the
Fulfillment app:

<figure markdown>
  <img src="../../diagrams/telemetry/telemetry_application.svg" width=100% height=100%>
  <figcaption>Figure 2 - Span-based telemetry for the Fulfillment App</figcaption>
</figure>

And you can visualize behavior by individual Component (Figure 3):

<figure markdown>
  <img src="../../diagrams/telemetry/telemetry_component.svg" width=100% height=100%>
  <figcaption>Figure 3 - Span-based telemetry for the Web UI component of the Web UI App</figcaption>
</figure>
