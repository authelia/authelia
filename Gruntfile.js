module.exports = function (grunt) {
  const buildDir = "dist";
  const schemaDir = "server/src/lib/configuration/Configuration.schema.json"

  grunt.initConfig({
    env: {
      "env-test-server-unit": {
        TS_NODE_PROJECT: "server/tsconfig.json"
      },
      "env-test-client-unit": {
        TS_NODE_PROJECT: "client/tsconfig.json"
      },
      "env-test-shared-unit": {
        TS_NODE_PROJECT: "server/tsconfig.json"
      }
    },
    run: {
      "compile-server": {
        cmd: "./node_modules/.bin/tsc",
        args: ['-p', 'server/tsconfig.json']
      },
      "generate-config-schema": {
        cmd: "./node_modules/.bin/typescript-json-schema",
        args: ["-o", schemaDir, "--strictNullChecks",
               "--required", "server/tsconfig.json", "Configuration"]
      },
      "compile-client": {
        cmd: "./node_modules/.bin/tsc",
        args: ['-p', 'client/tsconfig.json']
      },
      "lint-server": {
        cmd: "./node_modules/.bin/tslint",
        args: ['-c', 'server/tslint.json', '-p', 'server/tsconfig.json']
      },
      "lint-client": {
        cmd: "./node_modules/.bin/tslint",
        args: ['-c', 'client/tslint.json', '-p', 'client/tsconfig.json']
      },
      "test-server-unit": {
        cmd: "./node_modules/.bin/mocha",
        args: ['--colors', '--require', 'ts-node/register', 'server/src/**/*.spec.ts']
      },
      "test-shared-unit": {
        cmd: "./node_modules/.bin/mocha",
        args: ['--colors', '--require', 'ts-node/register', 'shared/**/*.spec.ts']
      },
      "test-client-unit": {
        cmd: "./node_modules/.bin/mocha",
        args: ['--colors', '--require', 'ts-node/register', 'client/test/**/*.test.ts']
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
      "minify": {
        cmd: "./node_modules/.bin/uglifyjs",
        args: [`${buildDir}/server/src/public_html/js/authelia.js`, '-o', `${buildDir}/server/src/public_html/js/authelia.min.js`]
      },
      "apidoc": {
        cmd: "./node_modules/.bin/apidoc",
        args: ["-i", "src/server", "-o", "doc"]
      },
      "include-minified-script": {
        cmd: "sed",
        args: ["-i", "s/authelia.\(js\|css\)/authelia.min.\1/", `${buildDir}/server/src/views/layout/layout.pug`]
      }
    },
    copy: {
      resources: {
        expand: true,
        cwd: 'server/src/resources/',
        src: '**',
        dest: `${buildDir}/server/src/resources/`
      },
      views: {
        expand: true,
        cwd: 'server/src/views/',
        src: '**',
        dest: `${buildDir}/server/src/views/`
      },
      images: {
        expand: true,
        cwd: 'client/src/img',
        src: '**',
        dest: `${buildDir}/server/src/public_html/img/`
      },
      thirdparties: {
        expand: true,
        cwd: 'client/src/thirdparties',
        src: '**',
        dest: `${buildDir}/server/src/public_html/js/`
      },
      schema: {
        src: schemaDir,
        dest: `${buildDir}/${schemaDir}`
      }
    },
    browserify: {
      dist: {
        src: ['dist/client/src/index.js'],
        dest: `${buildDir}/server/src/public_html/js/authelia.js`,
        options: {
          browserifyOptions: {
            standalone: 'authelia'
          },
        },
      },
    },
    watch: {
      views: {
        files: ['server/src/views/**/*.pug'],
        tasks: ['copy:views'],
        options: {
          interrupt: false,
          atBegin: true
        }
      },
      resources: {
        files: ['server/src/resources/*.ejs'],
        tasks: ['copy:resources'],
        options: {
          interrupt: false,
          atBegin: true
        }
      },
      images: {
        files: ['client/src/img/**'],
        tasks: ['copy:images'],
        options: {
          interrupt: false,
          atBegin: true
        }
      },
      css: {
        files: ['client/src/**/*.css'],
        tasks: ['concat:css', 'cssmin'],
        options: {
          interrupt: true,
          atBegin: true
        }
      },
      client: {
        files: ['client/src/**/*.ts'],
        tasks: ['build-dev'],
        options: {
          interrupt: true,
          atBegin: true
        }
      },
      server: {
        files: ['server/src/**/*.ts'],
        tasks: ['build-dev', 'run:docker-restart', 'run:make-dev-views' ],
        options: {
          interrupt: true,
          atBegin: true
        }
      }
    },
    concat: {
      css: {
        src: ['client/src/css/*.css'],
        dest: `${buildDir}/server/src/public_html/css/authelia.css`
      },
    },
    cssmin: {
      target: {
        files: {
          [`${buildDir}/server/src/public_html/css/authelia.min.css`]: [`${buildDir}/server/src/public_html/css/authelia.css`]
        }
      }
    }
  });

  grunt.loadNpmTasks('grunt-browserify');
  grunt.loadNpmTasks('grunt-contrib-concat');
  grunt.loadNpmTasks('grunt-contrib-copy');
  grunt.loadNpmTasks('grunt-contrib-cssmin');
  grunt.loadNpmTasks('grunt-contrib-watch');
  grunt.loadNpmTasks('grunt-run');
  grunt.loadNpmTasks('grunt-env');


  grunt.registerTask('compile-server', ['run:lint-server', 'run:compile-server'])
  grunt.registerTask('compile-client', ['run:lint-client', 'run:compile-client'])

  grunt.registerTask('test-server', ['env:env-test-server-unit', 'run:test-server-unit'])
  grunt.registerTask('test-shared', ['env:env-test-shared-unit', 'run:test-shared-unit'])
  grunt.registerTask('test-client', ['env:env-test-client-unit', 'run:test-client-unit'])
  grunt.registerTask('test-unit', ['test-server', 'test-client', 'test-shared']);
  grunt.registerTask('test-int', ['run:test-cucumber', 'run:test-minimal-config', 'run:test-complete-config', 'run:test-inactivity']);

  grunt.registerTask('copy-resources', ['copy:resources', 'copy:views', 'copy:images', 'copy:thirdparties', 'concat:css']);
  grunt.registerTask('generate-config-schema', ['run:generate-config-schema', 'copy:schema']);

  grunt.registerTask('build-client', ['compile-client', 'browserify']);
  grunt.registerTask('build-server', ['compile-server', 'copy-resources', 'generate-config-schema']);

  grunt.registerTask('build', ['build-client', 'build-server']);
  grunt.registerTask('build-dist', ['build', 'run:minify', 'cssmin', 'run:include-minified-script']);

  grunt.registerTask('schema', ['run:generate-config-schema'])

  grunt.registerTask('docker-build', ['run:docker-build']);

  grunt.registerTask('default', ['build-dist']);
};
