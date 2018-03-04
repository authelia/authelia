FROM clems4ever/openldap

ENV SLAPD_ORGANISATION=MyCompany
ENV SLAPD_DOMAIN=example.com
ENV SLAPD_PASSWORD=password
ENV SLAPD_CONFIG_PASSWORD=password
ENV SLAPD_ADDITIONAL_MODULES=memberof
ENV SLAPD_ADDITIONAL_SCHEMAS=openldap
ENV SLAPD_FORCE_RECONFIGURE=true

ADD base.ldif /etc/ldap.dist/prepopulate/base.ldif
ADD access.rules /etc/ldap.dist/prepopulate/access.rules
