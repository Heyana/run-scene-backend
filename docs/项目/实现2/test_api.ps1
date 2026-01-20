# AmbientCG API 测试脚本
# 用于验证 API 可用性和数据结构

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "AmbientCG API 测试脚本" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# 1. 测试获取材质列表
Write-Host "1️⃣  测试获取材质列表..." -ForegroundColor Yellow
$listUrl = "https://ambientcg.com/api/v2/full_json?type=Material&sort=Popular&limit=5&offset=0"
try {
    $response = Invoke-RestMethod -Uri $listUrl -Method Get
    Write-Host "   ✅ 成功获取材质列表" -ForegroundColor Green
    Write-Host "   📊 总材质数: $($response.numberOfResults)" -ForegroundColor White
    Write-Host "   📦 本次返回: $($response.foundAssets.Count) 个" -ForegroundColor White
    Write-Host ""
    
    # 显示材质列表
    Write-Host "   材质列表:" -ForegroundColor White
    foreach ($asset in $response.foundAssets) {
        Write-Host "   - $($asset.assetId) | $($asset.displayName) | $($asset.displayCategory) | 下载量: $($asset.downloadCount)" -ForegroundColor Gray
    }
    Write-Host ""
    
    # 保存第一个材质的 ID 用于后续测试
    $testAssetId = $response.foundAssets[0].assetId
    
} catch {
    Write-Host "   ❌ 获取材质列表失败: $_" -ForegroundColor Red
    exit 1
}

# 2. 测试获取单个材质详情
Write-Host "2️⃣  测试获取材质详情 ($testAssetId)..." -ForegroundColor Yellow
$detailUrl = "https://ambientcg.com/api/v2/full_json?id=$testAssetId&include=downloadData"
try {
    $detail = Invoke-RestMethod -Uri $detailUrl -Method Get
    $asset = $detail.foundAssets[0]
    Write-Host "   ✅ 成功获取材质详情" -ForegroundColor Green
    Write-Host "   📝 名称: $($asset.displayName)" -ForegroundColor White
    Write-Host "   📂 分类: $($asset.displayCategory)" -ForegroundColor White
    Write-Host "   🏷️  标签: $($asset.tags -join ', ')" -ForegroundColor White
    Write-Host "   🗺️  贴图类型: $($asset.maps -join ', ')" -ForegroundColor White
    Write-Host ""
    
    # 显示下载选项
    if ($asset.downloadFolders) {
        Write-Host "   下载选项:" -ForegroundColor White
        $downloads = $asset.downloadFolders.default.downloadFiletypeCategories.zip.downloads
        foreach ($download in $downloads) {
            $sizeMB = [math]::Round($download.size / 1MB, 2)
            Write-Host "   - $($download.attribute): $sizeMB MB | $($download.fileName)" -ForegroundColor Gray
        }
        Write-Host ""
        
        # 保存第一个下载链接用于测试
        $testDownloadUrl = $downloads[0].downloadLink
        $testFileName = $downloads[0].fileName
        
    } else {
        Write-Host "   ⚠️  未找到下载信息" -ForegroundColor Yellow
        exit 1
    }
    
} catch {
    Write-Host "   ❌ 获取材质详情失败: $_" -ForegroundColor Red
    exit 1
}

# 3. 测试下载材质包（只下载最小的 1K 版本）
Write-Host "3️⃣  测试下载材质包 ($testFileName)..." -ForegroundColor Yellow
$outputPath = "test_download_$testAssetId.zip"
try {
    Write-Host "   ⏳ 正在下载..." -ForegroundColor Gray
    Invoke-WebRequest -Uri $testDownloadUrl -OutFile $outputPath -TimeoutSec 60
    
    $fileInfo = Get-Item $outputPath
    $sizeMB = [math]::Round($fileInfo.Length / 1MB, 2)
    Write-Host "   ✅ 下载成功" -ForegroundColor Green
    Write-Host "   📦 文件大小: $sizeMB MB" -ForegroundColor White
    Write-Host "   💾 保存路径: $outputPath" -ForegroundColor White
    Write-Host ""
    
} catch {
    Write-Host "   ❌ 下载失败: $_" -ForegroundColor Red
    exit 1
}

# 4. 测试解压材质包
Write-Host "4️⃣  测试解压材质包..." -ForegroundColor Yellow
$extractPath = "test_extract_$testAssetId"
try {
    if (Test-Path $extractPath) {
        Remove-Item $extractPath -Recurse -Force
    }
    Expand-Archive -Path $outputPath -DestinationPath $extractPath -Force
    
    $files = Get-ChildItem $extractPath
    Write-Host "   ✅ 解压成功" -ForegroundColor Green
    Write-Host "   📁 文件数量: $($files.Count)" -ForegroundColor White
    Write-Host ""
    
    Write-Host "   文件列表:" -ForegroundColor White
    foreach ($file in $files) {
        $fileSizeKB = [math]::Round($file.Length / 1KB, 2)
        Write-Host "   - $($file.Name) ($fileSizeKB KB)" -ForegroundColor Gray
    }
    Write-Host ""
    
} catch {
    Write-Host "   ❌ 解压失败: $_" -ForegroundColor Red
    exit 1
}

# 5. 统计信息
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "📊 测试统计" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "✅ 所有测试通过！" -ForegroundColor Green
Write-Host ""
Write-Host "测试材质: $testAssetId" -ForegroundColor White
Write-Host "下载文件: $outputPath" -ForegroundColor White
Write-Host "解压目录: $extractPath" -ForegroundColor White
Write-Host ""
Write-Host "💡 提示: 测试文件已保存，可以手动查看" -ForegroundColor Yellow
Write-Host ""

# 6. 清理选项
Write-Host "是否清理测试文件? (Y/N): " -NoNewline -ForegroundColor Yellow
$cleanup = Read-Host
if ($cleanup -eq "Y" -or $cleanup -eq "y") {
    Remove-Item $outputPath -Force
    Remove-Item $extractPath -Recurse -Force
    Write-Host "✅ 测试文件已清理" -ForegroundColor Green
} else {
    Write-Host "ℹ️  测试文件已保留" -ForegroundColor Cyan
}
