# Введение в пространство имен linux

## Внутреннее устройство неймспейсов

Структура ядра

```shell
bpftool btf dump file /sys/kernel/btf/vmlinux format c | grep -A 273 "struct task_struct {" | egrep "cred|nsproxy|{"

struct task_struct {
        const struct cred *ptracer_cred;
        const struct cred *real_cred;
        const struct cred *cred;
        struct nsproxy *nsproxy;


bpftool btf dump file /sys/kernel/btf/vmlinux format c | grep -A 40 "struct cred {" | egrep "user_ns|cred {"

struct cred {
        struct user_namespace *user_ns;

   
bpftool btf dump file /sys/kernel/btf/vmlinux format c | grep -A 10 "struct nsproxy {"

struct nsproxy {
        refcount_t count;
        struct uts_namespace *uts_ns;
        struct ipc_namespace *ipc_ns;
        struct mnt_namespace *mnt_ns;
        struct pid_namespace *pid_ns_for_children;
        struct net *net_ns;
        struct time_namespace *time_ns;
        struct time_namespace *time_ns_for_children;
        struct cgroup_namespace *cgroup_ns;
};

     
bpftool btf dump file /sys/kernel/btf/vmlinux format c | grep -A 15 "struct user_namespace {" | grep namespace

struct user_namespace {
        struct user_namespace *parent;
        ...
};
     
bpftool btf dump file /sys/kernel/btf/vmlinux format c | grep -A 15 "struct pid_namespace {" | grep namespace

struct pid_namespace {
        struct ns_common ns;
        struct pid_namespace *parent;
        struct user_namespace *user_ns;
        ...
};


bpftool btf dump file /sys/kernel/btf/vmlinux format c | grep -A 15 "struct mnt_namespace {" | grep namespace

struct mnt_namespace {
        struct ns_common ns;
        struct user_namespace *user_ns;
        ...

}};

```

Можно посмотреть с какими пространствами запускается процесс с помощью контрольной точки ядра
sched:sched_process_exec, которая срабатывает в момент выполнения системного вызова execve()

В терминале от привилегированного пользователя

```shell

root@vm-ubuntu-min:~# bpftrace -e '
tracepoint:sched:sched_process_exec
{
  $t = (struct task_struct *)curtask;

  printf("PID: %d\n", pid);
  printf("  user_ns:   %u\n", $t->cred->user_ns->ns.inum);
  printf("  puser_ns:   %u\n", $t->cred->user_ns->parent->ns.inum);
  printf("  uts_ns:    %u\n", $t->nsproxy->uts_ns->ns.inum);
  printf("  ipc_ns:    %u\n", $t->nsproxy->ipc_ns->ns.inum);
  printf("  mnt_ns:    %u\n", $t->nsproxy->mnt_ns->ns.inum);
  printf("  ppid_ns:    %u\n", $t->nsproxy->pid_ns_for_children->parent->ns.inum);
  printf("  pid_ns:    %u\n", $t->nsproxy->pid_ns_for_children->ns.inum);
  printf("  net_ns:    %u\n", $t->nsproxy->net_ns->ns.inum);
  printf("  time_ns:   %u\n", $t->nsproxy->time_ns->ns.inum);
  printf("  cgroup_ns: %u\n", $t->nsproxy->cgroup_ns->ns.inum);
  printf("  command: %s\n", $t->comm);
}'
Attaching 1 probe...
PID: 18228
  user_ns:   4026531837
  puser_ns:   0
  uts_ns:    4026531838
  ipc_ns:    4026531839
  mnt_ns:    4026531841
  ppid_ns:    0
  pid_ns:    4026531836
  net_ns:    4026531840
  time_ns:   4026531834
  cgroup_ns: 4026531835
  command: unshare
PID: 18228
  user_ns:   4026532642
  puser_ns:   4026531837
  uts_ns:    4026532644
  ipc_ns:    4026532645
  mnt_ns:    4026532643
  ppid_ns:    4026531836
  pid_ns:    4026532646
  net_ns:    4026532648
  time_ns:   4026532706
  cgroup_ns: 4026532647
  command: lsns
```
В другом терминале от непривилегированного пользователя

```text
если в ядре включен apparmor и параметр ядра kernel.apparmor_restrict_unprivileged_userns=1
то будет ошибка unshare: write failed /proc/self/uid_map: Operation not permitted
т.к. срабатывает защита записи
openat(AT_FDCWD, "/proc/self/uid_map", O_WRONLY) = 3
write(3, "0 1001 1", 8)                 = -1 EPERM (Operation not permitted)

можно изменить параметр ядра

```

```shell
unshare --map-root-user -UupinmCT lsns
        NS TYPE   NPROCS   PID USER COMMAND
4026531836 pid         1 18228 root lsns
4026532642 user        1 18228 root lsns
4026532643 mnt         1 18228 root lsns
4026532644 uts         1 18228 root lsns
4026532645 ipc         1 18228 root lsns
4026532647 cgroup      1 18228 root lsns
4026532648 net         1 18228 root lsns
4026532706 time        1 18228 root lsns


```

Или в терминале
```text
ubunti@vm-ubuntu:~$ unshare --map-root-user -UupinmCT sleep 10&
[1] 18231
ubunti@vm-ubuntu:~$ stat -Lc "%i %n" /proc/18231/ns/*;
4026532647 /proc/18231/ns/cgroup
4026532645 /proc/18231/ns/ipc
4026532643 /proc/18231/ns/mnt
4026532648 /proc/18231/ns/net
4026531836 /proc/18231/ns/pid
stat: cannot statx '/proc/18231/ns/pid_for_children': No such file or directory
4026532706 /proc/18231/ns/time
4026532706 /proc/18231/ns/time_for_children
4026532642 /proc/18231/ns/user
4026532644 /proc/18231/ns/uts

```