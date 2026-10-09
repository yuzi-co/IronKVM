#include "slot_scripts.h"

#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>

int slot_refuses(const char *list, const char *name)
{
	FILE *fp = fopen(list, "r");
	if (fp == NULL) return 0;

	char line[256];
	int found = 0;
	while (!found && fgets(line, sizeof(line), fp) != NULL) {
		char *s = line;
		while (*s == ' ' || *s == '\t') s++;
		if (*s == '#') continue;
		size_t n = strlen(s);
		while (n > 0 && (s[n - 1] == '\n' || s[n - 1] == '\r' ||
				 s[n - 1] == ' ' || s[n - 1] == '\t'))
			s[--n] = '\0';
		if (n > 0 && strcmp(s, name) == 0) found = 1;
	}
	fclose(fp);
	return found;
}

int install_boot_script(const char *list, const char *src_dir,
			const char *dest_dir, const char *name)
{
	if (slot_refuses(list, name)) {
		printf("new_app_init: %s is refused by this slot (%s), not installed\n",
		       name, list);
		return 0;
	}
	char cmd[512];
	snprintf(cmd, sizeof(cmd), "cp -f %s/%s %s/", src_dir, name, dest_dir);
	system(cmd);
	return 1;
}

enum s30wifi_action s30wifi_action(const char *list, int vendor, int wifi)
{
	if (slot_refuses(list, "S30wifi")) return S30WIFI_REMOVE;
	if (vendor && wifi) return S30WIFI_INSTALL;
	if (access(list, F_OK) == 0) return S30WIFI_KEEP;
	return S30WIFI_REMOVE;
}
