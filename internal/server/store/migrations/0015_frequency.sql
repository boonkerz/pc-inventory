-- +goose Up
-- Ausführungs-Häufigkeit je Check/Task. '' = jeden Checkin (Check) bzw. Legacy-Plan (Task).
-- Werte: 1m,5m,15m,30m,1h,2h,6h,12h,daily,weekly,monthly,yearly
ALTER TABLE policy_checks ADD COLUMN frequency TEXT NOT NULL DEFAULT '';
ALTER TABLE policy_tasks ADD COLUMN frequency TEXT NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE policy_tasks DROP COLUMN frequency;
ALTER TABLE policy_checks DROP COLUMN frequency;
