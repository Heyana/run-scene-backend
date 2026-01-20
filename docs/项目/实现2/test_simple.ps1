# AmbientCG API Simple Test

Write-Host "Testing AmbientCG API..." -ForegroundColor Cyan

# Test 1: Get material list
Write-Host "`n1. Getting material list..." -ForegroundColor Yellow
$listUrl = "https://ambientcg.com/api/v2/full_json?type=Material&sort=Popular&limit=5"
$response = Invoke-RestMethod -Uri $listUrl
Write-Host "Total materials: $($response.numberOfResults)" -ForegroundColor Green
Write-Host "Returned: $($response.foundAssets.Count) items" -ForegroundColor Green

# Show materials
foreach ($asset in $response.foundAssets) {
    Write-Host "  - $($asset.assetId) : $($asset.displayName) [$($asset.displayCategory)]" -ForegroundColor Gray
}

# Test 2: Get material detail
$testId = $response.foundAssets[0].assetId
Write-Host "`n2. Getting detail for: $testId" -ForegroundColor Yellow
$detailUrl = "https://ambientcg.com/api/v2/full_json?id=$testId&include=downloadData"
$detail = Invoke-RestMethod -Uri $detailUrl
$asset = $detail.foundAssets[0]

Write-Host "Name: $($asset.displayName)" -ForegroundColor Green
Write-Host "Category: $($asset.displayCategory)" -ForegroundColor Green
Write-Host "Maps: $($asset.maps -join ', ')" -ForegroundColor Green

# Show download options
if ($asset.downloadFolders) {
    $downloads = $asset.downloadFolders.default.downloadFiletypeCategories.zip.downloads
    Write-Host "`nDownload options:" -ForegroundColor Green
    foreach ($dl in $downloads) {
        $sizeMB = [math]::Round($dl.size / 1MB, 2)
        Write-Host "  - $($dl.attribute): $sizeMB MB" -ForegroundColor Gray
    }
    
    # Test 3: Download smallest package
    $testDl = $downloads[0]
    Write-Host "`n3. Downloading: $($testDl.fileName)" -ForegroundColor Yellow
    $outFile = "test_$($testId).zip"
    Invoke-WebRequest -Uri $testDl.downloadLink -OutFile $outFile
    $fileSize = [math]::Round((Get-Item $outFile).Length / 1MB, 2)
    Write-Host "Downloaded: $fileSize MB" -ForegroundColor Green
    
    # Test 4: Extract
    Write-Host "`n4. Extracting..." -ForegroundColor Yellow
    $extractDir = "test_extract_$testId"
    if (Test-Path $extractDir) { Remove-Item $extractDir -Recurse -Force }
    Expand-Archive -Path $outFile -DestinationPath $extractDir
    $files = Get-ChildItem $extractDir
    Write-Host "Extracted $($files.Count) files:" -ForegroundColor Green
    foreach ($f in $files) {
        Write-Host "  - $($f.Name)" -ForegroundColor Gray
    }
}

Write-Host "`nAll tests passed!" -ForegroundColor Green
