@echo off

echo Stopping the Report Agent...
taskkill /F /IM AgentUpdate.exe
taskkill /F /IM EccReportAgent.exe

echo Starting Daemon for Report Agent...
sc start GoReportAgent
echo Done!

pause