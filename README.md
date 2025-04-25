# HomeKit Garage Shutter

## Build

1. Install and/or make sure Docker is running
2. Run `make build/rpi/v6`


## Installation

## Linux

Install [Log2Ram](https://github.com/azlux/log2ram)

```
echo "deb [signed-by=/usr/share/keyrings/azlux-archive-keyring.gpg] http://packages.azlux.fr/debian/ bullseye main" | sudo tee /etc/apt/sources.list.d/azlux.list
sudo curl -o /usr/share/keyrings/azlux-archive-keyring.gpg  https://azlux.fr/repo.gpg
sudo apt update
sudo apt install log2ram
```


## App

1. Open a shell on the Raspberry Pi and create a directory for the service
   ```
   mkdir /opt/homekit-garage-shutter
   ```
2. Copy the binary to the Raspberry Pi
   ```
   scp out/homekit-garage-shutter-linux-armel-6 user@rpi.local/opt/homekit-garage-shutter/homekit-garage-shutter
   ```
3. Copy the Systemd service file  tot he Raspberry Pi
   ```
   scp init/homekit-garage-shutter.service user@rpi.local/opt/homekit-garage-shutter/
   ```
4. From the shell on the Raspberry Pi to install and run the Systemd service
   ```
   sudo cp /opt/homekit-garage-shutter/homekit-garage-shutter.service /etc/systemd/system/
   sudo systemctl daemon-reload
   sudo systemctl enable homekit-garage-shutter.service
   sudo systemctl start homekit-garage-shutter.service
   ```