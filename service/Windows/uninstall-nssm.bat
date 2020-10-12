@echo off
cd /D "%~dp0"

set spath=%cd%\..\..
set sname=Galago

goto adminCheck

:adminCheck
	net session >nul 2>&1
	if %errorLevel% == 0 (
		goto main
	) else (
		echo Error: Admin rights are needed!
		echo Right click on the file and run it as administrator!
		pause
		exit /B 1
	)
	pause

:main
	echo Stopping and uninstalling "%sname%" service

	nssm stop %sname%
	nssm remove %sname% confirm

	pause