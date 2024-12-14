#!bin/bash

sudo killall rabbit-connect
sudo ./bin/rabbit-connect -S -l=:3001 -c=172.16.0.1/24 -p=ws &
echo "started!"
