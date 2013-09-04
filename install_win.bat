sc stop GoReportAgent
sc delete GoReportAgent
sc create GoReportAgent binPath= "%~dp0bin\srvany.exe" start= auto
sc description GoReportAgent "GoReportAgent"
reg add HKLM\SYSTEM\CurrentControlSet\Services\GoReportAgent\Parameters /v Application /d "%~dp0bin\AgentDaemon.exe" /f
reg add HKLM\SYSTEM\CurrentControlSet\Services\GoReportAgent\Parameters /f /v AppDirectory /d %~dp0bin\
sc start GoReportAgent
::@echo.
::@echo ReportAgent has started¡£
::@echo.
pause