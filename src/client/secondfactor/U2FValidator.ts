
import U2fApi = require("u2f-api");
import U2f = require("u2f");
import BluebirdPromise = require("bluebird");
import { SignMessage } from "../../server/lib/routes/secondfactor/u2f/sign_request/SignMessage";
import Endpoints = require("../../server/endpoints");

function finishU2fAuthentication(responseData: U2fApi.SignResponse, $: JQueryStatic): BluebirdPromise<void> {
    return new BluebirdPromise<void>(function (resolve, reject) {
        $.ajax({
            url: Endpoints.SECOND_FACTOR_U2F_SIGN_POST,
            data: responseData,
            method: "POST",
            dataType: "json"
        } as JQueryAjaxSettings)
            .done(function (data) {
                resolve(data);
            })
            .fail(function (xhr: JQueryXHR, textStatus: string) {
                reject(new Error(textStatus));
            });
    });
}

function startU2fAuthentication($: JQueryStatic, u2fApi: typeof U2fApi): BluebirdPromise<void> {
    return new BluebirdPromise<void>(function (resolve, reject) {
        $.get(Endpoints.SECOND_FACTOR_U2F_SIGN_REQUEST_GET, {}, undefined, "json")
            .done(function (signResponse: SignMessage) {
                $.notify("Please touch the token", "info");

                const signRequest: U2fApi.SignRequest = {
                    appId: signResponse.request.appId,
                    challenge: signResponse.request.challenge,
                    keyHandle: signResponse.keyHandle, // linked to the client session cookie
                    version: "U2F_V2"
                };

                u2fApi.sign([signRequest], 60)
                    .then(function (signResponse: U2fApi.SignResponse) {
                        finishU2fAuthentication(signResponse, $)
                            .then(function (data) {
                                resolve(data);
                            }, function (err) {
                                $.notify("Error when finish U2F transaction", "error");
                                reject(err);
                            });
                    })
                    .catch(function (err: Error) {
                        reject(err);
                    });
            })
            .fail(function (xhr: JQueryXHR, textStatus: string) {
                reject(new Error(textStatus));
            });
    });
}


export function validate($: JQueryStatic, u2fApi: typeof U2fApi): BluebirdPromise<void> {
    return startU2fAuthentication($, u2fApi);
}
