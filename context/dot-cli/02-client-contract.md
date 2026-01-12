# Kernel Client Contract (Dot CLI) v0.1

Dot CLI calls these endpoints:
- GET /v1/healthz
- POST /v1/plan
- POST /v1/apply
- GET /v1/expand
- GET /v1/history
- GET /v1/diff

## Error handling
Map kernel error codes to exit codes:
- POLICY_DENIED => 2
- CONFLICT => 3
- VALIDATION => 1
- NOT_FOUND => 1
- INTERNAL/others => 4
