# https://legacy.imagemagick.org/Usage/distorts/#barrel
convert tmp/image.jpg -virtual-pixel black -distort Barrel "0.0 0.0 -0.37 1.5" tmp/image_out.jpg && open tmp/image_out.jpg
