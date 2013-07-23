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

### get ubuntu 

If you are installing Marvin on your BeagleBone we've been using 13.04 from [ARMhf](http://www.armhf.com/index.php/boards/beaglebone-black/).

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
    mkdir $HOME/go
    export GOPATH=$HOME/go
    export GOROOT=/opt/go

### install latest nodejs

Marvin's needs a nodejs environment for managing external javascript and css
dependencies using [Bower](https://github.com/bower/bower).

    wget http://nodejs.org/dist/v0.10.13/node-v0.10.13.tar.gz
    tar xvfz node-v0.10.13.tar.gz
    cd node-v0.10.13
    ./configure --without-snapshot
    sudo make install

### install marvin

    go get -v -u github.com/eikeon/marvin/marvin
    export MARVIN_HOME=$HOME/go/src/github.com/eikeon/marvin
    export PATH=$MARVIN_HOME/web/node_modules/.bin:$PATH
    pushd $MARVIN_HOME/web; make install; popd
    sudo cp conf/marvin.json /etc/marvin.json
    sudo cp conf/marvin.conf /etc/init/marvin.conf
    sudo start marvin
    # point browser at http://{your-hostname}:8000
