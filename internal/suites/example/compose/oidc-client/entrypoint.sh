#!/bin/bash

# SPDX-FileCopyrightText: 2019 Authelia
#
# SPDX-License-Identifier: Apache-2.0

while true;
do
    oidc-tester-app --issuer=https://login.example.com:8080 --id=oidc-tester-app --secret=foobar --scopes=openid,profile,email,groups --public-url=https://oidc.example.com:8080
    sleep 5
done