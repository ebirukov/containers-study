# Изолированный linux контейнер с сетью

Цель: создать изолированный процесс с сетевым доступом

## План создания процесса на примере реализации через shell:

Создаем изолированный процесс в со статически скомпилированным бинарным файлом
[busybox](https://github.com/mirror/busybox/), который занимает 2Мб после сборки.

Структура файловой системы:

```text
/
├── bin
    └── busybox (исполняемый файл)
├── proc (для монтирования псевдофайловой системы proc)
└── sys (не обязательно, для монтирования псевдофайловой системы sysfs)

```

Создаем контейнер

```shell
unshare --root=$(PWD)/container \ # монтируем файловую систему
        --map-root-user \ # мепит юзера на хосте на суперюзера в контейнере
        --uts \ # изолированное пространство хоста (для смены hostname и domainname)
        --net \ # изолированное пространство сети
        --ipc \ # изолированное пространство межпроцессорного взаимодействия
        --mount \ # изолированное файловое пространство
        --pid \ # изолированное пространство процессов
        --user \ # изолированное пространство пользователя
        --cgroup \ # изолированное пространство контрольных групп
        --fork \ # запуск в отдельном процессе
        --mount-proc \ # монтируем псевдофайловую систему proc
        busybox sh \ # запускаем оболочку из бинарника busybox
```

В запустившейся оболочке контейнера для читаемости можно сделать так: 
```shell
# если запущенная оболочка не имеет ссылки /proc/self/exe -> busybox,
# и для запуска команд можно используется busybox --install
# или создать ссылки вручную for cmd in $(busybox --list); do busybox ln -sf busybox /bin/$cmd; done
hostname busybox
PS1='\h:$ '
```

Убеждаемся что появился изолированное пространство процессов и изолированная сеть
```text
busybox:$ ps
PID   USER     TIME  COMMAND
    1 0         0:00 /busybox sh
    6 0         0:00 /busybox ps
busybox:$ ip link
1: lo: <LOOPBACK> mtu 65536 qdisc noop qlen 1000
    link/loopback 00:00:00:00:00:00 brd 00:00:00:00:00:00

```

А также изолированное пространство межпроцесного взаимодействия 
(т.е. не отображается семафоров очередей и сегменов совместно используемой памяти)
```text
busybox:$ ipcs

------ Message Queues --------
key        msqid      owner      perms      used-bytes   messages    

------ Shared Memory Segments --------
key        shmid      owner      perms      bytes      nattch     status      

------ Semaphore Arrays --------
key        semid      owner      perms      nsems  
```

А также можно убедится в изоляции cgroup
```text
busybox:$ mount -t sysfs sysfs /sys
busybox:$ ls -lF /sys/fs/cgroup/
total 0
busybox:$ mount -t cgroup2 none /sys/fs/cgroup
busybox:$ tree /sys/fs/cgroup/
/sys/fs/cgroup/
├── cgroup.controllers
...
└── pids.peak

0 directories, 49 files
```

### Передача трафика в контейнер

- Виртуальная сетевая пара (необходимы CAP_SYS_ADMIN на хосте)

Создаем виртуальную сетевую пару veth1 (хост) <-> veth2 (контейнер) 

```shell
# виртуальная пара между процессами 
# в разных анонимных (идентифицируем по pid) сетевых пространствах
PID=$(ps -C unshare -o pid=); \
 sudo ip link add veth1 netns $PID type veth peer name veth2 netns 1
```

Конфигурируем сетевой интерфейс на хосте

```shell
# Добавляем адрес
sudo ip a add 192.168.100.200 dev veth2
# Поднимаем интерфейс
sudo ip link set veth2 
# Добавляем маршруты для подсети
sudo ip route add 192.168.100.0/24 dev veth2
```

Аналогично конфигурируем сетевой интерфейс в контейнере

```shell
busybox:$ ip a add 192.168.100.100/24 dev veth1
busybox:$ ip link set veth1 up
busybox:$ ip route add 192.168.100.0/24 dev veth1
```

Проверяем корректность настройки в контейнере
```shell
busybox:$ ip a
1: lo: <LOOPBACK> mtu 65536 qdisc noop qlen 1000
    link/loopback 00:00:00:00:00:00 brd 00:00:00:00:00:00
2: veth1@if707: <BROADCAST,MULTICAST,UP,LOWER_UP,M-DOWN> mtu 1500 qdisc noqueue qlen 1000
    link/ether be:99:12:0f:db:9d brd ff:ff:ff:ff:ff:ff
    inet 192.168.100.100/24 scope global veth1
       valid_lft forever preferred_lft forever
    inet6 fe80::bc99:12ff:fe0f:db9d/64 scope link 
       valid_lft forever preferred_lft forever
busybox:$ ip route
192.168.100.0/24 dev veth1 scope link  src 192.168.100.100 
```

И работу сети
```shell
# В контейнере поднимаем tcp сервер
busybox:$ nc -vl -p 8080
listening on [::]:8080 ...

# На хосте отправляем в контейнер сообщение
echo "Hi" | nc -v 192.168.100.100 8080
```

В контейнере должны увидеть

```text
connect to [::ffff:192.168.100.100]:8080 from (null) ([::ffff:192.168.100.200]:25746)
Hi
```

Для подключения к контейнеру с хоста нужно выполнить:

```shell
PID=$(ps -C unshare -o pid=); \
sudo nsenter -a -t $PID chroot /tmp/container /busybox sh
```