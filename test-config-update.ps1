# 测试配置更新功能

Write-Host "Testing configuration update..."

# 测试当前配置
Write-Host "1. Testing current configuration..."
try {
    $response = Invoke-WebRequest -Uri "http://localhost:8080/sample" -UseBasicParsing
    Write-Host "Sample path response: $($response.StatusCode)"
} catch {
    Write-Host "Error accessing /sample: $($_.Exception.Message)"
}

# 更新配置
Write-Host "2. Updating configuration..."
$newConfig = @{
    host = "go.example.com"
    cacheMaxAge = 3600
    paths = @{
        sample = @{
            repo = "https://github.com/example/sample"
            vcs = "git"
        }
        fuck = @{
            repo = "https://github.com/example/fuck"
            vcs = "git"
        }
    }
} | ConvertTo-Json -Depth 3

try {
    $response = Invoke-WebRequest -Uri "http://localhost:8080/api/config" -Method POST -Body $newConfig -ContentType "application/json" -UseBasicParsing
    Write-Host "Config update response: $($response.StatusCode)"
    Write-Host "Response content: $($response.Content)"
} catch {
    Write-Host "Error updating config: $($_.Exception.Message)"
}

# 等待一下让配置生效
Start-Sleep -Seconds 2

# 测试新配置
Write-Host "3. Testing new configuration..."
try {
    $response = Invoke-WebRequest -Uri "http://localhost:8080/fuck" -UseBasicParsing
    Write-Host "Fuck path response: $($response.StatusCode)"
    Write-Host "Response content: $($response.Content)"
} catch {
    Write-Host "Error accessing /fuck: $($_.Exception.Message)"
}

Write-Host "Test completed." 