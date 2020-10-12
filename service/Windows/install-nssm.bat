@echo off
cd /D "%~dp0"

set spath=%cd%\..\..
set sname=Galago
set sdesc=A web-gallery written in go
set sfile=Galago.exe

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
	echo Installing "%sname%" service

	nssm install %sname% "%spath%\%sfile%"
	nssm set %sname% AppDirectory "%spath%"
	nssm set %sname% Description "%sdesc%"
	nssm set %sname% ObjectName "NT AUTHORITY\NetworkService" ""
	nssm set %sname% DisplayName "%sname%"
	nssm set %sname% Start SERVICE_AUTO_START
	nssm start %sname%

	echo Successfully installed and started "%sname%" service
	pause
