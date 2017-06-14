module.exports = function (grunt) {
  const buildDir = "dist";

  grunt.initConfig({
    run: {
      options: {},
      "build": {
        cmd: "npm",
        args: ['run', 'build']
      },
      "tslint": {
        cmd: "npm",
        args: ['run', 'tslint']
      },
      "test": {
        cmd: "npm",
        args: ['run', 'test']
      },
      "docker-build": {
        cmd: "docker",
        args: ['build', '-t', 'clems4ever/authelia', '.']
      },
      "docker-restart": {
        cmd: "docker-compose",
        args: ['-f', 'docker-compose.yml', '-f', 'docker-compose.dev.yml', 'restart', 'auth']
      },
      "minify": {
        cmd: "./node_modules/.bin/uglifyjs",
        args: [`${buildDir}/src/server/public_html/js/authelia.js`, '-o', `${buildDir}/src/server/public_html/js/authelia.min.js`]
      },
      "apidoc": {
        cmd: "./node_modules/.bin/apidoc",
        args: ["-i", "src/server", "-o", "doc"]
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
        tasks: ['build'],
        options: {
          interrupt: true,
          atBegin: true
        }
      },
      server: {
        files: ['src/server/**/*.ts', 'test/server/**/*.ts'],
        tasks: ['build', 'run:docker-restart'],
        options: {
          interrupt: true,
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

  grunt.registerTask('default', ['build']);

  grunt.registerTask('build-resources', ['copy:resources', 'copy:views', 'copy:images', 'copy:thirdparties', 'concat:css', 'cssmin']);
  grunt.registerTask('build', ['run:tslint', 'run:build', 'browserify:dist']);
  grunt.registerTask('dist', ['build', 'build-resources', 'run:minify', 'cssmin']);

  grunt.registerTask('docker-build', ['run:docker-build']);
  grunt.registerTask('docker-restart', ['run:docker-restart']);

  grunt.registerTask('test', ['run:test']);
};
