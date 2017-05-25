
import sinon = require("sinon");
import BluebirdPromise = require("bluebird");
import assert = require("assert");
import U2FSignPost = require("../../../../../../src/server/lib/routes/secondfactor/u2f/sign/post");
import AuthenticationSession = require("../../../../../../src/server/lib/AuthenticationSession");
import winston = require("winston");

import ExpressMock = require("../../../../mocks/express");
import UserDataStoreMock = require("../../../../mocks/UserDataStore");
import ServerVariablesMock = require("../../../../mocks/ServerVariablesMock");
import U2FMock = require("../../../../mocks/u2f");
import U2f = require("u2f");

describe("test u2f routes: sign", function () {
    let req: ExpressMock.RequestMock;
    let res: ExpressMock.ResponseMock;
    let userDataStore: UserDataStoreMock.UserDataStore;
    let mocks: ServerVariablesMock.ServerVariablesMock;
    let authSession: AuthenticationSession.AuthenticationSession;

    beforeEach(function () {
        req = ExpressMock.RequestMock();
        req.app = {};

        mocks = ServerVariablesMock.mock(req.app);
        mocks.logger = winston;

        req.session = {};
        AuthenticationSession.reset(req as any);
        authSession = AuthenticationSession.get(req as any);
        authSession.userid = "user";
        authSession.first_factor = true;
        authSession.second_factor = false;
        authSession.identity_check = {
            challenge: "u2f-register",
            userid: "user"
        };
        req.headers = {};
        req.headers.host = "localhost";

        const options = {
            inMemoryOnly: true
        };

        userDataStore = UserDataStoreMock.UserDataStore();
        userDataStore.set_u2f_meta = sinon.stub().returns(BluebirdPromise.resolve({}));
        userDataStore.get_u2f_meta = sinon.stub().returns(BluebirdPromise.resolve({}));
        mocks.userDataStore = userDataStore;

        res = ExpressMock.ResponseMock();
        res.send = sinon.spy();
        res.json = sinon.spy();
        res.status = sinon.spy();
    });

    describe("test signing", () => {
        it("should return status code 204", function () {
            const expectedStatus = {
                keyHandle: "keyHandle",
                publicKey: "pbk",
                certificate: "cert"
            };
            const u2f_mock = U2FMock.U2FMock();
            u2f_mock.checkSignature.returns(expectedStatus);

            authSession.sign_request = {
                appId: "app",
                challenge: "challenge",
                keyHandle: "key",
                version: "U2F_V2"
            };
            mocks.u2f = u2f_mock;
            return U2FSignPost.default(req as any, res as any)
                .then(function () {
                    assert(authSession.second_factor);
                });
        });

        it("should return unauthorized error on registration request internal error", function (done) {
            res.send = sinon.spy(function (data: any) {
                assert.equal(500, res.status.getCall(0).args[0]);
                done();
            });

            const u2f_mock = U2FMock.U2FMock();
            u2f_mock.checkSignature.returns({ errorCode: 500 });

            authSession.sign_request = {
                appId: "app",
                challenge: "challenge",
                keyHandle: "key",
                version: "U2F_V2"
            };
            mocks.u2f = u2f_mock;
            U2FSignPost.default(req as any, res as any);
        });
    });
});

