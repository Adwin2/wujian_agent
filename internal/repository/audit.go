package repository

import (
	"context"
	"encoding/json"
	"fmt"

	appmodel "github.com/adwin2/youthvital/internal/model"
)

// SaveAuditLog persists a PHI access audit record.
func (db *DB) SaveAuditLog(ctx context.Context, record appmodel.AuditLogRecord) error {
	if db == nil || db.pool == nil {
		return nil
	}
	toolInput, err := json.Marshal(record.ToolInput)
	if err != nil {
		return fmt.Errorf("marshal audit tool input: %w", err)
	}
	toolOutput, err := json.Marshal(record.ToolOutput)
	if err != nil {
		return fmt.Errorf("marshal audit tool output: %w", err)
	}

	_, err = db.pool.Exec(ctx, `
		INSERT INTO audit_log (
			user_id, action, resource_type, resource_id, tool_name, tool_input, tool_output
		)
		VALUES (NULLIF($1, '')::uuid, $2, $3, NULLIF($4, '')::uuid, $5, $6::jsonb, $7::jsonb)
	`, record.UserID, record.Action, record.ResourceType, record.ResourceID, record.ToolName, string(toolInput), string(toolOutput))
	if err != nil {
		return fmt.Errorf("insert audit log: %w", err)
	}
	return nil
}
