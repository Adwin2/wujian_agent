package repository

import (
	"context"
	"encoding/json"
	"fmt"

	appmodel "github.com/adwin2/youthvital/internal/model"
)

// SaveAssessment persists a completed chat turn, including tool trace and HITL state.
func (db *DB) SaveAssessment(ctx context.Context, record appmodel.AssessmentRecord) error {
	if db == nil || db.pool == nil {
		return nil
	}
	if record.SessionID == "" {
		record.SessionID = "default"
	}

	toolsCalled, err := json.Marshal(record.ToolCalls)
	if err != nil {
		return fmt.Errorf("marshal tool calls: %w", err)
	}
	riskFlags, err := json.Marshal(record.RiskFlags)
	if err != nil {
		return fmt.Errorf("marshal risk flags: %w", err)
	}

	_, err = db.pool.Exec(ctx, `
		INSERT INTO assessments (
			user_id, session_id, input_text, output_text, agents_called,
			tools_called, risk_flags, hitl_triggered
		)
		VALUES (NULLIF($1, '')::uuid, $2, $3, $4, $5, $6::jsonb, $7::jsonb, $8)
	`, record.UserID, record.SessionID, record.InputText, record.OutputText,
		record.AgentsCalled, string(toolsCalled), string(riskFlags), record.HITLTriggered)
	if err != nil {
		return fmt.Errorf("insert assessment: %w", err)
	}
	return nil
}
