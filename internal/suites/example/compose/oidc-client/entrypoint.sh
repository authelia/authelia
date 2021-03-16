#!/bin/bash

while true;
do
    oidc-tester-app --oidc-provider-url https://login.example.com:8080 --client-id oidc-tester-app --client-secret foobar --scopes openid,profile,email --redirect-uri https://oidc.example.com:8080/oauth2/callback
    sleep 5
done