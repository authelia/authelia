module.exports = function (grunt) {
  const buildDir = "dist";
  const schemaDir = "server/src/lib/configuration/Configuration.schema.json"

  grunt.initConfig({
    clean: {
        dist: ['dist'],
    },
    run: {
      "test-server-unit": {
        cmd: "./node_modules/.bin/mocha",
        args: ['--colors', '--require', 'ts-node/register', 'server/src/**/*.spec.ts']
      },
      "test-shared-unit": {
        cmd: "./node_modules/.bin/mocha",
        args: ['--colors', '--require', 'ts-node/register', 'shared/**/*.spec.ts']
      },
      "test-cucumber": {
        cmd: "./scripts/run-cucumber.sh",
        args: ["./test/features"]
      },
      "test-complete-config": {
        cmd: "./node_modules/.bin/mocha",
        args: ['--colors', '--require', 'ts-node/register', 'test/complete-config/**/*.ts']
      },
      "test-minimal-config": {
        cmd: "./node_modules/.bin/mocha",
        args: ['--colors', '--require', 'ts-node/register', 'test/minimal-config/**/*.ts']
      },
      "test-inactivity": {
        cmd: "./node_modules/.bin/mocha",
        args: ['--colors', '--require', 'ts-node/register', 'test/inactivity/**/*.ts']
      },
      "apidoc": {
        cmd: "./node_modules/.bin/apidoc",
        args: ["-i", "src/server", "-o", "doc"]
      },
    },
  });
  
  grunt.loadNpmTasks('grunt-contrib-copy');
  grunt.loadNpmTasks('grunt-contrib-clean');
  grunt.loadNpmTasks('grunt-run');

  grunt.registerTask('test-server', ['run:test-server-unit'])
  grunt.registerTask('test-shared', ['run:test-shared-unit'])
  grunt.registerTask('test-unit', ['test-server', 'test-shared']);
  grunt.registerTask('test-int', ['run:test-cucumber', 'run:test-minimal-config', 'run:test-complete-config', 'run:test-inactivity']);
};
