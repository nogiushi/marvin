Marvin
======

Marvin is a home automation program. Currently Marvin is executing our light scedule by controlling our [hue light bulbs](http://www.meethue.com/) and is currently running on a [beaglebone](http://beagleboard.org/bone).


sudo npm install bower -g
sudo npm install recess connect uglify-js jshint -g

bower install
cd components/bootstrap/
make
make bootstrap
cd ../..
make
