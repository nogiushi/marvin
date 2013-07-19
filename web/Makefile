all:  angular angular-ui bootstrap jquery

clean:
	rm -rf static/bootstrap
	rm -rf static/jquery

bower:
	bower install

bootstrap:  bower
	mkdir -p static
	pushd bower_components/bootstrap/; npm install; make; make bootstrap; popd
	cp -r bower_components/bootstrap/bootstrap static/

jquery:
	mkdir -p static
	cp -r bower_components/jquery static/

angular:
	curl --create-dirs -o static/js/angular.min.js https://ajax.googleapis.com/ajax/libs/angularjs/1.0.7/angular.min.js

angular-ui:
	curl --create-dirs -o static/js/ui-bootstrap-tpls-0.4.0.min.js https://raw.github.com/angular-ui/bootstrap/gh-pages/ui-bootstrap-tpls-0.4.0.min.js


