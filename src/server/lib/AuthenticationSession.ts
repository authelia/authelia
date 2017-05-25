

import express = require("express");
import U2f = require("u2f");

export interface AuthenticationSession {
    userid: string;
    first_factor: boolean;
    second_factor: boolean;
    identity_check?: {
        challenge: string;
        userid: string;
    };
    register_request?: U2f.Request;
    sign_request?: U2f.Request;
    email: string;
    groups: string[];
    redirect?: string;
}

export function reset(req: express.Request): void {
    const authSession: AuthenticationSession = {
        first_factor: false,
        second_factor: false,
        userid: undefined,
        email: undefined,
        groups: [],
        register_request: undefined,
        sign_request: undefined,
        identity_check: undefined,
        redirect: undefined
    };
    req.session.auth = authSession;
}

export function get(req: express.Request): AuthenticationSession {
    if (!req.session.auth)
        reset(req);
    return req.session.auth;
}