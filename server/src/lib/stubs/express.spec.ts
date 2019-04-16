
import * as Sinon from "sinon";
import * as Express from "express";
import { GET_VARIABLE_KEY } from "../constants";
import { RequestLoggerStub } from "../logging/RequestLoggerStub.spec";

export interface ResponseMock {
    send: Sinon.SinonStub | Sinon.SinonSpy;
    sendStatus: Sinon.SinonStub;
    sendFile: Sinon.SinonStub;
    sendfile: Sinon.SinonStub;
    status: Sinon.SinonStub | Sinon.SinonSpy;
    json: Sinon.SinonStub | Sinon.SinonSpy;
    links: Sinon.SinonStub;
    jsonp: Sinon.SinonStub;
    download: Sinon.SinonStub;
    contentType: Sinon.SinonStub;
    type: Sinon.SinonStub;
    format: Sinon.SinonStub;
    attachment: Sinon.SinonStub;
    set: Sinon.SinonStub;
    header: Sinon.SinonStub;
    headersSent: boolean;
    get: Sinon.SinonStub;
    clearCookie: Sinon.SinonStub;
    cookie: Sinon.SinonStub;
    location: Sinon.SinonStub;
    redirect: Sinon.SinonStub | Sinon.SinonSpy;
    render: Sinon.SinonStub | Sinon.SinonSpy;
    locals: Sinon.SinonStub;
    charset: string;
    vary: Sinon.SinonStub;
    app: any;
    write: Sinon.SinonStub;
    writeContinue: Sinon.SinonStub;
    writeHead: Sinon.SinonStub;
    statusCode: number;
    statusMessage: string;
    setHeader: Sinon.SinonStub;
    setTimeout: Sinon.SinonStub;
    sendDate: boolean;
    getHeader: Sinon.SinonStub;
}

export function RequestMock(): Express.Request {
    const getMock = Sinon.mock()
      .withArgs(GET_VARIABLE_KEY).atLeast(0).returns({
          logger: new RequestLoggerStub()
      });
    return {
        id: '1234',
        headers: {},
        app: {
            get: getMock,
            set: Sinon.mock(),
        },
        body: {},
        query: {},
        session: {
            id: '1234',
            regenerate: function() {},
            reload: function() {},
            destroy: function() {},
            save: function() {},
            touch: function() {},
            cookie: {
                domain: 'example.com',
                expires: true,
                httpOnly: true,
                maxAge: 36000,
                originalMaxAge: 36000,
                path: '/',
                secure: true,
                serialize: () => '',
            }
        }
    } as any;
}
export function ResponseMock(): ResponseMock {
    return {
        send: Sinon.stub(),
        status: Sinon.stub(),
        json: Sinon.stub(),
        sendStatus: Sinon.stub(),
        links: Sinon.stub(),
        jsonp: Sinon.stub(),
        sendFile: Sinon.stub(),
        sendfile: Sinon.stub(),
        download: Sinon.stub(),
        contentType: Sinon.stub(),
        type: Sinon.stub(),
        format: Sinon.stub(),
        attachment: Sinon.stub(),
        set: Sinon.stub(),
        header: Sinon.stub(),
        headersSent: true,
        get: Sinon.stub(),
        clearCookie: Sinon.stub(),
        cookie: Sinon.stub(),
        location: Sinon.stub(),
        redirect: Sinon.stub(),
        render: Sinon.stub(),
        locals: Sinon.stub(),
        charset: "utf-8",
        vary: Sinon.stub(),
        app: Sinon.stub(),
        write: Sinon.stub(),
        writeContinue: Sinon.stub(),
        writeHead: Sinon.stub(),
        statusCode: 200,
        statusMessage: "message",
        setHeader: Sinon.stub(),
        setTimeout: Sinon.stub(),
        sendDate: true,
        getHeader: Sinon.stub()
    };
}
