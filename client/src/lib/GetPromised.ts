import BluebirdPromise = require("bluebird");

export default function ($: JQueryStatic, url: string, data: Object, fn: any,
  dataType: string): BluebirdPromise<any> {
  return new BluebirdPromise<any>(function (resolve, reject) {
    $.get(url, {}, undefined, dataType)
      .done(function (data: any) {
        resolve(data);
      })
      .fail(function (xhr: JQueryXHR, textStatus: string) {
        reject(textStatus);
      });
  });
}