# Redis 2 Sheet

Redis event to Google Sheet

## Usage

```bash
go run main.go run --spreadsheetID 1b6yWP-uurheUvLgg0PENgYyZmLknqGHkRP9gsuHTBmI --channels=events,copy
``````

## Test

```bash
go test -v -run TestPublishEvent
```

## Download

### Linux

```bash
curl -L https://github.com/addodelgrossi/redis2sheet/releases/download/v0.0.3/redis2sheet_linux-amd64 -o redis2sheet
chmod +x redis2sheet
```


## Configure Linux Services

## Create user 

```bash
sudo useradd -r -s /bin/false redis2sheet
```

### Download Application

```bash
sudo mkdir -p /opt/redis2sheet/
cd /opt/redis2sheet/
sudo curl -L https://github.com/addodelgrossi/redis2sheet/releases/download/v0.0.3/redis2sheet_linux-amd64 -o redis2sheet_linux-amd64
sudo chmod +x redis2sheet_linux-amd64
sudo chown -R redis2sheet:redis2sheet /opt/redis2sheet
```

### Create Service File


```bash
cat << EOF | sudo tee -a /etc/systemd/system/redis2sheet.service
[Unit]
Description=Redis to Sheet Service
After=multi-user.target
Conflicts=getty@tty1.service

[Service]
Type=simple
ExecStart=/opt/redis2sheet/redis2sheet_linux-amd64
#StandardInput=tty-force
StandardOutput=syslog
StandardError=syslog
SyslogIdentifier=redis2sheet
User=redis2sheet
WorkingDirectory=/opt/redis2sheet
TimeoutStopSec=10
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF
```

### Configure Service and start

https://www.baeldung.com/linux/systemd-create-user-services
https://betterprogramming.pub/unleashing-your-daemons-creating-services-on-ubuntu-731cd933e02e

$ sudo systemctl daemon-reload
$ sudo systemctl enable mydaemon
$ sudo systemctl start  mydaemon
$ sudo systemctl status mydaemon
$ sudo journalctl --unit=mydaemon

systemctl --user daemon-reload
systemctl --user start redis2sheet.service
systemctl --user enable redis2sheet.service
journalctl --user -u user_service


```bash
sudo systemctl enable redis2sheet.service
sudo systemctl daemon-reload
sudo systemctl status redis2sheet.service
sudo systemctl start redis2sheet.service
```

### Testing

```bash
docker run -it ubuntu /bin/bash
apt update
apt install sudo curl systemctl systemd
apt install sudo curl systemd
```
journalctl -u redis2sheet.service