/*
 * The frame buffers handed to the server. Same contract as the vendor library:
 * a pointer returned by kvmv_read_* stays valid until free_kvmv_data names it,
 * and the allocation is kept for the next frame that fits.
 */
#ifndef KVMV_SLOTS_H
#define KVMV_SLOTS_H

#include <stddef.h>
#include <stdint.h>

#define KVMV_SLOT_COUNT 4

struct kvmv_slot {
	uint8_t *data;
	uint32_t size;
	uint32_t capacity;
	uint8_t type;
	uint8_t in_use; /* read and written with __atomic */
};

struct kvmv_slots {
	struct kvmv_slot slot[KVMV_SLOT_COUNT];
	unsigned int next;
};

/* Claim a free slot able to hold size bytes. NULL when all are taken or out of memory. */
struct kvmv_slot *kvmv_slot_claim(struct kvmv_slots *slots, uint32_t size);
/* Give a claimed slot back without handing it out. */
void kvmv_slot_abandon(struct kvmv_slot *slot);
/* free_kvmv_data: returns the slot's type, or -1 when data names no slot. */
int kvmv_slot_release(struct kvmv_slots *slots, const uint8_t *data);
/* free_all_kvmv_data. */
void kvmv_slots_free_all(struct kvmv_slots *slots);

#endif
