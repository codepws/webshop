
@echo off

echo ��ǰ�̷���%~d0
echo ��ǰ�̷���·����%~dp0
echo ��ǰ������ȫ·����%~f0
echo ��ǰ�̷���·���Ķ��ļ�����ʽ��%~sdp0
echo ��ǰCMDĬ��Ŀ¼��%cd%
echo Ŀ¼���пո�Ҳ���Լ���""�����Ҳ���·��
echo ��ǰ�̷���"%~d0"
echo ��ǰ�̷���·����"%~dp0"
echo ��ǰ������ȫ·����"%~f0"
echo ��ǰ�̷���·���Ķ��ļ�����ʽ��"%~sdp0"
echo ��ǰCMDĬ��Ŀ¼��"%cd%"

echo =======================================================
echo %~dp0 
set exe_dir_path=%~dp0
cd /d %exe_dir_path% 

::vcffileΪ��Ҫ���Ƶ��ļ�����
set vcffile=*.go

echo ��ʼProto����...


echo �û�����Proto����...
protoc --proto_path=. --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative --proto_path="%GOROOT%/src"  user/*.proto
::ԴĿ¼������·����
set source_path=%exe_dir_path%\user
::Ŀ��Ŀ¼�����·����
set dest_srv_path=..\..\webshop-service\user-srv\proto
set dest_api_path=..\..\webshop-api\user-web\proto
cd %source_path%               
for /f "delims=" %%s in ('dir /b/a-d/s "%source_path%"\"%vcffile%"') do (
    echo %%s
    copy /y "%%s" %dest_api_path%
    copy /y "%%s" %dest_srv_path%
)
cd /d %exe_dir_path% 
 
echo ��Ʒ����Proto����...
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

echo ������Proto����...
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

echo ��������Proto����...
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
::������Ʒ����Proto����������
set source_path=%exe_dir_path%\goods
cd %source_path%
for /f "delims=" %%s in ('dir /b/a-d/s "%source_path%"\"%vcffile%"') do (
    echo %%s
    copy /y "%%s" %dest_api_path%
    copy /y "%%s" %dest_srv_path%
)
::����������Proto����������
set source_path=%exe_dir_path%\inventory
cd %source_path%       
for /f "delims=" %%s in ('dir /b/a-d/s "%source_path%"\"%vcffile%"') do (
    echo %%s
    copy /y "%%s" %dest_api_path%
    copy /y "%%s" %dest_srv_path%
)

cd /d %exe_dir_path% 



echo Proto�������... 

echo =======================================================

pause



:: protoc --proto_path=. --go_out=./goods --go_out=./inventory --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative --proto_path="%GOROOT%/src"  user/*.proto


