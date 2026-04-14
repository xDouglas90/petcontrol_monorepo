-- name: InsertCompanySystemConfig :execrows
INSERT INTO company_system_configs(company_id, schedule_init_time, schedule_pause_init_time, schedule_pause_end_time, schedule_end_time, min_schedules_per_day, max_schedules_per_day, schedule_days, dynamic_cages, total_small_cages, total_medium_cages, total_large_cages, total_giant_cages, whatsapp_notifications, whatsapp_conversation, whatsapp_business_phone)
    VALUES (sqlc.arg('CompanyID'), sqlc.arg('ScheduleInitTime'), sqlc.arg('SchedulePauseInitTime'), sqlc.arg('SchedulePauseEndTime'), sqlc.arg('ScheduleEndTime'), sqlc.arg('MinSchedulesPerDay'), sqlc.arg('MaxSchedulesPerDay'), sqlc.arg('ScheduleDays'), sqlc.arg('DynamicCages'), sqlc.arg('TotalSmallCages'), sqlc.arg('TotalMediumCages'), sqlc.arg('TotalLargeCages'), sqlc.arg('TotalGiantCages'), sqlc.arg('WhatsappNotifications'), sqlc.arg('WhatsappConversation'), sqlc.arg('WhatsappBusinessPhone'));

-- name: UpdateCompanySystemConfig :execrows
UPDATE
    company_system_configs
SET
    schedule_init_time = COALESCE(sqlc.narg('ScheduleInitTime'), schedule_init_time),
    schedule_pause_init_time = COALESCE(sqlc.narg('SchedulePauseInitTime'), schedule_pause_init_time),
    schedule_pause_end_time = COALESCE(sqlc.narg('SchedulePauseEndTime'), schedule_pause_end_time),
    schedule_end_time = COALESCE(sqlc.narg('ScheduleEndTime'), schedule_end_time),
    min_schedules_per_day = COALESCE(sqlc.narg('MinSchedulesPerDay'), min_schedules_per_day),
    max_schedules_per_day = COALESCE(sqlc.narg('MaxSchedulesPerDay'), max_schedules_per_day),
    schedule_days = COALESCE(sqlc.narg('ScheduleDays'), schedule_days),
    dynamic_cages = COALESCE(sqlc.narg('DynamicCages'), dynamic_cages),
    total_small_cages = COALESCE(sqlc.narg('TotalSmallCages'), total_small_cages),
    total_medium_cages = COALESCE(sqlc.narg('TotalMediumCages'), total_medium_cages),
    total_large_cages = COALESCE(sqlc.narg('TotalLargeCages'), total_large_cages),
    total_giant_cages = COALESCE(sqlc.narg('TotalGiantCages'), total_giant_cages),
    whatsapp_notifications = COALESCE(sqlc.narg('WhatsappNotifications'), whatsapp_notifications),
    whatsapp_conversation = COALESCE(sqlc.narg('WhatsappConversation'), whatsapp_conversation),
    whatsapp_business_phone = COALESCE(sqlc.narg('WhatsappBusinessPhone'), whatsapp_business_phone)
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

