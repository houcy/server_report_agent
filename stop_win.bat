@echo off

echo Stopping the Report Agent...
sc stop GoReportAgent
taskkill /F /IM AgentUpdate.exe
taskkill /F /IM EccReportAgent.exe

echo Done!
pause