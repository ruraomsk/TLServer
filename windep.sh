#!/bin/bash
echo "Start to Vano deploy"
GOOS=windows GOARCH=amd64 go build -o "TLServer.exe" cmd/main.go
FILE=/mnt//mnt/TLServer/TLServer.exe
if [ -f "$FILE" ]; then
    echo "Mounted the server drive"
else
    echo "Mounting the server drive"
    sudo mount -t cifs -o username=Ivan,password=162747 \\\\192.168.115.120\\JanFant /mnt
fi
#sudo  cp ./data/*.sql /mnt/asud/setup
#sudo  cp ./data/*.mrk /mnt/asud/setup
#sudo  cp ./data/*.xml /mnt/asud/setup
sudo  cp *.exe /mnt/TLServer
#sudo  cp *.toml /mnt/asud/cmd
#sudo  cp save.bat /mnt/asud/cmd

# cp ./data/*.sql ~/vm/asud/setup
# cp ./data/*.mrk ~/vm/asud/setup
# cp ./data/*.xml ~/vm/asud/setup
# cp ag-server.exe ~/vm/asud/cmd
# cp *.toml ~/vm/asud/cmd
