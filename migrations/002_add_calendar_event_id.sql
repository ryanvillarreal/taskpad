-- Add calendar_event_id to track synced CalDAV events.
ALTER TABLE todos ADD COLUMN calendar_event_id TEXT NOT NULL DEFAULT '';
