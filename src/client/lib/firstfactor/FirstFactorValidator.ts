
import BluebirdPromise = require("bluebird");
import Endpoints = require("../../../server/endpoints");

export function validate(username: string, password: string, $: JQueryStatic): BluebirdPromise<void> {
    return new BluebirdPromise<void>(function (resolve, reject) {
        $.post(Endpoints.FIRST_FACTOR_POST, {
            username: username,
            password: password,
        })
            .done(function () {
                resolve();
            })
            .fail(function (xhr: JQueryXHR, textStatus: string) {
                reject(new Error("Authetication failed. Please check your credentials."));
            });
    });
}
