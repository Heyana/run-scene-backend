-- 导出混元3D数据的SQL脚本
-- 使用方法：sqlite3 data/app.db < scripts/export_hunyuan_data.sql > scripts/hunyuan_data_export.sql

-- 开始事务
BEGIN TRANSACTION;

-- 导出表结构（如果表不存在则创建）
.output scripts/hunyuan_data_export.sql
.mode insert hunyuan_tasks

-- 导出数据
SELECT * FROM hunyuan_tasks;

COMMIT;
