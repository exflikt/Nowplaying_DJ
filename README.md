ブラウザで現在のDJの名前を表示します。
適当なコンピュータにnginxなど任意のWebサーバーをインストールしてpublicブランチを`/var/www/html`や`/usr/share/nginx/html`などに配置してください。
`go run .` を実行することで `public/index.html` と `public/obs.html` (OBSのブラウザソース用) が生成されます。
[timetable.csv](/timetable.csv)を編集しmainブランチにプッシュすることで、publicブランチに変換後のHTMLドキュメントが自動的に作成されます。
