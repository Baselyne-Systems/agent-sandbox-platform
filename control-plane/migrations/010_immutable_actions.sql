-- 010: Enforce immutability on action_records via triggers.
-- Prevents UPDATE and DELETE at the database level.

CREATE OR REPLACE FUNCTION prevent_action_record_modification()
RETURNS TRIGGER AS $$
BEGIN
    RAISE EXCEPTION 'action_records table is immutable: % operations are not allowed', TG_OP;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER action_records_no_update
    BEFORE UPDATE ON action_records
    FOR EACH ROW EXECUTE FUNCTION prevent_action_record_modification();

CREATE TRIGGER action_records_no_delete
    BEFORE DELETE ON action_records
    FOR EACH ROW EXECUTE FUNCTION prevent_action_record_modification();
