var { teardown } = require(`../test/suites/${process.argv[2]}/environment`);

(async function() {
  try  {
    await teardown();
  } catch(err) {
    console.error(err);
    process.exit(1);
  }
})()
