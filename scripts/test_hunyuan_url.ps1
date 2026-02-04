# 测试混元3D URL生成
# 此脚本用于测试 AfterFind 钩子是否正确生成 URL

Write-Host "测试混元3D URL生成..." -ForegroundColor Green
Write-Host ""

# 测试路径
$testPaths = @(
    "\\192.168.3.10\project\editor_v2\static\hunyuan\2026\02\6b9eb5df9ff992c0.glb",
    "\\192.168.3.10\project\editor_v2\static\hunyuan\2026\02\6b9eb5df9ff992c0.png",
    "static/hunyuan/2026/02/file.glb",
    "./static/hunyuan/2026/02/file.glb"
)

Write-Host "测试路径:" -ForegroundColor Yellow
foreach ($path in $testPaths) {
    Write-Host "  - $path"
}
Write-Host ""

# 调用API测试
$baseUrl = "http://192.168.3.39:23359"
$apiUrl = "$baseUrl/api/hunyuan/tasks"

Write-Host "调用API: $apiUrl" -ForegroundColor Yellow
Write-Host ""

try {
    $response = Invoke-RestMethod -Uri $apiUrl -Method Get -ContentType "application/json"
    
    if ($response.code -eq 0) {
        Write-Host "✓ API调用成功" -ForegroundColor Green
        Write-Host ""
        
        $tasks = $response.data.list
        Write-Host "任务列表 (共 $($tasks.Count) 个):" -ForegroundColor Cyan
        Write-Host ""
        
        foreach ($task in $tasks) {
            if ($task.status -eq "DONE") {
                Write-Host "任务 #$($task.id) - $($task.name)" -ForegroundColor White
                Write-Host "  状态: $($task.status)" -ForegroundColor Green
                Write-Host "  NAS路径: $($task.nasPath)" -ForegroundColor Gray
                Write-Host "  文件URL: $($task.fileUrl)" -ForegroundColor Cyan
                Write-Host "  缩略图URL: $($task.thumbnailUrl)" -ForegroundColor Cyan
                Write-Host ""
                
                # 测试URL是否可访问
                if ($task.fileUrl) {
                    try {
                        $testResponse = Invoke-WebRequest -Uri $task.fileUrl -Method Head -TimeoutSec 5
                        if ($testResponse.StatusCode -eq 200) {
                            Write-Host "  ✓ 文件URL可访问" -ForegroundColor Green
                        }
                    } catch {
                        Write-Host "  ✗ 文件URL无法访问: $($_.Exception.Message)" -ForegroundColor Red
                    }
                }
                
                if ($task.thumbnailUrl) {
                    try {
                        $testResponse = Invoke-WebRequest -Uri $task.thumbnailUrl -Method Head -TimeoutSec 5
                        if ($testResponse.StatusCode -eq 200) {
                            Write-Host "  ✓ 缩略图URL可访问" -ForegroundColor Green
                        }
                    } catch {
                        Write-Host "  ✗ 缩略图URL无法访问: $($_.Exception.Message)" -ForegroundColor Red
                    }
                }
                
                Write-Host ""
            }
        }
    } else {
        Write-Host "✗ API返回错误: $($response.msg)" -ForegroundColor Red
    }
} catch {
    Write-Host "✗ API调用失败: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host ""
Write-Host "测试完成" -ForegroundColor Green
