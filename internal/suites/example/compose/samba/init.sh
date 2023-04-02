#!/bin/bash

# SPDX-FileCopyrightText: 2019 Authelia
#
# SPDX-License-Identifier: Apache-2.0

set -e

appSetup () {

	# Set variables
	DOMAIN=${DOMAIN:-SAMDOM.LOCAL}
	DOMAINPASS=${DOMAINPASS:-youshouldsetapassword}
	NOCOMPLEXITY=${NOCOMPLEXITY:-false}
	INSECURELDAP=${INSECURELDAP:-false}

	LDOMAIN=${DOMAIN,,}
	UDOMAIN=${DOMAIN^^}
	URDOMAIN=${UDOMAIN%%.*}

	# Set up samba
	mv /etc/krb5.conf /etc/krb5.conf.orig
	echo "[libdefaults]" > /etc/krb5.conf
	echo "    dns_lookup_realm = false" >> /etc/krb5.conf
	echo "    dns_lookup_kdc = true" >> /etc/krb5.conf
	echo "    default_realm = ${UDOMAIN}" >> /etc/krb5.conf
	# If the finished file isn't there, this is brand new, we're not just moving to a new container
	if [[ ! -f /etc/samba/external/smb.conf ]]; then
		mv /etc/samba/smb.conf /etc/samba/smb.conf.orig
		samba-tool domain provision --use-rfc2307 --domain=${URDOMAIN} --realm=${UDOMAIN} --server-role=dc --dns-backend=SAMBA_INTERNAL --adminpass=${DOMAINPASS}
		if [[ ${NOCOMPLEXITY,,} == "true" ]]; then
			samba-tool domain passwordsettings set --complexity=off
			samba-tool domain passwordsettings set --history-length=0
			samba-tool domain passwordsettings set --min-pwd-length=3
			samba-tool domain passwordsettings set --min-pwd-age=0
			samba-tool domain passwordsettings set --max-pwd-age=0
		fi
		sed -i "/\[global\]/a \
			\\\tidmap_ldb:use rfc2307 = yes\\n\
			wins support = yes\\n\
			template shell = /bin/bash\\n\
			winbind nss info = rfc2307\\n\
			idmap config ${URDOMAIN}: range = 10000-20000\\n\
			idmap config ${URDOMAIN}: backend = ad\
			" /etc/samba/smb.conf
		if [[ ${INSECURELDAP,,} == "true" ]]; then
			sed -i "/\[global\]/a \
				\\\tldap server require strong auth = no\
				" /etc/samba/smb.conf
		fi
		# Once we are set up, we'll make a file so that we know to use it if we ever spin this up again
		mkdir -p /etc/samba/external
		cp /etc/samba/smb.conf /etc/samba/external/smb.conf
	else
		cp /etc/samba/external/smb.conf /etc/samba/smb.conf
	fi

	# Set up supervisor
	mkdir /etc/supervisor.d/
	echo "[supervisord]" > /etc/supervisor.d/supervisord.ini
	echo "nodaemon=true" >> /etc/supervisor.d/supervisord.ini
	echo "" >> /etc/supervisor.d/supervisord.ini
	echo "[program:samba]" >> /etc/supervisor.d/supervisord.ini
	echo "command=/usr/sbin/samba -i" >> /etc/supervisor.d/supervisord.ini

	appProvision
	appStart
}

appStart () {
	/usr/bin/supervisord
}

appProvision () {
  samba-tool user setpassword administrator --newpassword=password
  samba-tool ou create "OU=Users"
  samba-tool ou create "OU=Groups"
  samba-tool group add dev --groupou=OU=Groups
  samba-tool group add admins --groupou=OU=Groups
  samba-tool user create john password --userou=OU=Users --use-username-as-cn --given-name John --surname Doe --mail-address john.doe@authelia.com
  samba-tool user create harry password --userou=OU=Users --use-username-as-cn --given-name Harry --surname Potter --mail-address harry.potter@authelia.com
  samba-tool user create bob password --userou=OU=Users --use-username-as-cn --given-name Bob --surname Dylan --mail-address bob.dylan@authelia.com
  samba-tool user create james password --userou=OU=Users --use-username-as-cn --given-name James --surname Dean --mail-address james.dean@authelia.com
  samba-tool group addmembers "dev" john,bob
  samba-tool group addmembers "admins" john
}

case "$1" in
	start)
		if [[ -f /etc/samba/external/smb.conf ]]; then
			cp /etc/samba/external/smb.conf /etc/samba/smb.conf
			appStart
		else
			echo "Config file is missing."
		fi
		;;
	setup)
		# If the supervisor conf isn't there, we're spinning up a new container
		if [[ -f /etc/supervisor.d/supervisord.ini ]]; then
			appStart
		else
			appSetup
		fi
		;;
esac

exit 0