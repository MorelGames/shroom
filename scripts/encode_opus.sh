ffmpeg -i "$1" -c:a libopus -b:a 64K -frame_duration 5 -compression_level 10 -f ogg "$2"
