module.exports = function (grunt) {
  const buildDir = "dist";

  grunt.initConfig({
    run: {
      options: {},
      "build": {
        cmd: "./node_modules/.bin/tsc",
        args: ['-p', 'tsconfig.json']
      },
      "tslint": {
        cmd: "./node_modules/.bin/tslint",
        args: ['-c', 'tslint.json', '-p', 'tsconfig.json']
      },
      "unit-tests": {
        cmd: "./node_modules/.bin/mocha",
        args: ['--compilers', 'ts:ts-node/register', '--recursive', 'test/unit']
      },
      "integration-tests": {
        cmd: "./node_modules/.bin/cucumber-js",
        args: ["--compiler", "ts:ts-node/register", "./test/features"]
      },
      "docker-build": {
        cmd: "docker",
        args: ['build', '-t', 'clems4ever/authelia', '.']
      },
      "docker-restart": {
        cmd: "./scripts/dc-dev.sh",
        args: ['up', '-d']
      },
      "minify": {
        cmd: "./node_modules/.bin/uglifyjs",
        args: [`${buildDir}/src/server/public_html/js/authelia.js`, '-o', `${buildDir}/src/server/public_html/js/authelia.min.js`]
      },
      "apidoc": {
        cmd: "./node_modules/.bin/apidoc",
        args: ["-i", "src/server", "-o", "doc"]
      },
      "make-dev-views": {
        cmd: "sed",
        args: ["-i", "s/authelia\.min/authelia/", `${buildDir}/src/server/views/layout/layout.pug`]
      }
    },
    copy: {
      resources: {
        expand: true,
        cwd: 'src/server/resources/',
        src: '**',
        dest: `${buildDir}/src/server/resources/`
      },
      views: {
        expand: true,
        cwd: 'src/server/views/',
        src: '**',
        dest: `${buildDir}/src/server/views/`
      },
      images: {
        expand: true,
        cwd: 'src/client/img',
        src: '**',
        dest: `${buildDir}/src/server/public_html/img/`
      },
      thirdparties: {
        expand: true,
        cwd: 'src/client/thirdparties',
        src: '**',
        dest: `${buildDir}/src/server/public_html/js/`
      },
    },
    browserify: {
      dist: {
        src: ['dist/src/client/index.js'],
        dest: `${buildDir}/src/server/public_html/js/authelia.js`,
        options: {
          browserifyOptions: {
            standalone: 'authelia'
          },
        },
      },
    },
    watch: {
      views: {
        files: ['src/server/views/**/*.pug'],
        tasks: ['copy:views'],
        options: {
          interrupt: false,
          atBegin: true
        }
      },
      resources: {
        files: ['src/server/resources/*.ejs'],
        tasks: ['copy:resources'],
        options: {
          interrupt: false,
          atBegin: true
        }
      },
      images: {
        files: ['src/client/img/**'],
        tasks: ['copy:images'],
        options: {
          interrupt: false,
          atBegin: true
        }
      },
      css: {
        files: ['src/client/**/*.css'],
        tasks: ['concat:css', 'cssmin'],
        options: {
          interrupt: true,
          atBegin: true
        }
      },
      client: {
        files: ['src/client/**/*.ts', 'test/client/**/*.ts'],
        tasks: ['build-dev'],
        options: {
          interrupt: true,
          atBegin: true
        }
      },
      server: {
        files: ['src/server/**/*.ts', 'test/server/**/*.ts'],
        tasks: ['build-dev', 'run:docker-restart', 'run:make-dev-views' ],
        options: {
          interrupt: true,
          atBegin: true
        }
      }
    },
    concat: {
      css: {
        src: ['src/client/css/*.css'],
        dest: `${buildDir}/src/server/public_html/css/authelia.css`
      },
    },
    cssmin: {
      target: {
        files: {
          [`${buildDir}/src/server/public_html/css/authelia.min.css`]: [`${buildDir}/src/server/public_html/css/authelia.css`]
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

  grunt.registerTask('default', ['build-dist']);

  grunt.registerTask('build-resources', ['copy:resources', 'copy:views', 'copy:images', 'copy:thirdparties', 'concat:css']);

  grunt.registerTask('build-common', ['run:tslint', 'run:build', 'browserify:dist', 'build-resources']);
  grunt.registerTask('build-dev', ['build-common', 'run:make-dev-views']);
  grunt.registerTask('build-dist', ['build-common', 'run:minify', 'cssmin']);

  grunt.registerTask('docker-build', ['run:docker-build']);
  grunt.registerTask('docker-restart', ['run:docker-restart']);

  grunt.registerTask('unit-tests', ['run:unit-tests']);
  grunt.registerTask('integration-tests', ['run:unit-tests']);

  grunt.registerTask('test', ['unit-tests']);
};
