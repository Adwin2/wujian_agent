CREATE EXTENSION IF NOT EXISTS "pgcrypto";
CREATE EXTENSION IF NOT EXISTS "vector";

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    age INT NOT NULL CHECK (age BETWEEN 2 AND 20),
    sex TEXT NOT NULL CHECK (sex IN ('male','female')),
    guardian_id UUID REFERENCES users(id),
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE IF NOT EXISTS health_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    height_cm NUMERIC(5,1) CHECK (height_cm > 0 AND height_cm <= 250),
    weight_kg NUMERIC(5,1) CHECK (weight_kg > 0 AND weight_kg <= 300),
    bmi NUMERIC(4,1),
    bmi_percentile NUMERIC(5,2),
    notes TEXT,
    recorded_at TIMESTAMPTZ DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_health_records_user ON health_records(user_id, recorded_at DESC);

CREATE TABLE IF NOT EXISTS assessments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id),
    session_id TEXT NOT NULL,
    input_text TEXT NOT NULL,
    output_text TEXT,
    agents_called TEXT[],
    tools_called JSONB DEFAULT '[]',
    trace JSONB,
    risk_flags JSONB DEFAULT '[]',
    hitl_triggered BOOLEAN DEFAULT false,
    hitl_resolved_by UUID REFERENCES users(id),
    hitl_resolved_at TIMESTAMPTZ,
    tokens_used INT DEFAULT 0,
    latency_ms INT DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_assessments_session ON assessments(session_id);
CREATE INDEX IF NOT EXISTS idx_assessments_user ON assessments(user_id, created_at DESC);

CREATE TABLE IF NOT EXISTS audit_log (
    id BIGSERIAL PRIMARY KEY,
    user_id UUID,
    action TEXT NOT NULL,
    resource_type TEXT NOT NULL,
    resource_id UUID,
    tool_name TEXT,
    tool_input JSONB,
    tool_output JSONB,
    ip_address INET,
    created_at TIMESTAMPTZ DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_audit_log_user ON audit_log(user_id, created_at DESC);
