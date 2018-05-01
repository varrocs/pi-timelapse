# Setup

Run
```
sudo raspi-config
```

ensure that "camera" has been enabled



```
sudo apt install vlc
```

Add sudo modprobe bcm2835-v4l2

# Scripts
## On the dev host
- mount_pi.sh
  Mounts the raspberry pi.
- copy_file_to_pi.sh
  Copy the necessary files to the pi

## On the pi
- start_streaming.sh
  Start RTP stream from the camera
