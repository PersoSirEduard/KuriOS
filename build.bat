@echo off
echo Building...
go build -o dist/kurios.exe
echo Setting up...
copy "help.txt" "dist/help.txt"
copy "config.json" "dist/config.json"
copy "directory.json" "dist/directory.json"
copy "LICENSE" "dist/LICENSE"
copy "kode.exe" "dist/kode.exe"
echo Done.