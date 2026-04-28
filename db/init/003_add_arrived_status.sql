-- schedules.status に 'arrived' を追加
ALTER TABLE schedules
  DROP CONSTRAINT IF EXISTS schedules_status_check;

ALTER TABLE schedules
  ADD CONSTRAINT schedules_status_check
  CHECK (status IN ('open', 'full', 'departed', 'arrived'));
