
@echo off

echo 当前盘符：%~d0
echo 当前盘符和路径：%~dp0
echo 当前批处理全路径：%~f0
echo 当前盘符和路径的短文件名格式：%~sdp0
echo 当前CMD默认目录：%cd%
echo 目录中有空格也可以加入""避免找不到路径
echo 当前盘符："%~d0"
echo 当前盘符和路径："%~dp0"
echo 当前批处理全路径："%~f0"
echo 当前盘符和路径的短文件名格式："%~sdp0"
echo 当前CMD默认目录："%cd%"

echo =======================================================
echo %~dp0 
set exe_dir_path=%~dp0
cd /d %exe_dir_path% 

::vcffile为需要复制的文件类型
set vcffile=*.go

echo 开始Proto生成...


echo 用户服务Proto生成...
protoc --proto_path=. --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative --proto_path="%GOROOT%/src"  user/*.proto
::源目录（绝对路径）
set source_path=%exe_dir_path%\user
::目标目录（相对路径）
set dest_srv_path=..\..\webshop-service\user-srv\proto
set dest_api_path=..\..\webshop-api\user-web\proto
cd %source_path%               
for /f "delims=" %%s in ('dir /b/a-d/s "%source_path%"\"%vcffile%"') do (
    echo %%s
    copy /y "%%s" %dest_api_path%
    copy /y "%%s" %dest_srv_path%
)
cd /d %exe_dir_path% 
 
echo 商品服务Proto生成...
protoc --proto_path=. --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative --proto_path="%GOROOT%/src"  goods/*.proto
set source_path=%exe_dir_path%\goods
set dest_srv_path=..\..\webshop-service\goods-srv\proto
set dest_api_path=..\..\webshop-api\goods-web\proto
cd %source_path%      
for /f "delims=" %%s in ('dir /b/a-d/s "%source_path%"\"%vcffile%"') do (
    echo %%s
    copy /y "%%s" %dest_api_path%
    copy /y "%%s" %dest_srv_path%
)
cd /d %exe_dir_path% 

echo 库存服务Proto生成...
protoc --proto_path=. --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative --proto_path="%GOROOT%/src"  inventory/*.proto
set source_path=%exe_dir_path%\inventory
set dest_api_path=..\..\webshop-api\inventory-web\proto
set dest_srv_path=..\..\webshop-service\inventory-srv\proto
cd %source_path%               
for /f "delims=" %%s in ('dir /b/a-d/s "%source_path%"\"%vcffile%"') do (
    echo %%s
    copy /y "%%s" %dest_api_path%
    copy /y "%%s" %dest_srv_path%
)
cd /d %exe_dir_path%  

echo 订单服务Proto生成...
protoc --proto_path=. --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative --proto_path="%GOROOT%/src"  order/*.proto
set source_path=%exe_dir_path%\order
set dest_api_path=..\..\webshop-api\order-web\proto
set dest_srv_path=..\..\webshop-service\order-srv\proto
cd %source_path%               
for /f "delims=" %%s in ('dir /b/a-d/s "%source_path%"\"%vcffile%"') do (
    echo %%s
    copy /y "%%s" %dest_api_path%
    copy /y "%%s" %dest_srv_path%
) 
::拷贝商品服务Proto到订单服务
set source_path=%exe_dir_path%\goods
cd %source_path%
for /f "delims=" %%s in ('dir /b/a-d/s "%source_path%"\"%vcffile%"') do (
    echo %%s
    copy /y "%%s" %dest_api_path%
    copy /y "%%s" %dest_srv_path%
)
::拷贝库存服务Proto到订单服务
set source_path=%exe_dir_path%\inventory
cd %source_path%       
for /f "delims=" %%s in ('dir /b/a-d/s "%source_path%"\"%vcffile%"') do (
    echo %%s
    copy /y "%%s" %dest_api_path%
    copy /y "%%s" %dest_srv_path%
)

cd /d %exe_dir_path% 



echo Proto生成完成... 

echo =======================================================

pause



:: protoc --proto_path=. --go_out=./goods --go_out=./inventory --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative --proto_path="%GOROOT%/src"  user/*.proto


