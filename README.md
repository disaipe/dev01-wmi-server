# Dev01 WMI Query daemon

```shell
dev01-wmi-server.exe -serve -addr=[:8090] -app.url=http://dev01.com/api/pcmon -app.secret=[secret word]
```

```shell
# install as Windows service
dev01-wmi-server.exe -srv.install

# start as Windows service
dev01-wmi-server.exe -srv ...

# uninstall Windows service
dev01-wmi-server.exe -srv.uninstall
```
