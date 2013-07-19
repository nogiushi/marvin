SHELL := /bin/bash

all:  angular angular-ui bowerinstall

clean:
	rm -rf static/bootstrap
	rm -rf static/jquery

bowerinstall:
	npm install bower
	./node_modules/bower/bin/bower install; 
	pushd bower_components/bootstrap/; npm install; make; make bootstrap; popd
	mkdir -p static
	cp -r bower_components/bootstrap/bootstrap static/
	cp -r bower_components/jquery static/

angular:
	curl --create-dirs -o static/js/angular.min.js https://ajax.googleapis.com/ajax/libs/angularjs/1.0.7/angular.min.js

angular-ui:
	curl --create-dirs -o static/js/ui-bootstrap-tpls-0.4.0.min.js https://raw.github.com/angular-ui/bootstrap/gh-pages/ui-bootstrap-tpls-0.4.0.min.js


