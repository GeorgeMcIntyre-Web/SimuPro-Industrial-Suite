$token = $env:GITHUB_TOKEN
$repos = Get-Content "repos_all_utf8.json" | ConvertFrom-Json
$targetDir = "C:\Users\George\source\repos\SimuPro-Industrial-Suite\Process-Simulation"

if (-not (Test-Path $targetDir)) {
    New-Item -ItemType Directory -Path $targetDir
}

Set-Location $targetDir

foreach ($repo in $repos) {
    $repoName = $repo.name
    $cloneUrl = $repo.clone_url.Replace("https://github.com", "https://$token@github.com")
    
    if (Test-Path $repoName) {
        Write-Host "Skipping $repoName (already exists)" -ForegroundColor Yellow
    } else {
        Write-Host "Cloning $repoName..." -ForegroundColor Cyan
        git clone $cloneUrl
    }
}
