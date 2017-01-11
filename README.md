# constellation
Spin up constellations of rkt pods in a controlled fashion.

# Notes
mention that -v paths must be absolute paths not relative

# Todo
- Have constellation detect when a filemonitor is monitoring a file that is not exported from the container, and monitor it via its
  actual location in the filesystem (e.g. /var/rkt/cas/...)
# Known Bugs
- Setting a Volume with Kind to "empty" will result in constellation not being able to monitor the files referenced in those volumes, 
  but also not reporting an error
