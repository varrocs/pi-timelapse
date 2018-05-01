PARENT_ID=1F3eNqaCxBnpDuUEYJ5R2eqiqUxsSaUB4
PATH=$1

if [ -z $PATH ]; then
	echo "Missing path parameter"
	exit 1
fi

./gdrive-linux-rpi upload -p "$PARENT_ID" "$PATH"
