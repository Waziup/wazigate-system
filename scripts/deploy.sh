#!/bin/sh
set -e

WAZIGATE_TAG="${WAZIGATE_TAG:-latest}"
WAZIGATE_ADDR="${WAZIGATE_ADDR:-raspberrypi.local}"
WAZIGATE_PASS="${WAZIGATE_PASS:-raspberry}"
WAZIGATE_USER="${WAZIGATE_USER:-pi}"

echo "[1 / 7] Test ping ..."
ping -c 1 $WAZIGATE_ADDR

echo "[2 / 7] Building image ..."
docker buildx build --platform=linux/arm/v7 --tag waziup/wazigate-system:$WAZIGATE_TAG --build-arg MODE=dev --load .

docker image save waziup/wazigate-system:$WAZIGATE_TAG -o wazigate-system.tar

echo "[3 / 7] Transfering image archive ..."
sshpass -p $WAZIGATE_PASS scp wazigate-system.tar $WAZIGATE_USER@$WAZIGATE_ADDR:/home/pi/wazigate-system.tar

echo "[4 / 7] Connecting to remote machine ..."

sshpass -p $WAZIGATE_PASS ssh -T $WAZIGATE_USER@$WAZIGATE_ADDR <<EOF
cd /var/lib/wazigate
echo "[5 / 7] Removing remote container ..."
docker rm -f waziup.wazigate-system
echo "[6 / 7] Loading image new ..."
docker image load -i /home/pi/wazigate-system.tar
echo "[7 / 7] Recreating remote container ..."
docker run -d --restart=unless-stopped --network=wazigate --name waziup.wazigate-system \
    -v "\$PWD/apps/waziup.wazigate-system:/var/lib/waziapp" \
    -v "/var/run:/var/run" \
    -v "/sys/class/gpio:/sys/class/gpio" \
    -v "/dev/mem:/dev/mem" \
    --privileged \
    --health-cmd="curl --fail --unix-socket /var/lib/waziapp/proxy.sock http://localhost/ || exit 1" \
    --health-interval=10s \
    waziup/wazigate-system
docker logs -f waziup.wazigate-system
EOF