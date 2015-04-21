import RPi.GPIO as GPIO
import time

input_pin = 18
output_pin = 15

def flash_led():
	for i in range(3):
		GPIO.output(output_pin, 1)
		time.sleep(0.250)
		GPIO.output(output_pin, 0)
		time.sleep(0.250)


def setup():
	GPIO.setmode(GPIO.BCM)

	GPIO.setup(input_pin, GPIO.IN, pull_up_down=GPIO.PUD_UP)
	GPIO.setup(output_pin, GPIO.OUT)
	GPIO.output(output_pin, 0)

def button_test():
	while True:
		input_state=GPIO.input(input_pin)
		if input_state == False:
			print("Button pressed")
			flash_led()
			time.sleep(0.2)

print("START WATCHING")
setup()
button_test()	
