
import sinon = require("sinon");
import express = require("express");

export interface RequestMock {
    app?: any;
    body?: any;
    session?: any;
    headers?: any;
    get?: any;
    query?: any;
}

export interface ResponseMock {
    send: sinon.SinonStub | sinon.SinonSpy;
    sendStatus: sinon.SinonStub;
    sendFile: sinon.SinonStub;
    sendfile: sinon.SinonStub;
    status: sinon.SinonStub | sinon.SinonSpy;
    json: sinon.SinonStub | sinon.SinonSpy;
    links: sinon.SinonStub;
    jsonp: sinon.SinonStub;
    download: sinon.SinonStub;
    contentType: sinon.SinonStub;
    type: sinon.SinonStub;
    format: sinon.SinonStub;
    attachment: sinon.SinonStub;
    set: sinon.SinonStub;
    header: sinon.SinonStub;
    headersSent: boolean;
    get: sinon.SinonStub;
    clearCookie: sinon.SinonStub;
    cookie: sinon.SinonStub;
    location: sinon.SinonStub;
    redirect: sinon.SinonStub | sinon.SinonSpy;
    render: sinon.SinonStub | sinon.SinonSpy;
    locals: sinon.SinonStub;
    charset: string;
    vary: sinon.SinonStub;
    app: any;
    write: sinon.SinonStub;
    writeContinue: sinon.SinonStub;
    writeHead: sinon.SinonStub;
    statusCode: number;
    statusMessage: string;
    setHeader: sinon.SinonStub;
    setTimeout: sinon.SinonStub;
    sendDate: boolean;
    getHeader: sinon.SinonStub;
}

export function RequestMock(): RequestMock {
    return {
        app: {
            get: sinon.stub()
        }
    };
}
export function ResponseMock(): ResponseMock {
    return {
        send: sinon.stub(),
        status: sinon.stub(),
        json: sinon.stub(),
        sendStatus: sinon.stub(),
        links: sinon.stub(),
        jsonp: sinon.stub(),
        sendFile: sinon.stub(),
        sendfile: sinon.stub(),
        download: sinon.stub(),
        contentType: sinon.stub(),
        type: sinon.stub(),
        format: sinon.stub(),
        attachment: sinon.stub(),
        set: sinon.stub(),
        header: sinon.stub(),
        headersSent: true,
        get: sinon.stub(),
        clearCookie: sinon.stub(),
        cookie: sinon.stub(),
        location: sinon.stub(),
        redirect: sinon.stub(),
        render: sinon.stub(),
        locals: sinon.stub(),
        charset: "utf-8",
        vary: sinon.stub(),
        app: sinon.stub(),
        write: sinon.stub(),
        writeContinue: sinon.stub(),
        writeHead: sinon.stub(),
        statusCode: 200,
        statusMessage: "message",
        setHeader: sinon.stub(),
        setTimeout: sinon.stub(),
        sendDate: true,
        getHeader: sinon.stub()
    };
}
