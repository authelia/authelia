#!/bin/sh

# PROVIDE: authelia
# REQUIRE: DAEMON NETWORKING
# KEYWORD: shutdown

# Add the following lines to /etc/rc.conf to enable authelia:
# authelia_enable : set to "YES" to enable the daemon, default is "NO"

. /etc/rc.subr

name=authelia
rcvar=authelia_enable

load_rc_config $name

authelia_enable=${authelia_enable:-"NO"}

logfile="/var/log/${name}.log"

procname=/usr/local/bin/authelia
command="/usr/sbin/daemon"
command_args="-u root -o ${logfile} -t ${name} /usr/local/bin/authelia --config /usr/local/etc/authelia.yml"

run_rc_command "$1"
