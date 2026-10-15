# Root cause tracing

Start at the observed failure and walk backward through callers, inputs, and
state transitions until reaching the earliest incorrect assumption or mutation.
At each boundary ask: what entered, what invariant should hold, and what
evidence shows the first divergence? Check the current revision and adjacent
tests before concluding that code or protection is absent. Prefer a small
reproduction with synthetic data; avoid broad logging or disclosing sensitive
values. Fix the cause at its owning boundary and record any remaining unknowns.
