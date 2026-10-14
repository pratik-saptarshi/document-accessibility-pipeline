# Defense in depth

For security or reliability work, identify the primary control and the
independent boundary checks that limit damage if it fails. Verify that each
control is actually on the path, has clear ownership, and is tested at the
boundary it protects. Prefer narrowly scoped, independently useful controls;
duplicated checks without a threat or failure model add complexity. Do not
weaken or bypass repository enforcement to make a check pass.
