TARGET=images
mkdir -p "$TARGET"
sudo mount -t ramfs -o size=20m ramfs "$TARGET"
sudo chown pi:pi "$TARGET"
