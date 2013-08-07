module.exports = function(grunt) {

    // Project configuration.
    grunt.initConfig({
        pkg: grunt.file.readJSON('package.json'),
        banner: '/**\n' +
                '* <%= pkg.name %>.js v<%= pkg.version %> \n' +
                '* <%= grunt.template.today("yyyy/mm/dd") %> \n' +
                '*/\n',
        clean: {
            static: ['static', 'build']
        },
        ngmin: {
            marvin: {
                src: ['js/<%= pkg.name %>.js'],
                dest: 'build/js/<%= pkg.name %>.annotate.js'
            },
        },
        concat: {
            options: {
                banner: '<%= banner %>',
                stripBanners: false
            },
            marvin: {
                src: ['bower_components/jquery/jquery.min.js', 'bower_components/angularjs/index.js', 'bower_components/angular-ui-bootstrap/index.js', 'build/js/colorconverter.js',  'js/marvin.js', 'bower_components/bootstrap/dist/js/bootstrap.min.js'],
                dest: 'static/<%= pkg.version %>/js/<%= pkg.name %>.js'
            }
        },
        uglify: {
            options: {
                banner: '<%= banner %>'
            },
            marvin: {
                files: {
                    'static/<%= pkg.version %>/js/<%= pkg.name %>.min.js': ['<%= concat.marvin.dest %>']
                }
            }
        },
        jshint: {
            options: {
                jshintrc: 'js/.jshintrc'
            },
            gruntfile: {
                src: 'Gruntfile.js'
            },
            src: {
                src: ['js/*.js']
            },
            test: {
                src: ['js/tests/unit/*.js']
            }
        },
        recess: {
            options: {
                compile: true
            },
            marvin: {
                files: {
                    'static/<%= pkg.version %>/css/<%= pkg.name %>.css': ['bower_components/bootstrap/less/bootstrap.less']
                }
            },
            min: {
                options: {
                    compress: true
                },
                files: {
                    'static/<%= pkg.version %>/css/<%= pkg.name %>.min.css': ['bower_components/bootstrap/less/bootstrap.less']
                }
            }
        },
        typescript: {
            base: {
                src: ['bower_components/hue-color-converter/colorconverter.ts'],
                dest: 'build/js/colorconverter.js'
            }
        },
        copy: {
            main: {
                files: [
                    {
                        src: 'images/*',
                        dest: 'static/<%= pkg.version %>/'
                    }
                ]
            }
        }
    });

    // These plugins provide necessary tasks.
    grunt.loadNpmTasks('grunt-contrib-uglify');
    grunt.loadNpmTasks('grunt-contrib-jshint');
    grunt.loadNpmTasks('grunt-contrib-clean');
    grunt.loadNpmTasks('grunt-contrib-concat');
    grunt.loadNpmTasks('grunt-ngmin');
    grunt.loadNpmTasks('grunt-recess');
    grunt.loadNpmTasks('grunt-typescript');
    grunt.loadNpmTasks('grunt-contrib-copy');

    // Test task.
    grunt.registerTask('test', ['jshint']);

    // JS distribution task.
    grunt.registerTask('static-js', ['typescript', 'ngmin', 'concat', 'uglify']); 

    // Default task(s).
    grunt.registerTask('default', ['uglify']);

    // CSS distribution task.
    grunt.registerTask('static-css', ['recess']);

    // Images distribution task
    grunt.registerTask('static-images', ['copy']);

    // Full distribution task.
    grunt.registerTask('static', ['clean', 'static-css', 'static-js', 'static-images']);

    // Default task.
    grunt.registerTask('default', ['test', 'static']);

};
