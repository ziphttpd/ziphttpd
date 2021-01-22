@echo off

set TARGET=%1
set BASE=%~dp0

if "%TARGET%" == "" (
	echo setup.cmd targetfolder\
	exit /B 1
)

cd %BASE%
git pull

set EXEID=ziphttpd
set BUILDEXE=%BASE%%EXEID%.exe
set TARGETEXE=%TARGET%%EXEID%.exe

go build -o %BUILDEXE% cmd/main.go

if exist %TARGETEXE%.old del /F %TARGETEXE%.old
if exist %TARGETEXE% ren %TARGETEXE% %EXEID%.old
copy %BUILDEXE% %TARGETEXE%

exit /B 0
