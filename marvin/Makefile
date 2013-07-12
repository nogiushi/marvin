all:  angular angular-ui bootstrap jquery

clean:
	rm -r static/bootstrap
	rm -r static/jquery

bootstrap:
	mkdir -p static
	cp -r components/bootstrap/bootstrap static/

jquery:
	mkdir -p static
	cp -r components/jquery static/

angular:
	curl --create-dirs -o static/js/angular.min.js https://ajax.googleapis.com/ajax/libs/angularjs/1.0.7/angular.min.js

angular-ui:
	curl --create-dirs -o static/js/ui-bootstrap-tpls-0.4.0.min.js https://raw.github.com/angular-ui/bootstrap/gh-pages/ui-bootstrap-tpls-0.4.0.min.js


