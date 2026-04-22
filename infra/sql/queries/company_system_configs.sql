-- name: InsertCompanySystemConfig :execrows
INSERT INTO company_system_configs(company_id, schedule_init_time, schedule_pause_init_time, schedule_pause_end_time, schedule_end_time, min_schedules_per_day, max_schedules_per_day, schedule_days, dynamic_cages, total_small_cages, total_medium_cages, total_large_cages, total_giant_cages, whatsapp_notifications, whatsapp_conversation, whatsapp_business_phone)
    VALUES (sqlc.arg('CompanyID'), sqlc.arg('ScheduleInitTime'), sqlc.arg('SchedulePauseInitTime'), sqlc.arg('SchedulePauseEndTime'), sqlc.arg('ScheduleEndTime'), sqlc.arg('MinSchedulesPerDay'), sqlc.arg('MaxSchedulesPerDay'), sqlc.arg('ScheduleDays'), sqlc.arg('DynamicCages'), sqlc.arg('TotalSmallCages'), sqlc.arg('TotalMediumCages'), sqlc.arg('TotalLargeCages'), sqlc.arg('TotalGiantCages'), sqlc.arg('WhatsappNotifications'), sqlc.arg('WhatsappConversation'), sqlc.arg('WhatsappBusinessPhone'));

-- name: UpdateCompanySystemConfig :execrows
UPDATE
    company_system_configs
SET
    schedule_init_time = sqlc.arg('ScheduleInitTime'),
    schedule_pause_init_time = sqlc.arg('SchedulePauseInitTime'),
    schedule_pause_end_time = sqlc.arg('SchedulePauseEndTime'),
    schedule_end_time = sqlc.arg('ScheduleEndTime'),
    min_schedules_per_day = sqlc.arg('MinSchedulesPerDay'),
    max_schedules_per_day = sqlc.arg('MaxSchedulesPerDay'),
    schedule_days = sqlc.arg('ScheduleDays'),
    dynamic_cages = sqlc.arg('DynamicCages'),
    total_small_cages = sqlc.arg('TotalSmallCages'),
    total_medium_cages = sqlc.arg('TotalMediumCages'),
    total_large_cages = sqlc.arg('TotalLargeCages'),
    total_giant_cages = sqlc.arg('TotalGiantCages'),
    whatsapp_notifications = sqlc.arg('WhatsappNotifications'),
    whatsapp_conversation = sqlc.arg('WhatsappConversation'),
    whatsapp_business_phone = sqlc.arg('WhatsappBusinessPhone'),
    updated_at = NOW()
WHERE
    company_id = sqlc.arg('CompanyID');

-- name: GetCompanySystemConfig :one
SELECT
    csc.company_id,
    csc.schedule_init_time,
    csc.schedule_pause_init_time,
    csc.schedule_pause_end_time,
    csc.schedule_end_time,
    csc.min_schedules_per_day,
    csc.max_schedules_per_day,
    csc.schedule_days,
    csc.dynamic_cages,
    csc.total_small_cages,
    csc.total_medium_cages,
    csc.total_large_cages,
    csc.total_giant_cages,
    csc.whatsapp_notifications,
    csc.whatsapp_conversation,
    csc.whatsapp_business_phone,
    csc.created_at,
    csc.updated_at
FROM
    company_system_configs csc
WHERE
    csc.company_id = sqlc.arg('CompanyID');
