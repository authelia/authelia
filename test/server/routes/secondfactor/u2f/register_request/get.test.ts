
import sinon = require("sinon");
import BluebirdPromise = require("bluebird");
import assert = require("assert");
import U2FRegisterRequestGet = require("../../../../../../src/server/lib/routes/secondfactor/u2f/register_request/get");
import AuthenticationSession = require("../../../../../../src/server/lib/AuthenticationSession");
import winston = require("winston");

import ExpressMock = require("../../../../mocks/express");
import UserDataStoreMock = require("../../../../mocks/UserDataStore");
import U2FMock = require("../../../../mocks/u2f");
import ServerVariablesMock = require("../../../../mocks/ServerVariablesMock");
import U2f = require("u2f");

describe("test u2f routes: register_request", function () {
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

    describe("test registration request", () => {
        it("should send back the registration request and save it in the session", function () {
            const expectedRequest = {
                test: "abc"
            };
            const user_key_container = {};
            const u2f_mock = U2FMock.U2FMock();
            u2f_mock.request.returns(BluebirdPromise.resolve(expectedRequest));

            mocks.u2f = u2f_mock;
            return U2FRegisterRequestGet.default(req as any, res as any)
                .then(function () {
                    assert.deepEqual(expectedRequest, res.json.getCall(0).args[0]);
                });
        });

        it("should return internal error on registration request", function (done) {
            res.send = sinon.spy(function (data: any) {
                assert.equal(500, res.status.getCall(0).args[0]);
                done();
            });
            const user_key_container = {};
            const u2f_mock = U2FMock.U2FMock();
            u2f_mock.request.returns(BluebirdPromise.reject("Internal error"));

            mocks.u2f = u2f_mock;
            U2FRegisterRequestGet.default(req as any, res as any);
        });

        it("should return forbidden if identity has not been verified", function (done) {
            res.send = sinon.spy(function (data: any) {
                assert.equal(403, res.status.getCall(0).args[0]);
                done();
            });
            authSession.identity_check = undefined;
            U2FRegisterRequestGet.default(req as any, res as any);
        });
    });
});

