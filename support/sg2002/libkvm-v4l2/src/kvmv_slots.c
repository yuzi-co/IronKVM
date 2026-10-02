#include "kvmv_slots.h"

#include <stdlib.h>

struct kvmv_slot *kvmv_slot_claim(struct kvmv_slots *slots, uint32_t size)
{
	unsigned int i;

	if (size == 0)
		return NULL;

	/* Walk the whole ring from where the last claim left off, so the
	 * buffers rotate rather than slot 0 being reused every time. */
	for (i = 0; i < KVMV_SLOT_COUNT; i++) {
		struct kvmv_slot *slot;

		slots->next = (slots->next + 1) % KVMV_SLOT_COUNT;
		slot = &slots->slot[slots->next];
		if (__atomic_load_n(&slot->in_use, __ATOMIC_ACQUIRE))
			continue;

		if (slot->data == NULL || slot->capacity < size) {
			uint8_t *grown = realloc(slot->data, size);

			if (grown == NULL)
				return NULL;
			slot->data = grown;
			slot->capacity = size;
		}
		slot->size = 0;
		slot->type = 0;
		__atomic_store_n(&slot->in_use, 1, __ATOMIC_RELEASE);
		return slot;
	}
	return NULL;
}

void kvmv_slot_abandon(struct kvmv_slot *slot)
{
	if (slot != NULL)
		__atomic_store_n(&slot->in_use, 0, __ATOMIC_RELEASE);
}

int kvmv_slot_release(struct kvmv_slots *slots, const uint8_t *data)
{
	unsigned int i;

	if (data == NULL)
		return -1;
	for (i = 0; i < KVMV_SLOT_COUNT; i++) {
		struct kvmv_slot *slot = &slots->slot[i];

		if (slot->data == data &&
		    __atomic_load_n(&slot->in_use, __ATOMIC_ACQUIRE)) {
			/* Read the type before the slot can be claimed again. */
			int type = slot->type;

			__atomic_store_n(&slot->in_use, 0, __ATOMIC_RELEASE);
			return type;
		}
	}
	return -1;
}

void kvmv_slots_free_all(struct kvmv_slots *slots)
{
	unsigned int i;

	for (i = 0; i < KVMV_SLOT_COUNT; i++) {
		free(slots->slot[i].data);
		slots->slot[i].data = NULL;
		slots->slot[i].capacity = 0;
		slots->slot[i].size = 0;
		__atomic_store_n(&slots->slot[i].in_use, 0, __ATOMIC_RELEASE);
	}
}
