-- 修复混元3D任务的路径格式
-- 此脚本用于修复已有任务的 nasPath 和 thumbnailPath 字段
-- 使其能够正确生成 fileUrl 和 thumbnailUrl

-- 查看当前的路径格式
SELECT 
    id,
    job_id,
    status,
    nas_path,
    thumbnail_path
FROM hunyuan_tasks
WHERE status = 'DONE'
ORDER BY id DESC
LIMIT 10;

-- 如果路径格式正确，应该看到类似：
-- \\192.168.3.10\project\editor_v2\static\hunyuan\2026\02\file.glb

-- 如果需要修复路径（通常不需要，因为 AfterFind 钩子会自动处理）
-- 但如果你想在数据库中也存储正确的相对路径，可以执行以下更新：

-- 备份表（可选，建议先备份）
-- CREATE TABLE hunyuan_tasks_backup AS SELECT * FROM hunyuan_tasks;

-- 注意：通常不需要修改数据库中的路径
-- AfterFind 钩子会在查询时自动生成正确的 URL
-- 以下SQL仅供参考，不建议执行

-- 示例：如果要将绝对路径转换为相对路径（不推荐）
-- UPDATE hunyuan_tasks
-- SET nas_path = REPLACE(
--     REPLACE(nas_path, '\\192.168.3.10\project\editor_v2\static\hunyuan\', ''),
--     '\', '/'
-- )
-- WHERE nas_path LIKE '\\192.168.3.10%';

-- 查看修复后的结果
SELECT 
    id,
    job_id,
    status,
    nas_path,
    thumbnail_path,
    file_size,
    created_at
FROM hunyuan_tasks
WHERE status = 'DONE'
ORDER BY id DESC
LIMIT 10;
