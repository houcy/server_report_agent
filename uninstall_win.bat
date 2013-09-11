sc stop GoReportAgent 
sc delete GoReportAgent
taskkill /F /IM AgentUpdate.exe
taskkill /F /IM EccReportAgent.exe
::@echo.
::@echo GoReportAgent has been deleted.
::@echo.
pause