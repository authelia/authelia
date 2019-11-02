var { setup } = require(`../test/suites/${process.argv[2]}/environment`);

(async function() {
  try  {
    await setup();
  } catch(err) {
    console.error(err);
    process.exit(1);
  }
})()
