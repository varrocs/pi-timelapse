IMAGE_NAME=$1

if [ -z $IMAGE_NAME ]; then
	echo Missing image name parameter $IMAGE_NAME
	exit 1
fi

echo Starting taking picture $IMAGE_NAME
raspistill -n \
	-ex night \
	-ifx denoise \
	-vf -hf \
	-e png \
	-o "$IMAGE_NAME"
