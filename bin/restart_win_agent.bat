@echo off

echo Stopping the Report Agent...
taskkill /F /IM AgentDaemon.exe
taskkill /F /IM AgentUpdate.exe
taskkill /F /IM EccReportAgent.exe

echo Starting Daemon for Report Agent...
start /b .\AgentDaemon.exe
echo Done!