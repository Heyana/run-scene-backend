# Dify 局域网访问代理
# 使用方法: 以管理员身份运行 PowerShell，然后执行：
# powershell -ExecutionPolicy Bypass -File dify-proxy.ps1

$localEndpoint = "http://+:8080/"
$targetUrl = "http://localhost:3001"

Write-Host "=========================================" -ForegroundColor Cyan
Write-Host "  Dify 局域网访问代理" -ForegroundColor Cyan
Write-Host "=========================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "监听地址: 0.0.0.0:8080" -ForegroundColor Green
Write-Host "目标地址: $targetUrl" -ForegroundColor Green
Write-Host "局域网访问: http://192.168.3.39:8080" -ForegroundColor Yellow
Write-Host ""
Write-Host "按 Ctrl+C 停止服务" -ForegroundColor Gray
Write-Host ""

# 添加防火墙规则
Write-Host "正在配置防火墙..." -ForegroundColor Gray
try {
    netsh advfirewall firewall delete rule name="Dify Proxy 8080" 2>$null | Out-Null
    netsh advfirewall firewall add rule name="Dify Proxy 8080" dir=in action=allow protocol=TCP localport=8080 | Out-Null
    Write-Host "✅ 防火墙规则已添加" -ForegroundColor Green
} catch {
    Write-Host "⚠️  防火墙配置失败，可能需要管理员权限" -ForegroundColor Yellow
}
Write-Host ""

# 创建 HTTP 监听器
$listener = New-Object System.Net.HttpListener
$listener.Prefixes.Add($localEndpoint)
$listener.Start()

Write-Host "✅ 代理服务器已启动！" -ForegroundColor Green
Write-Host ""

try {
    while ($listener.IsListening) {
        $context = $listener.GetContext()
        $request = $context.Request
        $response = $context.Response
        
        # 构建目标 URL
        $targetUri = $targetUrl + $request.RawUrl
        
        Write-Host "[$(Get-Date -Format 'HH:mm:ss')] $($request.HttpMethod) $($request.RawUrl)" -ForegroundColor Gray
        
        try {
            # 转发请求
            $webRequest = [System.Net.WebRequest]::Create($targetUri)
            $webRequest.Method = $request.HttpMethod
            
            # 复制请求头
            foreach ($header in $request.Headers.AllKeys) {
                if ($header -notin @('Host', 'Connection')) {
                    try {
                        $webRequest.Headers.Add($header, $request.Headers[$header])
                    } catch {}
                }
            }
            
            # 复制请求体
            if ($request.HasEntityBody) {
                $webRequest.ContentLength = $request.ContentLength64
                $webRequest.ContentType = $request.ContentType
                $requestStream = $webRequest.GetRequestStream()
                $request.InputStream.CopyTo($requestStream)
                $requestStream.Close()
            }
            
            # 获取响应
            $webResponse = $webRequest.GetResponse()
            $response.StatusCode = [int]$webResponse.StatusCode
            
            # 复制响应头
            foreach ($header in $webResponse.Headers.AllKeys) {
                if ($header -notin @('Transfer-Encoding')) {
                    try {
                        $response.Headers.Add($header, $webResponse.Headers[$header])
                    } catch {}
                }
            }
            
            # 复制响应体
            $responseStream = $webResponse.GetResponseStream()
            $responseStream.CopyTo($response.OutputStream)
            $responseStream.Close()
            $webResponse.Close()
            
        } catch {
            Write-Host "  ❌ 错误: $_" -ForegroundColor Red
            $response.StatusCode = 500
            $buffer = [System.Text.Encoding]::UTF8.GetBytes("Proxy Error: $_")
            $response.OutputStream.Write($buffer, 0, $buffer.Length)
        }
        
        $response.Close()
    }
} finally {
    $listener.Stop()
    Write-Host "代理服务器已停止" -ForegroundColor Yellow
}
