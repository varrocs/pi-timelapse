WIDTH=320
HEIGHT=240
HTTP_PORT=8080

raspivid -o - -t 9999999 -fps 3 -w $WIDTH -h $HEIGHT --vflip --hflip | cvlc -vvv stream:///dev/stdin --sout "#standard{access=http,mux=ts,dst=:$HTTP_PORT}" :demux=h264
