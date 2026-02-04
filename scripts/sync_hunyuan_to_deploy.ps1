# 同步混元3D数据到部署数据库
# 用途：将本地开发数据库的混元3D数据同步到 deploy/data/app.db

Write-Host "=== 混元3D数据同步脚本 ===" -ForegroundColor Cyan
Write-Host ""

$sourceDB = "data/app.db"
$targetDB = "deploy/data/app.db"
$exportFile = "scripts/hunyuan_data_export.sql"

# 检查源数据库
if (-not (Test-Path $sourceDB)) {
    Write-Host "错误: 源数据库不存在: $sourceDB" -ForegroundColor Red
    exit 1
}

Write-Host "1. 检查源数据库..." -ForegroundColor Yellow
$count = sqlite3 $sourceDB "SELECT COUNT(*) FROM hunyuan_tasks;"
Write-Host "   源数据库有 $count 条混元3D任务记录" -ForegroundColor Green

# 检查目标数据库
if (-not (Test-Path $targetDB)) {
    Write-Host "警告: 目标数据库不存在，将创建新数据库" -ForegroundColor Yellow
    New-Item -ItemType File -Path $targetDB -Force | Out-Null
}

Write-Host ""
Write-Host "2. 导出混元3D数据..." -ForegroundColor Yellow

# 导出表结构和数据
$exportSQL = @"
-- 混元3D数据导出
-- 导出时间: $(Get-Date -Format "yyyy-MM-dd HH:mm:ss")

-- 创建表（如果不存在）
CREATE TABLE IF NOT EXISTS hunyuan_tasks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    created_at DATETIME,
    updated_at DATETIME,
    job_id TEXT,
    status TEXT,
    input_type TEXT,
    prompt TEXT,
    image_url TEXT,
    model TEXT DEFAULT '3.1',
    face_count INTEGER,
    generate_type TEXT DEFAULT 'Normal',
    enable_pbr NUMERIC DEFAULT true,
    result_format TEXT DEFAULT 'GLB',
    error_code TEXT,
    error_message TEXT,
    result_files TEXT,
    local_path TEXT,
    nas_path TEXT,
    thumbnail_path TEXT,
    file_size INTEGER,
    file_hash TEXT,
    name TEXT,
    description TEXT,
    tags TEXT,
    category TEXT DEFAULT 'AI生成',
    created_by TEXT,
    created_ip TEXT
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_hunyuan_tasks_status ON hunyuan_tasks(status);
CREATE UNIQUE INDEX IF NOT EXISTS idx_hunyuan_tasks_job_id ON hunyuan_tasks(job_id);

-- 导出数据
"@

# 导出数据为INSERT语句
$data = sqlite3 $sourceDB ".mode insert hunyuan_tasks" "SELECT * FROM hunyuan_tasks;"
$exportSQL += "`n$data"

# 保存到文件
$exportSQL | Out-File -FilePath $exportFile -Encoding UTF8
Write-Host "   数据已导出到: $exportFile" -ForegroundColor Green

Write-Host ""
Write-Host "3. 导入到目标数据库..." -ForegroundColor Yellow

# 检查目标数据库是否已有混元表
$hasTable = sqlite3 $targetDB "SELECT name FROM sqlite_master WHERE type='table' AND name='hunyuan_tasks';"
if ($hasTable) {
    Write-Host "   目标数据库已有混元表" -ForegroundColor Yellow
    $targetCount = sqlite3 $targetDB "SELECT COUNT(*) FROM hunyuan_tasks;"
    Write-Host "   目标数据库当前有 $targetCount 条记录" -ForegroundColor Yellow
    
    Write-Host ""
    $confirm = Read-Host "是否要清空目标数据库的混元数据并重新导入？(y/N)"
    if ($confirm -eq 'y' -or $confirm -eq 'Y') {
        Write-Host "   清空目标数据库的混元数据..." -ForegroundColor Yellow
        sqlite3 $targetDB "DELETE FROM hunyuan_tasks;"
    } else {
        Write-Host "   将追加数据到目标数据库..." -ForegroundColor Yellow
    }
}

# 导入数据
sqlite3 $targetDB < $exportFile

Write-Host "   数据导入完成" -ForegroundColor Green

# 验证
Write-Host ""
Write-Host "4. 验证导入结果..." -ForegroundColor Yellow
$finalCount = sqlite3 $targetDB "SELECT COUNT(*) FROM hunyuan_tasks;"
Write-Host "   目标数据库现在有 $finalCount 条混元3D任务记录" -ForegroundColor Green

Write-Host ""
Write-Host "=== 同步完成 ===" -ForegroundColor Cyan
Write-Host ""
Write-Host "下一步操作：" -ForegroundColor Yellow
Write-Host "1. 将 deploy/data/app.db 复制到服务器的 data/ 目录" -ForegroundColor White
Write-Host "2. 确保服务器上的 NAS 路径配置正确" -ForegroundColor White
Write-Host "3. 重启服务器上的后端服务" -ForegroundColor White
Write-Host ""
