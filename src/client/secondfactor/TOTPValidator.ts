
import BluebirdPromise = require("bluebird");
import Endpoints = require("../../server/endpoints");

export function validate(token: string, $: JQueryStatic): BluebirdPromise<string> {
    return new BluebirdPromise<string>(function (resolve, reject) {
        $.ajax({
            url: Endpoints.SECOND_FACTOR_TOTP_POST,
            data: {
                token: token,
            },
            method: "POST",
            dataType: "json"
        } as JQueryAjaxSettings)
            .done(function (data: any) {
                resolve(data);
            })
            .fail(function (xhr: JQueryXHR, textStatus: string) {
                reject(new Error(textStatus));
            });
    });
}