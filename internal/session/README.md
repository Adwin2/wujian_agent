# Session Store Boundary

`internal/session` provides Redis-backed short-term conversation storage for the upcoming memory-management stage.

Current Phase 5 scope keeps this module intentionally unintegrated with the chat flow because profile memory requires explicit PHI rules before automatic reads/writes are safe.

Memory-management work should define:

- which conversation messages can be persisted;
- which structured health facts can become long-term profile memory;
- retention/TTL and user deletion behavior;
- how high-risk/HITL state is represented without claiming human review occurred;
- how Redis session history is injected into agent prompts without leaking PHI into traces.

Until that stage is implemented, `session_id` remains persisted in assessment records only; Redis is present in the local stack as a ready dependency, not an active memory backend.
