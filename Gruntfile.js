module.exports = function(grunt) {
  grunt.initConfig({
    run: {
      options: {},
      "build-ts": {
        cmd: "npm",
        args: ['run', 'build-ts']
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
      }
    },
    copy: {
      resources: {
        expand: true,
        cwd: 'src/resources/',
        src: '**',
        dest: 'dist/src/resources/'
      },
      views: {
        expand: true,
        cwd: 'src/views/',
        src: '**',
        dest: 'dist/src/views/'
      },
      public_html: {
        expand: true,
        cwd: 'src/public_html/',
        src: '**',
        dest: 'dist/src/public_html/'
      }
    }
  });

  grunt.loadNpmTasks('grunt-run');
  grunt.loadNpmTasks('grunt-contrib-copy'); 

  grunt.registerTask('default', ['build']);
  
  grunt.registerTask('res', ['copy:resources', 'copy:views', 'copy:public_html']);

  grunt.registerTask('build', ['run:tslint', 'run:build-ts', 'res']);
  grunt.registerTask('docker-build', ['run:docker-build']);

  grunt.registerTask('test', ['run:test']);
};
