#include "common.h"
#include "vmlinux.h"
#include <bpf/bpf_core_read.h>
#include <bpf/bpf_helpers.h>
#include <bpf/bpf_tracing.h>
#include <linux/errno.h>

char LICENSE[] SEC("license") = "Dual BSD/GPL";

#define FILE_NAME_LEN	32
#define PATH_LEN 256
#define NAME_MAX 255

struct file_path {
    unsigned char path[PATH_LEN];
};

struct callback_ctx {
    unsigned char *path;
    bool found;
};

struct file_open_audit_event {
    unsigned char path[NAME_MAX];
};

struct {
	__uint(type, BPF_MAP_TYPE_PERF_EVENT_ARRAY);
	__uint(key_size, sizeof(u32));
	__uint(value_size, sizeof(u32));
} fileopen_events SEC(".maps");

BPF_HASH(allowed_access_files, u32, struct file_path, 256);
BPF_HASH(denied_access_files, u32, struct file_path, 256);

static u64 cb_check_path(struct bpf_map *map, u32 *key, struct file_path *map_path, struct callback_ctx *ctx) {
    bpf_printk("checking ctx->found: %d, path: map_path: %s, ctx_path: %s", ctx->found, map_path->path, ctx->path);

    size_t size = strlen(map_path->path, NAME_MAX);
    if (strcmp(map_path->path, ctx->path, size) == 0) {
        ctx->found = 1;
    }

    return 0;
}

SEC("lsm/file_open")
int BPF_PROG(restricted_file_open, struct file *file)
{
    char task[TASK_COMM_LEN];

    bpf_get_current_comm(&task, sizeof(task));

    struct file *fp;
    struct dentry *dentry;
    const __u8 *filename;
    struct file_open_audit_event event = {};
    int ret = -1;

    if (bpf_d_path(&file->f_path, event.path, NAME_MAX) < 0) {
        return 0;
    }

    // bpf_printk("%s open %s\n", task, full_path);

    struct callback_ctx cb = { .path = event.path, .found = false};
    cb.found = false;
    bpf_for_each_map_elem(&denied_access_files, cb_check_path, &cb, 0);
    if (cb.found) {
        bpf_printk("Access Denied: %s\n", cb.path);
        ret = -EPERM;
        goto out;
    }

    bpf_for_each_map_elem(&allowed_access_files, cb_check_path, &cb, 0);
    if (cb.found) {
        ret = 0;
        goto out;
    }

out:
    bpf_perf_event_output((void *)ctx, &fileopen_events, BPF_F_CURRENT_CPU, &event, sizeof(event));
    return ret;
}