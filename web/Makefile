SHELL := /bin/bash

clean:
	rm -rf bower_components/
	rm -rf static/bootstrap
	rm -rf static/jquery
	rm -rf static/js/angular.min.js
	rm -rf static/js/colorconverter.js

bowerinstall:
	# install node dependencies in package.json
	npm install
	bower install
	pushd bower_components/bootstrap/; npm install; make; make bootstrap; popd
	pushd bower_components/hue-color-converter/; tsc colorconverter.ts; popd

install: bowerinstall
	mkdir -p static
	cp -r bower_components/bootstrap/bootstrap static/
	cp -r bower_components/jquery static/
	cp -r js/ static/js/
	cp bower_components/angular/angular.min.js static/js/
	cp bower_components/angular-bootstrap/ui-bootstrap-tpls.min.js static/js/
	cp bower_components/hue-color-converter/colorconverter.js static/js/

developjs:
	ln -sf ../../js/marvin.js static/js/marvin.js
