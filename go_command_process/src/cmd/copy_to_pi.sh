FILENAME=./cmd
md5sum	"$FILENAME"
scp -P22022 "$FILENAME"  pi@raspberrypi.local:/home/pi/timelapse
