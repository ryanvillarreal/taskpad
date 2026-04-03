ALTER TABLE todos ADD COLUMN status TEXT NOT NULL DEFAULT 'active';
ALTER TABLE todos ADD COLUMN urgency TEXT NOT NULL DEFAULT 'normal';
ALTER TABLE todos ADD COLUMN tags TEXT NOT NULL DEFAULT '[]';

UPDATE todos
SET status = CASE
    WHEN completed = 1 THEN 'done'
    ELSE 'active'
END;

UPDATE todos
SET urgency = CASE priority
    WHEN 'high' THEN 'high'
    WHEN 'low' THEN 'low'
    ELSE 'normal'
END;

CREATE INDEX IF NOT EXISTS idx_todos_status ON todos(status);
CREATE INDEX IF NOT EXISTS idx_todos_urgency ON todos(urgency);
