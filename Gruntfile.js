module.exports = function (grunt) {
  const buildDir = "dist";
  const schemaDir = "server/src/lib/configuration/Configuration.schema.json"

  grunt.initConfig({
    clean: {
        dist: ['dist'],
    },
    run: {
      "compile-server": {
        cmd: "./node_modules/.bin/tsc",
        args: ['-p', 'server/tsconfig.json']
      },
      "compile-client": {
        exec: 'cd ./client && npm run build && mv build ../dist/client',
      },
      "generate-config-schema": {
        cmd: "./node_modules/.bin/typescript-json-schema",
        args: ["-o", schemaDir, "--strictNullChecks",
               "--required", "server/tsconfig.json", "Configuration"]
      },
      "lint-server": {
        cmd: "./node_modules/.bin/tslint",
        args: ['-c', 'server/tslint.json', '-p', 'server/tsconfig.json']
      },
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
      "docker-build": {
        cmd: "docker",
        args: ['build', '-t', 'clems4ever/authelia', '.']
      },
      "apidoc": {
        cmd: "./node_modules/.bin/apidoc",
        args: ["-i", "src/server", "-o", "doc"]
      },
    },
    copy: {
        backup: {
          files: [{
            expand: true,
            src: ['dist/**'],
            dest: 'backup'
            }]
        },
        resources: {
            expand: true,
            cwd: 'server/src/resources',
            src: '**',
            dest: `${buildDir}/server/src/resources/`
        },
        schema: {
            src: schemaDir,
            dest: `${buildDir}/${schemaDir}`
        }
    },
  });
  
  grunt.loadNpmTasks('grunt-contrib-copy');
  grunt.loadNpmTasks('grunt-contrib-clean');
  grunt.loadNpmTasks('grunt-run');

  grunt.registerTask('compile-server', ['run:lint-server', 'run:compile-server'])
  grunt.registerTask('compile-client', ['run:lint-client', 'run:compile-client'])

  grunt.registerTask('test-server', ['env:env-test-server-unit', 'run:test-server-unit'])
  grunt.registerTask('test-shared', ['env:env-test-shared-unit', 'run:test-shared-unit'])
  grunt.registerTask('test-unit', ['test-server', 'test-client', 'test-shared']);
  grunt.registerTask('test-int', ['run:test-cucumber', 'run:test-minimal-config', 'run:test-complete-config', 'run:test-inactivity']);

  grunt.registerTask('generate-config-schema', ['run:generate-config-schema', 'copy:schema']);
  
  grunt.registerTask('build-client', ['compile-client', 'browserify']);

  grunt.registerTask('build-server', ['compile-server', 'copy:resources', 'generate-config-schema']);
  
  grunt.registerTask('build', ['build-client', 'build-server']);
  grunt.registerTask('build-dist', ['clean:dist', 'build']);
  grunt.registerTask('schema', ['run:generate-config-schema'])

  grunt.registerTask('docker-build', ['run:docker-build']);
};
