SHELL := /bin/bash

all:  bowerinstall

clean:
	rm -rf bower_components/
	rm -rf static/bootstrap
	rm -rf static/jquery
	rm -rf static/js/angular.min.js
	rm -rf static/js/colorconverter.js

bowerinstall:
	bower install
	pushd bower_components/bootstrap/; npm install; make; make bootstrap; popd
	pushd bower_components/hue-color-converter/; npm install typescript; ./node_modules/typescript/bin/tsc colorconverter.ts

install: bowerinstall
	mkdir -p static
	cp -r bower_components/bootstrap/bootstrap static/
	cp -r bower_components/jquery static/
	mkdir -p static/js
	cp -r bower_components/angular/angular.min.js static/js/
	cp -r bower_components/angular-bootstrap/ui-bootstrap-tpls.min.js static/js/
	cp -r bower_components/hue-color-converter/colorconverter.js static/js/
