module.exports = function(grunt) {

    // Project configuration.
    grunt.initConfig({
        pkg: grunt.file.readJSON('package.json'),
        bower: grunt.file.readJSON('bower.json'),
        banner: '/**\n' +
                '* <%= bower.name %>.js v<%= bower.version %> \n' +
                '* <%= grunt.template.today("yyyy/mm/dd") %> \n' +
                '*/\n',
	shell: {
            goinstall: {
                options: {
                    failOnError: true,
                    stdout: true,
                    execOptions: {
			cwd: '.'
                    }
		},
		command: 'go build -v .'
            },
            fpm: {
                options: {
                    failOnError: true,
                    stdout: true,
                    execOptions: {
                        cwd: '.'
                    }
                },
                command: 'fpm -s dir -t deb -n marvin -v <%= bower.version %>-1 -C dest --deb-user root --deb-group root --deb-compression xz --description "Marvin is ..." --category "home" --url http://nogiushi.com/ -m "info<info@nogiushi.com>"  --architecture armhf -p marvin-<%= bower.version %>-1_armhf.deb -d "golang (>= 1.1.2)" etc usr'
            }
        },
        clean: {
            static: ['static', 'build']
        },
        ngmin: {
            marvin: {
                src: ['js/<%= bower.name %>.js'],
                dest: 'build/js/<%= bower.name %>.annotate.js'
            },
        },
        concat: {
            options: {
                banner: '<%= banner %>',
                stripBanners: false
            },
            marvin: {
                src: ['bower_components/jquery/jquery.min.js', 'bower_components/angularjs/index.js', 'bower_components/angular-ui-bootstrap/index.js', 'build/js/colorconverter.js',  'js/marvin.js', 'bower_components/bootstrap/dist/js/bootstrap.min.js'],
                dest: 'dest/usr/share/marvin/static/<%= bower.version %>/js/<%= bower.name %>.js'
            }
        },
        uglify: {
            options: {
                banner: '<%= banner %>'
            },
            marvin: {
                files: {
                    'dest/usr/share/marvin/static/<%= bower.version %>/js/<%= bower.name %>.min.js': ['<%= concat.marvin.dest %>']
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
                    'dest/usr/share/marvin/static/<%= bower.version %>/css/<%= bower.name %>.css': ['less/marvin.less']
                }
            },
            min: {
                options: {
                    compress: true
                },
                files: {
                    'dest/usr/share/marvin/static/<%= bower.version %>/css/<%= bower.name %>.min.css': ['less/marvin.less']
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
            images: {
                files: [
                    {src: 'bower.json', dest: 'dest/usr/share/marvin/'
                    },
                    {
                        src: 'images/*',
                        dest: 'dest/usr/share/marvin/static/<%= bower.version %>/'
                    },
                    {
			expand: true,
			cwd: 'bower_components/bootstrap/dist/',
                        src: ['fonts/*'],
                        dest: 'dest/usr/share/marvin/static/<%= bower.version %>/'
                    }
                ]
            },
            templates: {
                files: [
                    {src: ['*/*.html'], dest: 'dest/usr/share/marvin/'}
                ]
            },
            json: {
                files: [
                    {src: ['conf/marvin.json'], dest: 'dest/etc/marvin.json'}
                ]
            },
            bin: {
                files: [
                    {src: ['marvin'], dest: 'dest/usr/bin/'}
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
    grunt.loadNpmTasks('grunt-shell');
    grunt.loadNpmTasks('grunt-typescript');
    grunt.loadNpmTasks('grunt-contrib-copy');

    // Test task.
    grunt.registerTask('test', ['jshint']);

    // JS distribution task.
    grunt.registerTask('static-js', ['typescript', 'ngmin', 'concat', 'uglify']); 

    // CSS distribution task.
    grunt.registerTask('static-css', ['recess']);

    // Images distribution task
    grunt.registerTask('static-images', ['copy']);

    // Full distribution task.
    grunt.registerTask('static', ['clean', 'static-css', 'static-js', 'static-images']);

    // Default task.
    grunt.registerTask('default', ['shell:goinstall', 'test', 'static', 'shell:fpm']);

    // Default task.
    grunt.registerTask('fpm', ['default', 'shell:fpm']);

};
