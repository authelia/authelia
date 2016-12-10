
module.exports = {
  'promisify': promisify,
  'resolve': resolve,
  'reject': reject
}

var Q = require('q');

function promisify(fn, context) {
  return function() {
    var defer = Q.defer();
    var args = Array.prototype.slice.call(arguments);
    args.push(function(err, val) {
      if (err !== null && err !== undefined) {
        return defer.reject(err);
      }
      return defer.resolve(val);
    });
    fn.apply(context || {}, args);
    return defer.promise;
  };
}

function resolve(data) {
  var defer = Q.defer();
  defer.resolve(data);
  return defer.promise;
}

function reject(err) {
  var defer = Q.defer();
  defer.reject(err);
  return defer.promise;
}
