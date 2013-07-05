Marvin
======

Marvin is a home automation program. Currently Marvin is executing our light scedule by controlling our [hue light bulbs](http://www.meethue.com/) and is currently running on a [beaglebone](http://beagleboard.org/bone).


    sudo npm install bower -g
    sudo npm install recess connect uglify-js jshint -g

    bower install
    cd components/bootstrap/
    npm install
    make
    make bootstrap
    cd ../..
    make

	cd static/js/
	wget https://ajax.googleapis.com/ajax/libs/angularjs/1.0.7/angular.min.js
	wget https://raw.github.com/angular-ui/bootstrap/gh-pages/ui-bootstrap-tpls-0.4.0.min.js
	cd ../..
