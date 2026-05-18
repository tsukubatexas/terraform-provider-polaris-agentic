# ADR 0003: Start with Generic REST Terraform Primitives

Date: 2026-05-18

Status: Accepted

## Context

Apache Polaris has many REST operations. Generating a fully typed Terraform resource for every endpoint directly from OpenAPI would be risky because Terraform needs stable identities, lifecycle behavior, import semantics, and drift handling that OpenAPI does not fully describe.

## Decision

Start with two generic primitives:

- `polaris_rest_call` for read-style calls.
- `polaris_rest_resource` for create/read/update/delete workflows.

Both can use generated Polaris OpenAPI `operationId`s or explicit methods and paths.

## Consequences

- Every generated Polaris operation can be called without pretending all endpoints are first-class Terraform resources.
- Early users can test and automate Polaris functionality while the provider learns which endpoints deserve typed resources.
- Typed resources can be added later, but each new typed resource must include normal unit tests and real Polaris static infra coverage when practical.
- Request and response bodies remain sensitive because Polaris responses may include operational metadata or credentials.

