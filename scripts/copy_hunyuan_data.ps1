# Quick copy Hunyuan 3D data
# Copy data from local database to deploy database

Write-Host "=== Copy Hunyuan 3D Data ===" -ForegroundColor Cyan
Write-Host ""

$sourceDB = "data/app.db"
$targetDB = "deploy/data/app.db"

# Check source database
if (-not (Test-Path $sourceDB)) {
    Write-Host "Error: Source database not found: $sourceDB" -ForegroundColor Red
    exit 1
}

# Ensure target directory exists
$targetDir = Split-Path $targetDB -Parent
if (-not (Test-Path $targetDir)) {
    Write-Host "Creating target directory: $targetDir" -ForegroundColor Yellow
    New-Item -ItemType Directory -Path $targetDir -Force | Out-Null
}

# Check target database
if (-not (Test-Path $targetDB)) {
    Write-Host "Target database not found, creating new database" -ForegroundColor Yellow
    New-Item -ItemType File -Path $targetDB -Force | Out-Null
}

Write-Host "Source database: $sourceDB" -ForegroundColor White
Write-Host "Target database: $targetDB" -ForegroundColor White
Write-Host ""

# Count source data
$sourceCount = sqlite3 $sourceDB "SELECT COUNT(*) FROM hunyuan_tasks;"
Write-Host "Source database has $sourceCount Hunyuan 3D records" -ForegroundColor Green

# Check target database
$hasTable = sqlite3 $targetDB "SELECT name FROM sqlite_master WHERE type='table' AND name='hunyuan_tasks';" 2>$null
if ($hasTable) {
    $targetCount = sqlite3 $targetDB "SELECT COUNT(*) FROM hunyuan_tasks;"
    Write-Host "Target database currently has $targetCount records" -ForegroundColor Yellow
    Write-Host ""
    $confirm = Read-Host "Overwrite target database Hunyuan data? (y/N)"
    if ($confirm -ne 'y' -and $confirm -ne 'Y') {
        Write-Host "Operation cancelled" -ForegroundColor Yellow
        exit 0
    }
}

Write-Host ""
Write-Host "Copying data..." -ForegroundColor Yellow

# Use SQLite command to copy data
$copySQL = @"
ATTACH DATABASE '$sourceDB' AS source;

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
    category TEXT DEFAULT 'AI',
    created_by TEXT,
    created_ip TEXT
);

CREATE INDEX IF NOT EXISTS idx_hunyuan_tasks_status ON hunyuan_tasks(status);
CREATE UNIQUE INDEX IF NOT EXISTS idx_hunyuan_tasks_job_id ON hunyuan_tasks(job_id);

DELETE FROM hunyuan_tasks;

INSERT INTO hunyuan_tasks 
SELECT * FROM source.hunyuan_tasks;

DETACH DATABASE source;
"@

# Execute copy
$copySQL | sqlite3 $targetDB

# Verify result
$finalCount = sqlite3 $targetDB "SELECT COUNT(*) FROM hunyuan_tasks;"
Write-Host ""
Write-Host "Copy completed!" -ForegroundColor Green
Write-Host "Target database now has $finalCount Hunyuan 3D records" -ForegroundColor Green

Write-Host ""
Write-Host "=== Done ===" -ForegroundColor Cyan
Write-Host ""
Write-Host "Next steps:" -ForegroundColor Yellow
Write-Host "1. Upload deploy/data/app.db to server" -ForegroundColor White
Write-Host "2. Replace data/app.db on server" -ForegroundColor White
Write-Host "3. Restart backend service" -ForegroundColor White
Write-Host ""
