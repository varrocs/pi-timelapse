echo Starting taking pictures
raspistill -n -ex night -ifx denoise -vf -e png -tl 10000 -t 360000000 -o "image-%04d.png"

