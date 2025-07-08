# Создание виртуалки

## В lxc (версия выше 6.х)

```shell

lxc init ubuntu:24.04 vm-ubuntu --vm

lxc config set vm-ubuntu user.user-data - < user-data.yaml

```

Добавим общую папку с хостом

```shell

lxc config device add vm-ubuntu shared-virtiofs disk source=/mnt/share path=/mnt/share readonly=false
```

Запускаем и заходим

```shell
    lxc start vm-ubuntu
    lxc exec vm-ubuntu -- bash
```

Добавляем пароль дефолтному непривелигерованному пользователю

```text

root@vm-ubuntu:~# passwd ubuntu
New password: 
Retype new password: 
passwd: password updated successfully
root@vm-ubuntu:~# login
vm-ubuntu login: ubuntu 
Password: 
Welcome to Ubuntu 24.04.2 LTS (GNU/Linux 6.8.0-62-generic x86_64)
...

ubuntu@vm-ubuntu:~$ groups
ubuntu adm cdrom sudo dip lxd

```