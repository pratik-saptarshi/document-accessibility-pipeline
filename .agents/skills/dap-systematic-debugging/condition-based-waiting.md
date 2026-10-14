# Condition-based waiting

When asynchronous work must finish before the next step, wait for its observable
completion condition with a bounded deadline and report timeout as incomplete.
Prefer the task runner's completion signal, a process exit, or a documented
status condition over arbitrary sleeps or repeated rapid polling. Do not start
duplicate work while the first operation's outcome is uncertain; inspect its
status first. Keep polling frequency and output bounded.
