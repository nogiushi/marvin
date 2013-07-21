Marvin
======

Marvin is a character for your home that enhances your life. Marvin is software that runs on a [beaglebone](http://beagleboard.org/bone) along with a cape that contains a number of added sensors (available upon request).

This github repository contains Marvin's software. Marvin currently has the following abilities:

* to interact via a web interface
* to control [hue light bulbs](http://www.meethue.com/)
* to detect motion via its motion sensor
  * turn on nightlights upon sensing motion when in sleep mode
  * track history of motion
* to detect ambient light levels via its light sensor
  * turn on lights when it gets dark if in awake mode
* to track and key off of activity
  * sleeping causes Marvin to turn off ligths and turn on motion triggered nightlights
  * waking causes Marvin to slowly fade on lights (AKA sunrise)
* to schedule lighting or activity transitions for different times of the day and week
* to turn on and off various behaviours
  * nightlights switch for turning off motion triggered nightlights
  * schedule switch for turning off schedule
* to detect presence

## Components ##

* marvin - package containing core functionality
* marvin/web - package containing web interface
* marvin/marvin - package containing program


## Install [![Build Status](https://api.travis-ci.org/eikeon/marvin.png?branch=master)](https://travis-ci.org/eikeon/marvin) ##

    apt-get install gcc g++

    wget http://nodejs.org/dist/v0.10.13/node-v0.10.13.tar.gz
    tar xvfz node-v0.10.13.tar.gz
    cd node-v0.10.13
    ./configure --without-snapshot
    sudo make install

    go get -v -u github.com/eikeon/marvin/marvin
    export MARVIN_HOME=`go list -f '{{.Dir}}' github.com/eikeon/marvin/web`
    export PATH=$MARVIN_HOME/node_modules/.bin:$PATH
    pushd $MARVIN_HOME; make; popd
