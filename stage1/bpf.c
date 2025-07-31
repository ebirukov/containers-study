#include "vmlinux.h"
#include <bpf/bpf_helpers.h>
#include <bpf/bpf_core_read.h>

char LICENSE[] SEC("license") = "GPL";

struct event {
    u32 pid;
    char cmd[16];
    u32 user_ns_inum;
    u32 uts_ns_inum;
    u32 mnt_ns_inum;
};

struct {
    __uint(type, BPF_MAP_TYPE_RINGBUF);
    __uint(max_entries, 1 << 12);
} events SEC(".maps");

SEC("tracepoint/sched/sched_process_exec")
int tracepoint__exec(struct trace_event_raw_sched_process_exec *ctx) {
    // Резервируем место под event в кольцевом буфере для передачи его из памяти ядра в память пространства пользователя
    struct event *ev = bpf_ringbuf_reserve(&events, sizeof(*ev), 0);
    if (!ev) return 0;

    struct task_struct *task = (struct task_struct *)bpf_get_current_task();

    ev->pid = (u32)(bpf_get_current_pid_tgid() >> 32);

    // Заполняем event используя CO-RE макрос

    BPF_CORE_READ_INTO(&ev->cmd, task, comm);
    // Чтение task->cred->user_ns->ns.inum
    BPF_CORE_READ_INTO(&ev->user_ns_inum, task, cred, user_ns, ns.inum);
    // Чтение task->nsproxy->uts_ns->ns.inum
    BPF_CORE_READ_INTO(&ev->uts_ns_inum, task, nsproxy, uts_ns, ns.inum);
    // Чтение task->nsproxy->mnt_ns->ns.inum
    BPF_CORE_READ_INTO(&ev->mnt_ns_inum, task, nsproxy, mnt_ns, ns.inum);
    // Отправляем event в буфер
    bpf_ringbuf_submit(ev, 0);
    return 0;
}