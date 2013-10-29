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

* marvin - program containing the server daemon
* marvin/nog - package containing the message center
* marvin/web - package containing web interface

## Install [![Build Status](https://api.travis-ci.org/nogiushi/marvin.png?branch=master)](https://travis-ci.org/nogiushi/marvin) ##

### get ubuntu 

If you are installing Marvin on your BeagleBone we are using 13.10 via [rcn-ee.net](https://rcn-ee.net/deb/flasher/saucy/BBB-eMMC-flasher-ubuntu-13.10-2013-10-25.img.xz).

### install build tools

    sudo apt-get install gcc g++ make mercurial

### install latest golang

Marvin is written in Go so you will need a Go environment to build and install
it. You will probably want to put the GOPATH and GOROOT environment variables
in your ~/.profile.

    hg clone -u release https://code.google.com/p/go
    sudo mv go /opt/go
    cd /opt/go
    ./all.bash
    export GOPATH=$HOME
    export GOROOT=/opt/go

### install latest nodejs

Marvin's needs a nodejs environment for managing external javascript and css
dependencies using [Bower](https://github.com/bower/bower).

    wget http://nodejs.org/dist/v0.10.21/node-v0.10.21.tar.gz
    tar xvfz node-v0.10.21.tar.gz
    cd node-v0.10.21
    ./configure --without-snapshot
    sudo make install

### install grunt and bower

	sudo npm install -g grunt-cli
	sudo npm install -g bower

### install marvin

    go get -v -u github.com/nogiushi/marvin
    pushd `go list -f '{{.Dir}}' github.com/nogiushi/marvin/web`; npm install; bower install; grunt
    sudo cp ../conf/marvin.json /etc/marvin.json
    sudo cp ../conf/marvin.conf /etc/init/marvin.conf
    sudo start marvin
    # point browser at http://{your-hostname}/

### Other Marvin channels

[Marvin Magazine](http://flip.it/MBhif)

![Marvin](https://raw.github.com/nogiushi/marvin/master/web/images/robot.png)
