import picamera
import fractions
import time

OUTPUT_FILE="video.h264"
#RESOLUTION_FULL=(2592, 1944)
RESOLUTION_FULL=(1920, 1080)
RESOLUTION_VIDEO=(1920, 1080)
FRAMERATE_VIDEO=fractions.Fraction(1,5)
TIME_VIDEO=3600*24

cam = picamera.PiCamera()

cam.vflip = True
cam.resolution = RESOLUTION_FULL

cam.start_preview()
time.sleep(5)
cam.stop_preview()

cam.framerate = FRAMERATE_VIDEO
cam.start_recording(OUTPUT_FILE, format='h264', resize=RESOLUTION_VIDEO)
cam.wait_recording(TIME_VIDEO)

cam.close()
