#!/bin/bash

while true;
do
    oidc-tester-app --issuer=https://login.example.com:8080 --id=oidc-tester-app --secret=foobar --scopes=openid,profile,email,groups --public-url=https://oidc.example.com:8080
    sleep 5
done