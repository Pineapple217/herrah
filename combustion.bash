#!/bin/bash
# combustion: network

# Redirect output to the console
exec > >(exec tee -a /dev/tty0) 2>&1

# Timezone
systemd-firstboot --force --timezone=Europe/Brussels

zypper --non-interactive install htop powertop glibc-locale
powertop --auto-tune

# Leave a marker
echo "Configured with combustion" > /etc/issue.d/85-combustion.conf
echo "Configured with combustion" > /etc/issue.d/85-combustion.issue

# Close outputs and wait for tee to finish.
exec 1>&- 2>&-; wait;