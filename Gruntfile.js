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
                command: 'fpm -s dir -t deb -n marvin -v <%= bower.version %>-1 -C dest --deb-user root --deb-group root --deb-compression xz --description "Marvin is ..." --category "home" --url http://nogiushi.com/ -m "info<info@nogiushi.com>"  --architecture armhf -p marvin-<%= bower.version %>-1_armhf.deb -d "golang (>= 1.1.2)" --config-files etc/init/marvin.conf usr'
            }
        },
        clean: {
            static: ['static', 'build', 'dest']
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
                src: ['bower_components/jquery/dist/jquery.min.js', 'bower_components/angular/angular.min.js', 'bower_components/angular-animate/angular-animate.min.js', 'bower_components/angular-ui-bootstrap/index.js', 'build/js/colorconverter.js',  'js/marvin.js', 'bower_components/bootstrap/dist/js/bootstrap.min.js'],
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
        less: {
            compileCore: {
                options: {
                    strictMath: true,
                    sourceMap: true,
                    outputSourceFiles: true,
                    sourceMapURL: '<%= pkg.name %>.css.map',
                    sourceMapFilename: 'dist/css/<%= pkg.name %>.css.map'
                },
                files: {
                    'dest/usr/share/marvin/static/<%= bower.version %>/css/<%= bower.name %>.css': ['less/marvin.less']
                }
            },
            compileTheme: {
                options: {
                    strictMath: true,
                    sourceMap: true,
                    outputSourceFiles: true,
                    sourceMapURL: '<%= pkg.name %>-theme.css.map',
                    sourceMapFilename: 'dist/css/<%= pkg.name %>-theme.css.map'
                },
                files: {
                    'dist/css/<%= pkg.name %>-theme.css': 'less/theme.less'
                }
            },
            minify: {
                options: {
                    cleancss: true,
                    report: 'min'
                },
                files: {
                    'dest/usr/share/marvin/static/<%= bower.version %>/css/<%= bower.name %>.min.css': 'dest/usr/share/marvin/static/<%= bower.version %>/css/<%= bower.name %>.css',
                    'dist/css/<%= pkg.name %>-theme.min.css': 'dist/css/<%= pkg.name %>-theme.css'
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
            conf: {
                files: [
                    {src: ['conf/marvin.conf'], dest: 'dest/etc/init/marvin.conf'}
                ]
            },
            bin: {
                files: [
                    {src: ['marvin'], dest: 'dest/usr/bin/'}
                ]
            }
        },
        chmod: {
            options: {
                mode: '755'
            },
            marvin: {
                // Target-specific file/dir lists and/or options go here.
                src: ['dest/usr/bin/marvin']
            }
        }
    });

    require('load-grunt-tasks')(grunt, {scope: 'devDependencies'});

    // Test task.
    grunt.registerTask('test', ['jshint']);

    // JS distribution task.
    grunt.registerTask('static-js', ['typescript', 'ngmin', 'concat', 'uglify']); 

    // CSS distribution task.
    grunt.registerTask('static-css', ['less']);

    // Images distribution task
    grunt.registerTask('static-images', ['copy']);

    // Full distribution task.
    grunt.registerTask('static', ['clean', 'static-css', 'static-js', 'static-images']);

    // Default task.
    grunt.registerTask('default', ['shell:goinstall', 'test', 'static']);

    // Default task.
    grunt.registerTask('fpm', ['default', 'chmod:marvin', 'shell:fpm']);

};
