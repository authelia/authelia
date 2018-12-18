import BluebirdPromise = require("bluebird");

export default function ($: JQueryStatic, url: string, data: Object, fn: any,
  dataType: string): BluebirdPromise<any> {
  return new BluebirdPromise<any>((resolve, reject) => {
    $.get(url, {}, undefined, dataType)
      .done((data: any) => {
        resolve(data);
      })
      .fail((xhr: JQueryXHR, textStatus: string) => {
        reject(textStatus);
      });
  });
}