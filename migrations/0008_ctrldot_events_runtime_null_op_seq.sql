-- Allow Ctrl Dot runtime to append events without a Kernel operation (op_seq nullable).
-- Existing rows keep op_seq; new runtime-only events use NULL.

BEGIN;

-- Drop FK so we can make op_seq nullable (runtime-only events have no operation)
ALTER TABLE ctrldot_events DROP CONSTRAINT IF EXISTS ctrldot_events_op_seq_fkey;

ALTER TABLE ctrldot_events ALTER COLUMN op_seq DROP NOT NULL;

COMMIT;
