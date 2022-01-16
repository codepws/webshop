@echo off

setlocal

if exist order-srv.bat goto ok
echo order-srv.bat must be run from its folder
goto end

: ok

::set OLDGOPATH=%GOPATH%
::echo old GOPATH=%OLDGOPATH%
::set GOPATH=%cd%
::echo new GOPATH=%GOPATH%

::临时环境变量
set WEBSHOP_DEBUG=true

:: 减小包体大小
:: go build -ldflags "-s"
:: -s:omit symbol table and debug info(忽略符号表和debug信息)


go run order-srv/main.go -work .

::set GOPATH=OLDGOPATH


:end
::echo finished 