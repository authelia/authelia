var ListSuites = require('./utils/ListSuites');

const suites = ListSuites();

suites.forEach(async (suite: string) => {
  var { teardown } = require(`../test/suites/${suite}/environment`);;
  teardown().catch((err: Error) => console.error(err));
});