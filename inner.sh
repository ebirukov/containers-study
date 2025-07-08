#!/bin/bash
set -eux

ROOTFS=/tmp/container
rm -rf "$ROOTFS"
mkdir -p "$ROOTFS"/{bin,proc,sys,dev,tmp}

#cp /bin/busybox "$ROOTFS/bin/"

# Симлинки для нужных команд
#for cmd in sh mount unshare cat ls echo sleep id; do
#    ln -sf /bin/busybox "$ROOTFS/bin/$cmd"
#done

# Создание /dev/null и других устройств
#mknod -m 666 "$ROOTFS/dev/null" c 1 3
#mknod -m 666 "$ROOTFS/dev/zero" c 1 5
#mknod -m 666 "$ROOTFS/dev/tty" c 5 0

# Входим в новый user+mount namespace и chroot'имся
unshare -Um --map-root-user --fork -- bash -c "
    #mount -t tmpfs tmpfs $ROOTFS/dev
    mount -t proc proc $ROOTFS/proc
    mount -t sysfs sysfs $ROOTFS/sys

    cd $ROOTFS
    chroot $ROOTFS /bin/sh -c '
        echo \"[+] Внутри chroot + user namespace\"
        echo \"UID: \$(id -u), GID: \$(id -g)\"

        echo \"[+] /proc/self/uid_map:\"
        cat /proc/self/uid_map || echo \"(ошибка)\"

        echo \"[+] Попытка создать вложенный user namespace:\"
        unshare -U --fork /bin/sh -c \"
            echo \\\"[++] Успех! Сейчас мы во вложенном user namespace\\\"
            cat /proc/self/uid_map
        \"
    '
"

echo "[✓] Готово."
