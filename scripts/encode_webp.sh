cwebp "$1" -metadata none -o "$2"

# Even better (AVIF support required)
avifencode "$1" -i "$2"
