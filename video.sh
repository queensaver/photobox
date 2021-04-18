ffmpeg -f video4linux2 -r 10 -s 1920x1080 -i /dev/video0 out.avi
# ffmpeg -f v4l2 -r 30 -s 1920x1080 -i /dev/video0 out.avi
