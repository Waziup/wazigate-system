#!/bin/sh

# WAZIGATE_TAG="${WAZIGATE_TAG:-latest}"
# WAZIGATE_ADDR="${WAZIGATE_ADDR:-raspberrypi.local}"
# WAZIGATE_PASS="${WAZIGATE_PASS:-raspberry}"
# WAZIGATE_USER="${WAZIGATE_USER:-pi}"

# echo "[1 / 3] Test ping ..."
# ping -c 1 $WAZIGATE_ADDR

# echo "[2 / 3] Copy files to remote ..."
# sshpass -p $WAZIGATE_PASS rsync --relative -a \
#     ./ui/./node_modules/react/umd \
#     ./ui/./node_modules/react-dom/umd \
#     ./ui/./index.html \
#     ./ui/./dev.html \
#     ./ui/./favicon.ico  \
#     ./ui/./icons \
#     ./ui/./dist \
#     $WAZIGATE_USER@$WAZIGATE_ADDR:/home/pi/wazigate-system-ui/

# echo "[3 / 3] Copy files to container ..."
# sshpass -p $WAZIGATE_PASS ssh -T $WAZIGATE_USER@$WAZIGATE_ADDR <<EOF
# docker cp wazigate-system-ui/. waziup.wazigate-system:/app/ui/
# EOF

# echo "[     ] OK"