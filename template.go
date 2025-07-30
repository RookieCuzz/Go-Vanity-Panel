package main

import (
	"html/template"
)

var indexTpl = template.Must(template.New("index").Parse(`<!DOCTYPE html>
<html>
<h1>{{.Host}}</h1>
<ul>
{{range .Repos}}<li><a href="{{.Source}}">{{.Import}} ({{.VCS}})</a></li>{{end}}
</ul>
</html>
`))

var repoTpl = template.Must(template.New("repo").Parse(`<!DOCTYPE html>
<html>
<head>
<meta http-equiv="Content-Type" content="text/html; charset=utf-8"/>
<meta name="go-import" content="{{.Import}} {{.VCS}} {{.Source}}">
<meta name="go-source" content="{{.Import}} {{.Display}}">
</head>
<body><a href="{{.Source}}">{{.Import}} ({{.VCS}})</a></body>
</html>`))

var configPanelTpl = template.Must(template.New("config").Parse(`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Go Vanity 配置面板</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; background-color: #f5f5f5; }
        .container { max-width: 800px; margin: 0 auto; background: white; padding: 20px; border-radius: 8px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
        h1 { color: #333; text-align: center; }
        .form-group { margin-bottom: 15px; }
        label { display: block; margin-bottom: 5px; font-weight: bold; color: #555; }
        input[type="text"], input[type="number"], select { width: 100%; padding: 8px; border: 1px solid #ddd; border-radius: 4px; box-sizing: border-box; }
        .path-group { border: 1px solid #ddd; padding: 15px; margin: 10px 0; border-radius: 4px; background-color: #f9f9f9; }
        .path-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 10px; }
        .btn { padding: 10px 20px; border: none; border-radius: 4px; cursor: pointer; font-size: 14px; }
        .btn-primary { background-color: #007bff; color: white; }
        .btn-danger { background-color: #dc3545; color: white; }
        .btn-success { background-color: #28a745; color: white; }
        .btn:hover { opacity: 0.8; }
        .status { padding: 10px; margin: 10px 0; border-radius: 4px; }
        .status.success { background-color: #d4edda; color: #155724; border: 1px solid #c3e6cb; }
        .status.error { background-color: #f8d7da; color: #721c24; border: 1px solid #f5c6cb; }
        .add-path-btn { margin: 10px 0; }
    </style>
</head>
<body>
    <div class="container">
        <h1>Go Vanity 配置面板</h1>
        
        <form id="configForm">
            <div class="form-group">
                <label for="host">Host:</label>
                <input type="text" id="host" name="host" value="{{.Config.Host}}" required>
            </div>
            
            <div class="form-group">
                <label for="cacheMaxAge">缓存最大时间 (秒):</label>
                <input type="number" id="cacheMaxAge" name="cacheMaxAge" value="{{.Config.CacheMaxAge}}" min="0">
            </div>
            
            <h3>路径配置</h3>
            <div id="paths">
                {{range $key, $path := .Config.Paths}}
                <div class="path-group" data-path-key="{{$key}}">
                    <div class="path-header">
                        <label>路径名称: <input type="text" name="pathName" value="{{$key}}" required></label>
                        <button type="button" class="btn btn-danger" onclick="removePath(this)">删除</button>
                    </div>
                    <div class="form-group">
                        <label>仓库地址:</label>
                        <input type="text" name="repo" value="{{$path.Repo}}" required>
                    </div>
                    <div class="form-group">
                        <label>版本控制系统:</label>
                        <select name="vcs">
                            <option value="git" {{if eq $path.VCS.String "git"}}selected{{end}}>Git</option>
                            <option value="hg" {{if eq $path.VCS.String "hg"}}selected{{end}}>Mercurial</option>
                            <option value="svn" {{if eq $path.VCS.String "svn"}}selected{{end}}>Subversion</option>
                            <option value="bzr" {{if eq $path.VCS.String "bzr"}}selected{{end}}>Bazaar</option>
                        </select>
                    </div>
                    <div class="form-group">
                        <label>显示信息 (可选):</label>
                        <input type="text" name="display" value="{{$path.Display}}">
                    </div>
                </div>
                {{end}}
            </div>
            
            <button type="button" class="btn btn-success add-path-btn" onclick="addPath()">添加路径</button>
            
            <div style="margin-top: 20px;">
                <button type="submit" class="btn btn-primary">保存配置</button>
                <button type="button" class="btn btn-success" onclick="reloadConfig()">重新加载</button>
            </div>
        </form>
        
        <div id="status"></div>
    </div>

    <script>
        function addPath() {
            const pathsDiv = document.getElementById('paths');
            const pathGroup = document.createElement('div');
            pathGroup.className = 'path-group';
            pathGroup.innerHTML = '<div class="path-header">' +
                '<label>路径名称: <input type="text" name="pathName" required></label>' +
                '<button type="button" class="btn btn-danger" onclick="removePath(this)">删除</button>' +
                '</div>' +
                '<div class="form-group">' +
                '<label>仓库地址:</label>' +
                '<input type="text" name="repo" required>' +
                '</div>' +
                '<div class="form-group">' +
                '<label>版本控制系统:</label>' +
                '<select name="vcs">' +
                '<option value="git" selected>Git</option>' +
                '<option value="hg">Mercurial</option>' +
                '<option value="svn">Subversion</option>' +
                '<option value="bzr">Bazaar</option>' +
                '</select>' +
                '</div>' +
                '<div class="form-group">' +
                '<label>显示信息 (可选):</label>' +
                '<input type="text" name="display">' +
                '</div>';
            pathsDiv.appendChild(pathGroup);
        }

        function removePath(button) {
            button.closest('.path-group').remove();
        }

        function showStatus(message, isError = false) {
            const statusDiv = document.getElementById('status');
            statusDiv.innerHTML = '<div class="status ' + (isError ? 'error' : 'success') + '">' + message + '</div>';
            setTimeout(() => {
                statusDiv.innerHTML = '';
            }, 3000);
        }

        document.getElementById('configForm').addEventListener('submit', async function(e) {
            e.preventDefault();
            
            const formData = new FormData(this);
            const config = {
                host: formData.get('host'),
                cacheMaxAge: parseInt(formData.get('cacheMaxAge')),
                paths: {}
            };

            // 收集路径配置
            const pathGroups = document.querySelectorAll('.path-group');
            pathGroups.forEach(group => {
                const pathName = group.querySelector('input[name="pathName"]').value;
                const repo = group.querySelector('input[name="repo"]').value;
                const vcs = group.querySelector('select[name="vcs"]').value;
                const display = group.querySelector('input[name="display"]').value;
                
                if (pathName && repo) {
                    config.paths[pathName] = {
                        repo: repo,
                        vcs: vcs,
                        display: display
                    };
                }
            });

            try {
                const response = await fetch('/api/config', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                    },
                    body: JSON.stringify(config)
                });

                if (response.ok) {
                    showStatus('配置保存成功！');
                } else {
                    const error = await response.text();
                    showStatus('保存失败: ' + error, true);
                }
            } catch (error) {
                showStatus('保存失败: ' + error.message, true);
            }
        });

        async function reloadConfig() {
            try {
                const response = await fetch('/api/config/reload', {
                    method: 'POST'
                });
                
                if (response.ok) {
                    showStatus('配置重新加载成功！');
                    setTimeout(() => {
                        window.location.reload();
                    }, 1000);
                } else {
                    const error = await response.text();
                    showStatus('重新加载失败: ' + error, true);
                }
            } catch (error) {
                showStatus('重新加载失败: ' + error.message, true);
            }
        }
    </script>
</body>
</html>`))
