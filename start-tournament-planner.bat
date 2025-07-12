@echo off
echo Starting Tournament Planner...
start "Backend" cmd /k "D:\TournamentPlanner\start-backend.bat"
timeout /t 5
start "Frontend" cmd /k "D:\TournamentPlanner\start-frontend.bat"
pause
